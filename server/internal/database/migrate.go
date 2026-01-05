package database

import (
	"log"
)

// Migrate 执行数据库迁移，创建所需的表结构
func Migrate() error {
	// 创建用户表
	userTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		username TEXT UNIQUE NOT NULL,
		password TEXT NOT NULL,
		salt TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`

	// 创建SSH连接表（支持多种连接类型）
	sshConnectionTableSQL := `
	CREATE TABLE IF NOT EXISTS ssh_connections (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		type TEXT NOT NULL DEFAULT 'ssh',
		host TEXT NOT NULL,
		port INTEGER NOT NULL DEFAULT 22,
		username TEXT NOT NULL,
		password TEXT,
		private_key TEXT,
		passphrase TEXT,
		is_active BOOLEAN NOT NULL DEFAULT true,
		category TEXT,
		last_used DATETIME,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`

	// 创建SSH分类表
	sshCategoryTableSQL := `
	CREATE TABLE IF NOT EXISTS ssh_categories (
		name TEXT PRIMARY KEY,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	`

	// 创建用户连接权限表
	userConnectionPermissionSQL := `
	CREATE TABLE IF NOT EXISTS user_connection_permissions (
		id TEXT PRIMARY KEY,
		user_id TEXT NOT NULL,
		connection_id TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY (connection_id) REFERENCES ssh_connections(id) ON DELETE CASCADE,
		UNIQUE(user_id, connection_id)
	);
	`

	// 执行创建表语句
	createQueries := []string{userTableSQL, sshConnectionTableSQL, sshCategoryTableSQL, userConnectionPermissionSQL}
	for _, query := range createQueries {
		_, err := DB.Exec(query)
		if err != nil {
			return err
		}
	}

	// 尝试添加type字段到已有表（兼容旧数据库）
	// SQLite不支持IF NOT EXISTS语法，所以单独执行并忽略错误
	addTypeColumnSQL := `
	ALTER TABLE ssh_connections ADD COLUMN type TEXT NOT NULL DEFAULT 'ssh';
	`
	if _, err := DB.Exec(addTypeColumnSQL); err != nil {
		// 忽略错误，因为列可能已存在
		log.Println("Warning: Failed to add type column (it may already exist):", err)
	}

	log.Println("Database migration completed successfully")
	return nil
}
