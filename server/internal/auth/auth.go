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

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"file-manager/internal/config"
	"file-manager/internal/log"
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

// Claims JWT声明
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthManager 认证管理器
type AuthManager struct {
	cfg        *config.Config
	users      map[string]*User
	usersDir   string
	encryptionKey []byte
}

// NewAuthManager 创建认证管理器
func NewAuthManager(cfg *config.Config) *AuthManager {
	manager := &AuthManager{
		cfg:        cfg,
		users:      make(map[string]*User),
		usersDir:   "./users",
		encryptionKey: []byte(cfg.PasswordSalt[:32]), // 使用密码盐作为加密密钥
	}

	// 创建用户目录
	if err := os.MkdirAll(manager.usersDir, 0755); err != nil {
		log.Error("Failed to create users directory: %v", err)
	}

	// 加载现有用户
	manager.loadUsers()

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
		ID:        m.generateID(),
		Username:  registerReq.Username,
		Password:  m.hashPassword(registerReq.Password),
		Email:     registerReq.Email,
		Role:      "user", // 默认角色
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 保存用户
	if err := m.saveUser(user); err != nil {
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
		"id":       user.ID,
		"username": user.Username,
		"email":    user.Email,
		"role":     user.Role,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	})
}

// AuthMiddleware 认证中间件
func (m *AuthManager) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从Authorization头获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// 提取令牌
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		tokenString := tokenParts[1]

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

// hashPassword 密码哈希
func (m *AuthManager) hashPassword(password string) string {
	// 使用MD5哈希密码（实际项目中应使用更安全的哈希算法如bcrypt）
	hash := md5.Sum([]byte(password + m.cfg.PasswordSalt))
	return hex.EncodeToString(hash[:])
}

// verifyPassword 验证密码
func (m *AuthManager) verifyPassword(password, hashedPassword string) bool {
	return m.hashPassword(password) == hashedPassword
}

// generateID 生成唯一ID
func (m *AuthManager) generateID() string {
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

// saveUser 保存用户
func (m *AuthManager) saveUser(user *User) error {
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

// createDefaultAdmin 创建默认管理员用户
func (m *AuthManager) createDefaultAdmin() {
	admin := &User{
		ID:        m.generateID(),
		Username:  "admin",
		Password:  m.hashPassword("admin123"),
		Email:     "admin@example.com",
		Role:      "admin",
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := m.saveUser(admin); err != nil {
		log.Error("Failed to create default admin: %v", err)
		return
	}

	log.Info("Created default admin user: admin/admin123")
}
