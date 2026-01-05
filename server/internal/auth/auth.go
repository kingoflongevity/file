package auth

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"remote-file-manager/internal/config"
	"remote-file-manager/internal/connection"
	"remote-file-manager/internal/log"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

// User 用户信息
type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Password  string    `json:"password,omitempty"` // 加密存储
	Email     string    `json:"email"`
	Role      string    `json:"role"` // admin, user
	Active    bool      `json:"active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UserConnectionPermission 用户连接权限
type UserConnectionPermission struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	ConnectionID string    `json:"connection_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Claims JWT声明
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthManager 认证管理器
type AuthManager struct {
	cfg            *config.Config
	users          map[string]*User
	usersDir       string
	encryptionKey  []byte
	permissions    map[string]*UserConnectionPermission
	permissionsDir string
}

// permissionFilePath 获取权限文件路径
func (m *AuthManager) permissionFilePath() string {
	return filepath.Join(m.permissionsDir, "permissions.json")
}

// loadPermissions 加载所有权限
func (m *AuthManager) loadPermissions() {
	// 创建权限目录
	if err := os.MkdirAll(m.permissionsDir, 0755); err != nil {
		log.Error("Failed to create permissions directory: %v", err)
		return
	}

	// 读取权限文件
	content, err := ioutil.ReadFile(m.permissionFilePath())
	if err != nil {
		// 文件不存在，初始化空权限
		if os.IsNotExist(err) {
			m.permissions = make(map[string]*UserConnectionPermission)
			return
		}
		log.Error("Failed to read permissions file: %v", err)
		return
	}

	// 解析权限数据
	var permissions []*UserConnectionPermission
	if err := json.Unmarshal(content, &permissions); err != nil {
		log.Error("Failed to unmarshal permissions: %v", err)
		return
	}

	// 添加到内存
	for _, perm := range permissions {
		m.permissions[perm.ID] = perm
	}
}

// savePermissions 保存所有权限
func (m *AuthManager) savePermissions() error {
	// 转换为切片
	permissions := make([]*UserConnectionPermission, 0, len(m.permissions))
	for _, perm := range m.permissions {
		permissions = append(permissions, perm)
	}

	// 保存到文件
	file, err := os.Create(m.permissionFilePath())
	if err != nil {
		return fmt.Errorf("failed to create permissions file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(permissions); err != nil {
		return fmt.Errorf("failed to encode permissions: %v", err)
	}

	return nil
}

// GetUserPermissions 获取用户的所有权限
func (m *AuthManager) GetUserPermissions(userID string) []*UserConnectionPermission {
	permissions := make([]*UserConnectionPermission, 0)
	for _, perm := range m.permissions {
		if perm.UserID == userID {
			permissions = append(permissions, perm)
		}
	}
	return permissions
}

// AddUserPermission 添加用户连接权限
func (m *AuthManager) AddUserPermission(permission *UserConnectionPermission) error {
	// 生成ID
	permission.ID = m.GenerateID()
	permission.CreatedAt = time.Now()
	permission.UpdatedAt = time.Now()

	// 添加到内存
	m.permissions[permission.ID] = permission

	// 保存到文件
	return m.savePermissions()
}

// DeleteUserPermission 删除用户连接权限
func (m *AuthManager) DeleteUserPermission(permissionID string) error {
	// 检查权限是否存在
	if _, exists := m.permissions[permissionID]; !exists {
		return fmt.Errorf("permission not found: %s", permissionID)
	}

	// 从内存中删除
	delete(m.permissions, permissionID)

	// 保存到文件
	return m.savePermissions()
}

// DeleteUserPermissionsByConnection 删除指定连接的所有权限
func (m *AuthManager) DeleteUserPermissionsByConnection(connectionID string) error {
	// 查找并删除所有相关权限
	for id, perm := range m.permissions {
		if perm.ConnectionID == connectionID {
			delete(m.permissions, id)
		}
	}

	// 保存到文件
	return m.savePermissions()
}

// DeleteUserPermissionsByUser 删除指定用户的所有权限
func (m *AuthManager) DeleteUserPermissionsByUser(userID string) error {
	// 查找并删除所有相关权限
	for id, perm := range m.permissions {
		if perm.UserID == userID {
			delete(m.permissions, id)
		}
	}

	// 保存到文件
	return m.savePermissions()
}

// NewAuthManager 创建认证管理器
func NewAuthManager(cfg *config.Config) *AuthManager {
	manager := &AuthManager{
		cfg:            cfg,
		users:          make(map[string]*User),
		usersDir:       "./users",
		permissions:    make(map[string]*UserConnectionPermission),
		permissionsDir: "./permissions",
		encryptionKey:  []byte(cfg.PasswordSalt[:32]), // 使用密码盐作为加密密钥
	}

	// 创建用户目录
	if err := os.MkdirAll(manager.usersDir, 0755); err != nil {
		log.Error("Failed to create users directory: %v", err)
	}

	// 创建权限目录
	if err := os.MkdirAll(manager.permissionsDir, 0755); err != nil {
		log.Error("Failed to create permissions directory: %v", err)
	}

	// 加载现有用户
	manager.loadUsers()

	// 加载现有权限
	manager.loadPermissions()

	// 如果没有用户，创建默认管理员
	if len(manager.users) == 0 {
		manager.createDefaultAdmin()
	}

	return manager
}

// Login 用户登录
func (m *AuthManager) Login(c *gin.Context) {
	var loginReq struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 查找用户
	user, exists := m.getUserByUsername(loginReq.Username)
	if !exists || !user.Active {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// 验证密码
	if !m.verifyPassword(loginReq.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// 生成JWT令牌
	token, err := m.generateToken(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})

	log.Info("User logged in: %s", user.Username)
}

// Register 用户注册
func (m *AuthManager) Register(c *gin.Context) {
	var registerReq struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Email    string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&registerReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	// 检查用户名是否已存在
	if _, exists := m.getUserByUsername(registerReq.Username); exists {
		c.JSON(http.StatusConflict, gin.H{"error": "Username already exists"})
		return
	}

	// 创建新用户
	user := &User{
		ID:        m.GenerateID(),
		Username:  registerReq.Username,
		Password:  m.HashPassword(registerReq.Password),
		Email:     registerReq.Email,
		Role:      "user", // 默认角色
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存用户
	if err := m.SaveUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})

	log.Info("User registered: %s", user.Username)
}

// GetMe 获取当前用户信息
func (m *AuthManager) GetMe(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// 查找用户
	user, exists := m.users[userID.(string)]
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"role":       user.Role,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

// AuthMiddleware 认证中间件
func (m *AuthManager) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string

		// 1. 优先从Authorization头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			// 提取令牌
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) == 2 && tokenParts[0] == "Bearer" {
				tokenString = tokenParts[1]
			}
		}

		// 2. 如果Authorization头没有提供令牌，从URL查询参数获取
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		// 3. 如果还是没有令牌，返回错误
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// 解析和验证令牌
		claims, err := m.parseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// RoleMiddleware 角色中间件
func (m *AuthManager) RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户角色
		userRole, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}

		// 检查角色是否在允许列表中
		role := userRole.(string)
		allowed := false
		for _, r := range roles {
			if r == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// generateToken 生成JWT令牌
func (m *AuthManager) generateToken(user *User) (string, error) {
	// 设置过期时间
	expirationTime := time.Now().Add(time.Duration(m.cfg.JWTExpiresIn) * time.Second)

	// 创建声明
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID,
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(m.cfg.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// parseToken 解析JWT令牌
func (m *AuthManager) parseToken(tokenString string) (*Claims, error) {
	// 解析令牌
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(m.cfg.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

// HashPassword 密码哈希
func (m *AuthManager) HashPassword(password string) string {
	// 使用MD5哈希密码（实际项目中应使用更安全的哈希算法如bcrypt）
	hash := md5.Sum([]byte(password + m.cfg.PasswordSalt))
	return hex.EncodeToString(hash[:])
}

// verifyPassword 验证密码
func (m *AuthManager) verifyPassword(password, hashedPassword string) bool {
	return m.HashPassword(password) == hashedPassword
}

// GenerateID 生成唯一ID
func (m *AuthManager) GenerateID() string {
	// 生成随机ID
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// getUserByUsername 通过用户名获取用户
func (m *AuthManager) getUserByUsername(username string) (*User, bool) {
	for _, user := range m.users {
		if user.Username == username {
			return user, true
		}
	}
	return nil, false
}

// loadUsers 加载所有用户
func (m *AuthManager) loadUsers() {
	// 读取目录中的所有用户文件
	files, err := os.ReadDir(m.usersDir)
	if err != nil {
		log.Error("Failed to read users directory: %v", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filename := filepath.Join(m.usersDir, file.Name())
		content, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Error("Failed to read user file %s: %v", filename, err)
			continue
		}

		var user User
		if err := json.Unmarshal(content, &user); err != nil {
			log.Error("Failed to unmarshal user file %s: %v", filename, err)
			continue
		}

		// 添加到内存
		m.users[user.ID] = &user
		log.Info("Loaded user: %s", user.Username)
	}
}

// SaveUser 保存用户
func (m *AuthManager) SaveUser(user *User) error {
	// 更新时间
	user.UpdatedAt = time.Now()

	// 保存到文件
	filename := filepath.Join(m.usersDir, user.ID+".json")
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create user file: %v", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(user); err != nil {
		return fmt.Errorf("failed to encode user: %v", err)
	}

	// 添加到内存
	m.users[user.ID] = user

	return nil
}

// GetAllUsers 获取所有用户
func (m *AuthManager) GetAllUsers() []*User {
	users := make([]*User, 0, len(m.users))
	for _, user := range m.users {
		users = append(users, user)
	}
	return users
}

// UpdateUser 更新用户信息
func (m *AuthManager) UpdateUser(userID string, updatedUser *User) error {
	// 检查用户是否存在
	user, exists := m.users[userID]
	if !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	// 更新用户信息
	if updatedUser.Username != "" && updatedUser.Username != user.Username {
		// 检查用户名是否已存在
		if _, exists := m.getUserByUsername(updatedUser.Username); exists {
			return fmt.Errorf("username already exists: %s", updatedUser.Username)
		}
		user.Username = updatedUser.Username
	}

	if updatedUser.Email != "" {
		user.Email = updatedUser.Email
	}

	if updatedUser.Password != "" {
		user.Password = m.HashPassword(updatedUser.Password)
	}

	if updatedUser.Role != "" {
		user.Role = updatedUser.Role
	}

	user.Active = updatedUser.Active
	user.UpdatedAt = time.Now()

	// 保存更新后的用户
	return m.SaveUser(user)
}

// DeleteUser 删除用户
func (m *AuthManager) DeleteUser(userID string) error {
	// 检查用户是否存在
	if _, exists := m.users[userID]; !exists {
		return fmt.Errorf("user not found: %s", userID)
	}

	// 删除用户文件
	filename := filepath.Join(m.usersDir, userID+".json")
	if err := os.Remove(filename); err != nil {
		return fmt.Errorf("failed to delete user file: %v", err)
	}

	// 删除用户的所有权限
	if err := m.DeleteUserPermissionsByUser(userID); err != nil {
		log.Error("Failed to delete user permissions: %v", err)
	}

	// 从内存中删除用户
	delete(m.users, userID)

	return nil
}

// GetUserConnections 获取用户有权限访问的连接
func (m *AuthManager) GetUserConnections(userID string, connManager *connection.ConnectionManager) []connection.Connection {
	// 如果是管理员，返回所有连接
	user, exists := m.users[userID]
	if !exists {
		return []connection.Connection{}
	}

	if user.Role == "admin" {
		return connManager.GetConnections()
	}

	// 非管理员用户，只返回有权限的连接
	// 获取用户的所有权限
	permissions := m.GetUserPermissions(userID)
	if len(permissions) == 0 {
		return []connection.Connection{}
	}

	// 获取所有连接
	allConnections := connManager.GetConnections()
	// 过滤出用户有权限的连接
	var userConnections []connection.Connection
	permissionMap := make(map[string]bool)

	// 构建权限映射
	for _, perm := range permissions {
		permissionMap[perm.ConnectionID] = true
	}

	// 过滤连接
	for _, conn := range allConnections {
		if permissionMap[conn.ID] {
			userConnections = append(userConnections, conn)
		}
	}

	return userConnections
}

// createDefaultAdmin 创建默认管理员用户
func (m *AuthManager) createDefaultAdmin() {
	admin := &User{
		ID:        m.GenerateID(),
		Username:  "admin",
		Password:  m.HashPassword("admin123"),
		Email:     "admin@example.com",
		Role:      "admin",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := m.SaveUser(admin); err != nil {
		log.Error("Failed to create default admin: %v", err)
		return
	}

	log.Info("Created default admin user: admin/admin123")
}
