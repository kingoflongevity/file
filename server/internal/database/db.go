package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB 全局数据库连接
var DB *sql.DB

// InitDatabase 初始化数据库连接
func InitDatabase() error {
	// 确保数据目录存在
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	// 数据库文件路径
	dbPath := filepath.Join(dataDir, "remote-file-manager.db")

	// 打开数据库连接
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}

	// 测试连接
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	log.Printf("Database connected successfully: %s", dbPath)
	DB = db

	// 执行数据库迁移
	if err := Migrate(); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	return nil
}

// CloseDatabase 关闭数据库连接
func CloseDatabase() error {
	if DB != nil {
		return DB.Close()
	}
	return nil
}
