package connection

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"sync"
	"time"

	"remote-file-manager/internal/config"
	"remote-file-manager/internal/database"
	"remote-file-manager/internal/log"

	"github.com/jlaffaye/ftp"
	"golang.org/x/crypto/ssh"
)

// Category 连接分类
type Category struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ConnectionType 连接类型枚举
type ConnectionType string

// 连接类型常量
const (
	ConnectionTypeSSH  ConnectionType = "ssh"
	ConnectionTypeSFTP ConnectionType = "sftp"
	ConnectionTypeFTP  ConnectionType = "ftp"
)

// Connection 通用连接配置
type Connection struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Type       ConnectionType `json:"type"` // 连接类型: ssh, ftp, sftp
	Host       string         `json:"host"`
	Port       int            `json:"port"`
	Username   string         `json:"username"`
	Password   string         `json:"password,omitempty"`    // 加密存储 - 适用于所有类型
	PrivateKey string         `json:"private_key,omitempty"` // 加密存储 - 仅SSH/SFTP使用
	Passphrase string         `json:"passphrase,omitempty"`  // 加密存储 - 仅SSH/SFTP使用
	IsActive   bool           `json:"is_active"`
	Category   string         `json:"category"` // 连接分类
	LastUsed   time.Time      `json:"last_used"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// ConnectionManager 通用连接管理器
type ConnectionManager struct {
	cfg           *config.Config
	connections   map[string]*Connection
	sshClientPool map[string]*ssh.Client
	ftpClientPool map[string]*ftp.ServerConn
	categories    map[string]*Category
	poolMutex     sync.Mutex
	encryptionKey []byte
}

// NewConnectionManager 创建连接管理器
func NewConnectionManager(cfg *config.Config) *ConnectionManager {
	manager := &ConnectionManager{
		cfg:           cfg,
		connections:   make(map[string]*Connection),
		sshClientPool: make(map[string]*ssh.Client),
		ftpClientPool: make(map[string]*ftp.ServerConn),
		categories:    make(map[string]*Category),
		encryptionKey: []byte(cfg.PasswordSalt[:32]), // 使用密码盐作为加密密钥
	}

	// 从数据库加载连接和分类
	manager.loadConnections()
	manager.loadCategories()

	return manager
}

// NewSSHManager 创建SSH连接管理器（兼容旧代码）
func NewSSHManager(cfg *config.Config) *ConnectionManager {
	return NewConnectionManager(cfg)
}

// AddConnection 添加连接
func (m *ConnectionManager) AddConnection(c *ssh.Client, conn *Connection) error {
	// 设置默认连接类型
	if conn.Type == "" {
		conn.Type = ConnectionTypeSSH
	}

	// 设置默认端口
	if conn.Port == 0 {
		switch conn.Type {
		case ConnectionTypeSSH, ConnectionTypeSFTP:
			conn.Port = 22
		case ConnectionTypeFTP:
			conn.Port = 21
		}
	}

	// 生成唯一ID
	if conn.ID == "" {
		conn.ID = fmt.Sprintf("%d-%s", time.Now().UnixNano(), conn.Name)
	}

	// 加密敏感信息
	conn.Password = m.encrypt(conn.Password)
	conn.PrivateKey = m.encrypt(conn.PrivateKey)
	conn.Passphrase = m.encrypt(conn.Passphrase)

	// 设置创建和更新时间
	now := time.Now()
	conn.CreatedAt = now
	conn.UpdatedAt = now

	// 插入到数据库
	_, err := database.DB.Exec(`
		INSERT INTO ssh_connections (
			id, name, type, host, port, username, password, private_key, passphrase, 
			is_active, category, last_used, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		conn.ID, conn.Name, conn.Type, conn.Host, conn.Port, conn.Username, conn.Password,
		conn.PrivateKey, conn.Passphrase, conn.IsActive, conn.Category,
		conn.LastUsed, conn.CreatedAt, conn.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert connection: %v", err)
	}

	// 添加到内存
	m.connections[conn.ID] = conn
	if c != nil {
		m.poolMutex.Lock()
		m.sshClientPool[conn.ID] = c
		m.poolMutex.Unlock()
	}

	log.Info("Added %s connection: %s", conn.Type, conn.Name)
	return nil
}

// GetConnections 获取所有连接
func (m *ConnectionManager) GetConnections() []Connection {
	conns := make([]Connection, 0, len(m.connections))
	for _, conn := range m.connections {
		// 返回时不包含敏感信息
		connCopy := *conn
		connCopy.Password = ""
		connCopy.PrivateKey = ""
		connCopy.Passphrase = ""
		conns = append(conns, connCopy)
	}
	return conns
}

// GetConnection 获取单个连接
func (m *ConnectionManager) GetConnection(id string) (*Connection, bool) {
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

// UpdateConnection 更新连接
func (m *ConnectionManager) UpdateConnection(id string, conn *Connection) error {
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

	// 确保类型不为空
	if conn.Type == "" {
		conn.Type = original.Type
	}

	// 更新数据库
	_, err := database.DB.Exec(`
		UPDATE ssh_connections SET
			name = ?, type = ?, host = ?, port = ?, username = ?, password = ?, private_key = ?, 
			passphrase = ?, is_active = ?, category = ?, last_used = ?, 
			created_at = ?, updated_at = ?
		WHERE id = ?
	`,
		conn.Name, conn.Type, conn.Host, conn.Port, conn.Username, conn.Password, conn.PrivateKey,
		conn.Passphrase, conn.IsActive, conn.Category, conn.LastUsed,
		conn.CreatedAt, conn.UpdatedAt, conn.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update connection: %v", err)
	}

	// 更新内存
	m.connections[id] = conn

	// 如果连接池中有该连接，关闭并移除
	m.poolMutex.Lock()
	if client, exists := m.sshClientPool[id]; exists {
		client.Close()
		delete(m.sshClientPool, id)
	}
	if client, exists := m.ftpClientPool[id]; exists {
		client.Quit()
		delete(m.ftpClientPool, id)
	}
	m.poolMutex.Unlock()

	log.Info("Updated %s connection: %s", conn.Type, conn.Name)
	return nil
}

// DeleteConnection 删除连接
func (m *ConnectionManager) DeleteConnection(id string) error {
	// 检查连接是否存在
	_, exists := m.connections[id]
	if !exists {
		return fmt.Errorf("connection not found: %s", id)
	}

	// 从数据库中删除
	_, err := database.DB.Exec(`DELETE FROM ssh_connections WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete connection from database: %v", err)
	}

	// 从内存中移除
	delete(m.connections, id)

	// 从连接池中移除并关闭连接
	m.poolMutex.Lock()
	if client, exists := m.sshClientPool[id]; exists {
		client.Close()
		delete(m.sshClientPool, id)
	}
	if client, exists := m.ftpClientPool[id]; exists {
		client.Quit()
		delete(m.ftpClientPool, id)
	}
	m.poolMutex.Unlock()

	log.Info("Deleted connection: %s", id)
	return nil
}

// GetSSHClient 获取SSH客户端连接（适用于SSH和SFTP类型）
func (m *ConnectionManager) GetSSHClient(id string) (*ssh.Client, error) {
	// 检查连接池
	m.poolMutex.Lock()
	client, exists := m.sshClientPool[id]
	m.poolMutex.Unlock()

	if exists {
		// 检查连接是否有效
		_, _, err := client.SendRequest("keepalive", true, nil)
		if err == nil {
			return client, nil
		}
		// 连接无效，移除
		m.poolMutex.Lock()
		delete(m.sshClientPool, id)
		m.poolMutex.Unlock()
	}

	// 获取连接配置
	conn, exists := m.connections[id]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", id)
	}

	// 检查连接类型是否支持SSH
	if conn.Type != ConnectionTypeSSH && conn.Type != ConnectionTypeSFTP {
		return nil, fmt.Errorf("connection type %s does not support SSH client", conn.Type)
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
	m.sshClientPool[id] = client
	m.poolMutex.Unlock()

	// 更新连接的最后使用时间
	conn.LastUsed = time.Now()

	// 更新到数据库
	_, err = database.DB.Exec(`
		UPDATE ssh_connections 
		SET last_used = ?
		WHERE id = ?
	`, conn.LastUsed, conn.ID)
	if err != nil {
		log.Error("Failed to update connection last_used: %v", err)
	}

	log.Info("Established %s connection to %s", conn.Type, addr)
	return client, nil
}

// GetFTPClient 获取FTP客户端连接
func (m *ConnectionManager) GetFTPClient(id string) (*ftp.ServerConn, error) {
	// 检查连接池
	m.poolMutex.Lock()
	client, exists := m.ftpClientPool[id]
	m.poolMutex.Unlock()

	if exists {
		// 检查连接是否有效
		if _, err := client.CurrentDir(); err == nil {
			return client, nil
		}
		// 连接无效，移除
		m.poolMutex.Lock()
		delete(m.ftpClientPool, id)
		m.poolMutex.Unlock()
	}

	// 获取连接配置
	conn, exists := m.connections[id]
	if !exists {
		return nil, fmt.Errorf("connection not found: %s", id)
	}

	// 检查连接类型是否支持FTP
	if conn.Type != ConnectionTypeFTP {
		return nil, fmt.Errorf("connection type %s does not support FTP client", conn.Type)
	}

	// 解密敏感信息
	password := m.decrypt(conn.Password)

	// 建立FTP连接
	addr := fmt.Sprintf("%s:%d", conn.Host, conn.Port)
	client, err := ftp.Dial(addr, ftp.DialWithTimeout(30*time.Second))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %v", addr, err)
	}

	// 登录FTP服务器
	if err := client.Login(conn.Username, password); err != nil {
		client.Quit()
		return nil, fmt.Errorf("failed to login to FTP server: %v", err)
	}

	// 添加到连接池
	m.poolMutex.Lock()
	m.ftpClientPool[id] = client
	m.poolMutex.Unlock()

	// 更新连接的最后使用时间
	conn.LastUsed = time.Now()

	// 更新到数据库
	_, err = database.DB.Exec(`
		UPDATE ssh_connections 
		SET last_used = ?
		WHERE id = ?
	`, conn.LastUsed, conn.ID)
	if err != nil {
		log.Error("Failed to update connection last_used: %v", err)
	}

	log.Info("Established %s connection to %s", conn.Type, addr)
	return client, nil
}

// GetClient 获取SSH客户端连接（兼容旧代码）
func (m *ConnectionManager) GetClient(id string) (*ssh.Client, error) {
	return m.GetSSHClient(id)
}

// TestConnection 测试连接
func (m *ConnectionManager) TestConnection(id string) error {
	// 获取连接配置
	conn, exists := m.connections[id]
	if !exists {
		return fmt.Errorf("connection not found: %s", id)
	}

	switch conn.Type {
	case ConnectionTypeSSH, ConnectionTypeSFTP:
		// 测试SSH/SFTP连接
		client, err := m.GetSSHClient(id)
		if err != nil {
			return err
		}

		// 关闭连接
		client.Close()

		// 从连接池中移除
		m.poolMutex.Lock()
		delete(m.sshClientPool, id)
		m.poolMutex.Unlock()

	case ConnectionTypeFTP:
		// 测试FTP连接
		client, err := m.GetFTPClient(id)
		if err != nil {
			return err
		}

		// 关闭连接
		client.Quit()

		// 从连接池中移除
		m.poolMutex.Lock()
		delete(m.ftpClientPool, id)
		m.poolMutex.Unlock()

	default:
		return fmt.Errorf("unknown connection type: %s", conn.Type)
	}

	return nil
}

// CloseAllConnections 关闭所有连接
func (m *ConnectionManager) CloseAllConnections() {
	m.poolMutex.Lock()
	defer m.poolMutex.Unlock()

	// 关闭所有SSH客户端连接
	for id, client := range m.sshClientPool {
		client.Close()
		delete(m.sshClientPool, id)
		log.Info("Closed SSH connection: %s", id)
	}

	// 关闭所有FTP客户端连接
	for id, client := range m.ftpClientPool {
		client.Quit()
		delete(m.ftpClientPool, id)
		log.Info("Closed FTP connection: %s", id)
	}
}

// loadConnections 从数据库加载所有连接
func (m *ConnectionManager) loadConnections() {
	// 从数据库查询所有连接
	rows, err := database.DB.Query(`
		SELECT id, name, type, host, port, username, password, private_key, passphrase, is_active, category, last_used, created_at, updated_at
		FROM ssh_connections
	`)
	if err != nil {
		log.Error("Failed to query connections: %v", err)
		return
	}
	defer rows.Close()

	// 清空现有连接
	m.connections = make(map[string]*Connection)

	// 遍历结果集
	for rows.Next() {
		var conn Connection
		var lastUsed sql.NullTime
		var connType string // 临时变量用于接收数据库中的type字段

		// 扫描行数据到连接结构体
		err := rows.Scan(
			&conn.ID,
			&conn.Name,
			&connType, // 读取连接类型
			&conn.Host,
			&conn.Port,
			&conn.Username,
			&conn.Password,
			&conn.PrivateKey,
			&conn.Passphrase,
			&conn.IsActive,
			&conn.Category,
			&lastUsed,
			&conn.CreatedAt,
			&conn.UpdatedAt,
		)
		if err != nil {
			log.Error("Failed to scan connection: %v", err)
			continue
		}

		// 设置连接类型，默认为SSH
		conn.Type = ConnectionType(connType)
		if conn.Type == "" {
			conn.Type = ConnectionTypeSSH
		}

		// 处理可空的last_used字段
		if lastUsed.Valid {
			conn.LastUsed = lastUsed.Time
		}

		// 添加到内存
		m.connections[conn.ID] = &conn
		log.Info("Loaded %s connection: %s", conn.Type, conn.Name)
	}

	// 检查遍历过程中是否有错误
	if err := rows.Err(); err != nil {
		log.Error("Error during connection iteration: %v", err)
	}
}

// loadCategories 从数据库加载所有分类
func (m *ConnectionManager) loadCategories() {
	// 从数据库查询所有分类
	rows, err := database.DB.Query(`
		SELECT name, created_at, updated_at
		FROM ssh_categories
	`)
	if err != nil {
		log.Error("Failed to query categories: %v", err)
		return
	}
	defer rows.Close()

	// 清空现有分类
	m.categories = make(map[string]*Category)

	// 遍历结果集
	for rows.Next() {
		var category Category

		// 扫描行数据到分类结构体
		err := rows.Scan(
			&category.Name,
			&category.CreatedAt,
			&category.UpdatedAt,
		)
		if err != nil {
			log.Error("Failed to scan category: %v", err)
			continue
		}

		// 添加到内存
		m.categories[category.Name] = &category
		log.Info("Loaded category: %s", category.Name)
	}

	// 检查遍历过程中是否有错误
	if err := rows.Err(); err != nil {
		log.Error("Error during category iteration: %v", err)
	}
}

// saveCategories 保存所有分类配置到数据库（已废弃，改用单条保存）
func (m *ConnectionManager) saveCategories() {
	// 此方法已废弃，分类现在通过单独的数据库操作保存
	// 保留此方法以保持兼容性
	log.Warn("saveCategories method is deprecated, categories are now saved individually")
}

// GetCategories 获取所有分类
func (m *ConnectionManager) GetCategories() []*Category {
	categories := make([]*Category, 0, len(m.categories))
	for _, category := range m.categories {
		categories = append(categories, category)
	}
	return categories
}

// Connections 获取连接映射的引用（用于API测试）
func (m *ConnectionManager) Connections() map[string]*Connection {
	return m.connections
}

// AddCategory 添加分类
func (m *ConnectionManager) AddCategory(name string) error {
	// 检查分类是否已存在
	if _, exists := m.categories[name]; exists {
		return fmt.Errorf("category already exists: %s", name)
	}

	// 创建新分类
	now := time.Now()
	category := &Category{
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// 插入到数据库
	_, err := database.DB.Exec(`
		INSERT INTO ssh_categories (name, created_at, updated_at)
		VALUES (?, ?, ?)
	`, name, now, now)
	if err != nil {
		return fmt.Errorf("failed to insert category: %v", err)
	}

	// 添加到内存
	m.categories[name] = category

	log.Info("Added category: %s", name)
	return nil
}

// UpdateCategory 更新分类
func (m *ConnectionManager) UpdateCategory(oldName, newName string) error {
	// 检查旧分类是否存在
	category, exists := m.categories[oldName]
	if !exists {
		return fmt.Errorf("category not found: %s", oldName)
	}

	// 检查新分类名是否已存在
	if _, exists := m.categories[newName]; exists && newName != oldName {
		return fmt.Errorf("category already exists: %s", newName)
	}

	// 更新分类名
	category.Name = newName
	category.UpdatedAt = time.Now()

	// 开始事务，确保原子性
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()

	// 更新分类表
	_, err = tx.Exec(`
		UPDATE ssh_categories 
		SET name = ?, updated_at = ?
		WHERE name = ?
	`, newName, category.UpdatedAt, oldName)
	if err != nil {
		return fmt.Errorf("failed to update category: %v", err)
	}

	// 更新所有使用该分类的连接
	_, err = tx.Exec(`
		UPDATE ssh_connections 
		SET category = ?, updated_at = ?
		WHERE category = ?
	`, newName, time.Now(), oldName)
	if err != nil {
		return fmt.Errorf("failed to update connections with old category: %v", err)
	}

	// 从内存中删除旧分类名
	delete(m.categories, oldName)
	// 添加新分类名
	m.categories[newName] = category

	// 重新加载连接，确保内存中的连接信息与数据库一致
	m.loadConnections()

	log.Info("Updated category: %s -> %s", oldName, newName)
	return nil
}

// DeleteCategory 删除分类
func (m *ConnectionManager) DeleteCategory(name string) error {
	// 检查分类是否存在
	if _, exists := m.categories[name]; !exists {
		return fmt.Errorf("category not found: %s", name)
	}

	// 开始事务，确保原子性
	tx, err := database.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()

	// 删除分类
	_, err = tx.Exec(`DELETE FROM ssh_categories WHERE name = ?`, name)
	if err != nil {
		return fmt.Errorf("failed to delete category: %v", err)
	}

	// 更新所有使用该分类的连接为未分类
	_, err = tx.Exec(`
		UPDATE ssh_connections 
		SET category = '', updated_at = ?
		WHERE category = ?
	`, time.Now(), name)
	if err != nil {
		return fmt.Errorf("failed to update connections with deleted category: %v", err)
	}

	// 从内存中删除分类
	delete(m.categories, name)

	// 重新加载连接，确保内存中的连接信息与数据库一致
	m.loadConnections()

	log.Info("Deleted category: %s", name)
	return nil
}

// Encrypt 加密字符串（导出方法，供API使用）
func (m *ConnectionManager) Encrypt(text string) string {
	return m.encrypt(text)
}

// encrypt 加密字符串
func (m *ConnectionManager) encrypt(text string) string {
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

// RemoveConnection 从内存中移除连接（用于移除临时测试连接）
func (m *ConnectionManager) RemoveConnection(id string) {
	// 从连接池中移除并关闭连接
	m.poolMutex.Lock()
	if client, exists := m.sshClientPool[id]; exists {
		client.Close()
		delete(m.sshClientPool, id)
	}
	m.poolMutex.Unlock()

	// 从内存中移除连接
	delete(m.connections, id)

	log.Info("Removed connection: %s", id)
}

// decrypt 解密字符串
func (m *ConnectionManager) decrypt(text string) string {
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
