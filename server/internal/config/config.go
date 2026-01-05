package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

// Config 应用配置结构体
type Config struct {
	// 服务器配置
	Port int

	// 日志配置
	LogLevel string

	// 认证配置
	JWTSecret    string
	JWTExpiresIn int
	PasswordSalt string

	// 数据库配置
	DBPath string

	// SSH配置
	MaxSSHConnections int

	// 下载配置
	DefaultDownloadDir string

	// 网站配置
	SiteName string
}

// LoadConfig 加载配置
func LoadConfig() (*Config, error) {
	// 尝试加载.env文件，如果存在的话
	_ = godotenv.Load()

	// 解析端口
	portStr := getEnv("PORT", "8080")
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, fmt.Errorf("invalid PORT: %v", err)
	}

	// 解析JWT过期时间
	jwtExpiresStr := getEnv("JWT_EXPIRES_IN", "7200")
	jwtExpires, err := strconv.Atoi(jwtExpiresStr)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRES_IN: %v", err)
	}

	// 解析最大SSH连接数
	maxSSHConnStr := getEnv("MAX_SSH_CONNECTIONS", "10")
	maxSSHConn, err := strconv.Atoi(maxSSHConnStr)
	if err != nil {
		return nil, fmt.Errorf("invalid MAX_SSH_CONNECTIONS: %v", err)
	}

	// 获取用户主目录，用于设置默认下载目录
	userHome, err := os.UserHomeDir()
	if err != nil {
		// 如果获取失败，使用当前目录
		userHome = "."
	}
	defaultDownloadDir := filepath.Join(userHome, "Downloads", "RemoteFileManager")

	// 构建配置
	cfg := &Config{
		Port:               port,
		LogLevel:           getEnv("LOG_LEVEL", "info"),
		JWTSecret:          getEnv("JWT_SECRET", "your-secret-key-change-me"),
		JWTExpiresIn:       jwtExpires,
		PasswordSalt:       getEnv("PASSWORD_SALT", "your-very-secure-password-salt-change-me-now"),
		DBPath:             getEnv("DB_PATH", "./data.db"),
		MaxSSHConnections:  maxSSHConn,
		DefaultDownloadDir: getEnv("DEFAULT_DOWNLOAD_DIR", defaultDownloadDir),
		SiteName:           getEnv("SITE_NAME", "远程连接文件管理"),
	}

	return cfg, nil
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
