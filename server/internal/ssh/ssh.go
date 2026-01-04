package ssh

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"file-manager/internal/config"
	"file-manager/internal/log"

	"golang.org/x/crypto/ssh"
)

// SSHConnection SSH连接配置
type SSHConnection struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Host       string    `json:"host"`
	Port       int       `json:"port"`
	Username   string    `json:"username"`
	Password   string    `json:"password,omitempty"`    // 加密存储
	PrivateKey string    `json:"private_key,omitempty"` // 加密存储
	Passphrase string    `json:"passphrase,omitempty"`  // 加密存储
	IsActive   bool      `json:"is_active"`
	LastUsed   time.Time `json:"last_used"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SSHManager SSH连接管理器
type SSHManager struct {
	cfg            *config.Config
	connections    map[string]*SSHConnection
	clientPool     map[string]*ssh.Client
	poolMutex      sync.Mutex
	connectionsDir string
	encryptionKey  []byte
}

// NewSSHManager 创建SSH连接管理器
func NewSSHManager(cfg *config.Config) *SSHManager {
	manager := &SSHManager{
		cfg:            cfg,
		connections:    make(map[string]*SSHConnection),
		clientPool:     make(map[string]*ssh.Client),
		connectionsDir: "./connections",
		encryptionKey:  []byte(cfg.PasswordSalt[:32]), // 使用密码盐作为加密密钥
	}

	// 创建连接配置目录
	if err := os.MkdirAll(manager.connectionsDir, 0755); err != nil {
		log.Error("Failed to create connections directory: %v", err)
	}

	// 加载现有连接配置
	manager.loadConnections()

	return manager
}

// AddConnection 添加SSH连接
func (m *SSHManager) AddConnection(c *ssh.Client, conn *SSHConnection) error {
	// 加密敏感信息
	conn.Password = m.encrypt(conn.Password)
	conn.PrivateKey = m.encrypt(conn.PrivateKey)
	conn.Passphrase = m.encrypt(conn.Passphrase)

	// 设置创建和更新时间
	now := time.Now()
	conn.CreatedAt = now
	conn.UpdatedAt = now

	// 保存到文件
	filename := filepath.Join(m.connectionsDir, conn.ID+".json")
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create connection file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(conn); err != nil {
		return fmt.Errorf("failed to encode connection: %v", err)
	}

	// 添加到内存
	m.connections[conn.ID] = conn
	if c != nil {
		m.poolMutex.Lock()
		m.clientPool[conn.ID] = c
		m.poolMutex.Unlock()
	}

	log.Info("Added SSH connection: %s", conn.Name)
	return nil
}

// GetConnections 获取所有SSH连接
func (m *SSHManager) GetConnections() []*SSHConnection {
	conns := make([]*SSHConnection, 0, len(m.connections))
	for _, conn := range m.connections {
		// 返回时不包含敏感信息
		connCopy := *conn
		connCopy.Password = ""
		connCopy.PrivateKey = ""
		connCopy.Passphrase = ""
		conns = append(conns, &connCopy)
	}
	return conns
}

// GetConnection 获取单个SSH连接
func (m *SSHManager) GetConnection(id string) (*SSHConnection, bool) {
	conn, exists := m.connections[id]
	if !exists {
		return nil, false
	}

	// 返回时不包含敏感信息
	connCopy := *conn
	connCopy.Password = ""
	connCopy.PrivateKey = ""
	connCopy.Passphrase = ""
	return &connCopy, true
}

// UpdateConnection 更新SSH连接
func (m *SSHManager) UpdateConnection(id string, conn *SSHConnection) error {
	// 检查连接是否存在
	original, exists := m.connections[id]
	if !exists {
		return fmt.Errorf("connection not found: %s", id)
	}

	// 保留原有加密信息，只更新非敏感字段
	if conn.Password != "" {
		conn.Password = m.encrypt(conn.Password)
	} else {
		conn.Password = original.Password
	}

	if conn.PrivateKey != "" {
		conn.PrivateKey = m.encrypt(conn.PrivateKey)
	} else {
		conn.PrivateKey = original.PrivateKey
	}

	if conn.Passphrase != "" {
		conn.Passphrase = m.encrypt(conn.Passphrase)
	} else {
		conn.Passphrase = original.Passphrase
	}

	// 设置更新时间
	conn.UpdatedAt = time.Now()
	conn.CreatedAt = original.CreatedAt
	conn.ID = original.ID

	// 保存到文件
	filename := filepath.Join(m.connectionsDir, conn.ID+".json")
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create connection file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(conn); err != nil {
		return fmt.Errorf("failed to encode connection: %v", err)
	}

	// 更新内存
	m.connections[id] = conn

	// 如果连接池中有该连接，关闭并移除
	m.poolMutex.Lock()
	if client, exists := m.clientPool[id]; exists {
		client.Close()
		delete(m.clientPool, id)
	}
	m.poolMutex.Unlock()

	log.Info("Updated SSH connection: %s", conn.Name)
	return nil
}

// DeleteConnection 删除SSH连接
func (m *SSHManager) DeleteConnection(id string) error {
	// 检查连接是否存在
	_, exists := m.connections[id]
	if !exists {
		return fmt.Errorf("connection not found: %s", id)
	}

	// 删除文件
	filename := filepath.Join(m.connectionsDir, id+".json")
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete connection file: %v", err)
	}

	// 从内存中移除
	delete(m.connections, id)

	// 从连接池中移除并关闭连接
	m.poolMutex.Lock()
	if client, exists := m.clientPool[id]; exists {
		client.Close()
		delete(m.clientPool, id)
	}
	m.poolMutex.Unlock()

	log.Info("Deleted SSH connection: %s", id)
	return nil
}

// GetClient 获取SSH客户端连接
func (m *SSHManager) GetClient(id string) (*ssh.Client, error) {
	// 检查连接池
	m.poolMutex.Lock()
	client, exists := m.clientPool[id]
	m.poolMutex.Unlock()

	if exists {
		// 检查连接是否有效
		_, _, err := client.SendRequest("keepalive", true, nil)
		if err == nil {
			return client, nil
		}
		// 连接无效，移除
		m.poolMutex.Lock()
		delete(m.clientPool, id)
		m.poolMutex.Unlock()
	}

	// 获取连接配置
	conn, exists := m.connections[id]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", id)
	}

	// 解密敏感信息
	password := m.decrypt(conn.Password)
	privateKey := m.decrypt(conn.PrivateKey)
	passphrase := m.decrypt(conn.Passphrase)

	// 构建SSH配置
	config := &ssh.ClientConfig{
		User:            conn.Username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 生产环境应使用更安全的验证
		Timeout:         30 * time.Second,
	}

	// 添加认证方式
	if password != "" {
		config.Auth = append(config.Auth, ssh.Password(password))
	}

	if privateKey != "" {
		// 解析私钥
		var signer ssh.Signer
		var err error

		if passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(privateKey), []byte(passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(privateKey))
		}

		if err == nil {
			config.Auth = append(config.Auth, ssh.PublicKeys(signer))
		} else {
			log.Warn("Failed to parse private key: %v", err)
		}
	}

	// 建立连接
	addr := fmt.Sprintf("%s:%d", conn.Host, conn.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %v", addr, err)
	}

	// 添加到连接池
	m.poolMutex.Lock()
	m.clientPool[id] = client
	m.poolMutex.Unlock()

	// 更新连接的最后使用时间
	conn.LastUsed = time.Now()
	m.saveConnection(conn)

	log.Info("Established SSH connection to %s", addr)
	return client, nil
}

// TestConnection 测试SSH连接
func (m *SSHManager) TestConnection(id string) error {
	client, err := m.GetClient(id)
	if err != nil {
		return err
	}

	// 关闭连接
	client.Close()

	// 从连接池中移除
	m.poolMutex.Lock()
	delete(m.clientPool, id)
	m.poolMutex.Unlock()

	return nil
}

// CloseAllConnections 关闭所有SSH连接
func (m *SSHManager) CloseAllConnections() {
	m.poolMutex.Lock()
	defer m.poolMutex.Unlock()

	for id, client := range m.clientPool {
		client.Close()
		delete(m.clientPool, id)
		log.Info("Closed SSH connection: %s", id)
	}
}

// loadConnections 加载所有SSH连接配置
func (m *SSHManager) loadConnections() {
	// 读取目录中的所有连接配置文件
	files, err := os.ReadDir(m.connectionsDir)
	if err != nil {
		log.Error("Failed to read connections directory: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := filepath.Join(m.connectionsDir, file.Name())
		content, err := os.ReadFile(filename)
		if err != nil {
			log.Error("Failed to read connection file %s: %v", filename, err)
			continue
		}

		var conn SSHConnection
		if err := json.Unmarshal(content, &conn); err != nil {
			log.Error("Failed to unmarshal connection file %s: %v", filename, err)
			continue
		}

		// 添加到内存
		m.connections[conn.ID] = &conn
		log.Info("Loaded SSH connection: %s", conn.Name)
	}
}

// saveConnection 保存SSH连接配置
func (m *SSHManager) saveConnection(conn *SSHConnection) {
	filename := filepath.Join(m.connectionsDir, conn.ID+".json")
	file, err := os.Create(filename)
	if err != nil {
		log.Error("Failed to create connection file: %v", err)
		return
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(conn); err != nil {
		log.Error("Failed to encode connection: %v", err)
	}
}

// encrypt 加密字符串
func (m *SSHManager) encrypt(text string) string {
	if text == "" {
		return ""
	}

	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		log.Error("Failed to create cipher: %v", err)
		return text
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Error("Failed to create GCM: %v", err)
		return text
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		log.Error("Failed to generate nonce: %v", err)
		return text
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(text), nil)
	return base64.StdEncoding.EncodeToString(ciphertext)
}

// decrypt 解密字符串
func (m *SSHManager) decrypt(text string) string {
	if text == "" {
		return ""
	}

	ciphertext, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		log.Error("Failed to base64 decode: %v", err)
		return ""
	}

	block, err := aes.NewCipher(m.encryptionKey)
	if err != nil {
		log.Error("Failed to create cipher: %v", err)
		return ""
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		log.Error("Failed to create GCM: %v", err)
		return ""
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		log.Error("Ciphertext too short")
		return ""
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		log.Error("Failed to decrypt: %v", err)
		return ""
	}

	return string(plaintext)
}
