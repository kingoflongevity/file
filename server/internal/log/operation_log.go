package log

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// OperationLog 操作日志结构体
type OperationLog struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Operation string    `json:"operation"`
	Resource  string    `json:"resource"`
	Action    string    `json:"action"`
	Result    string    `json:"result"`
	Details   string    `json:"details,omitempty"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"created_at"`
}

// OperationLogger 操作日志记录器
type OperationLogger struct {
	logDir string
}

// NewOperationLogger 创建操作日志记录器
func NewOperationLogger() *OperationLogger {
	logDir := "./operation_logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		Error("Failed to create operation logs directory: %v", err)
	}

	return &OperationLogger{
		logDir: logDir,
	}
}

// LogOperation 记录操作日志
func (ol *OperationLogger) LogOperation(log *OperationLog) {
	// 设置默认值
	log.CreatedAt = time.Now()
	log.ID = fmt.Sprintf("%d", log.CreatedAt.UnixNano())

	// 确保日志结果不为空
	if log.Result == "" {
		log.Result = "success"
	}

	// 序列化日志
	logData, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		Error("Failed to marshal operation log: %v", err)
		return
	}

	// 获取日志文件名（按日期分文件）
	logFile := filepath.Join(ol.logDir, fmt.Sprintf("%s.log", log.CreatedAt.Format("2006-01-02")))

	// 打开日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Error("Failed to open operation log file: %v", err)
		return
	}
	defer file.Close()

	// 写入日志
	_, err = file.Write(append(logData, '\n'))
	if err != nil {
		Error("Failed to write operation log: %v", err)
		return
	}

	// 同时输出到控制台
	Info("Operation: %s %s %s by %s (%s) - %s", log.Resource, log.Action, log.Operation, log.Username, log.UserID, log.Result)
}

// LogLogin 记录登录操作
func (ol *OperationLogger) LogLogin(userID, username, ip, result, details string) {
	ol.LogOperation(&OperationLog{
		UserID:    userID,
		Username:  username,
		Operation: "login",
		Resource:  "auth",
		Action:    "login",
		Result:    result,
		Details:   details,
		IP:        ip,
	})
}

// LogLogout 记录登出操作
func (ol *OperationLogger) LogLogout(userID, username, ip string) {
	ol.LogOperation(&OperationLog{
		UserID:    userID,
		Username:  username,
		Operation: "logout",
		Resource:  "auth",
		Action:    "logout",
		Result:    "success",
		IP:        ip,
	})
}

// LogSSHConnection 记录SSH连接操作
func (ol *OperationLogger) LogSSHConnection(userID, username, ip, action, connectionID, result, details string) {
	ol.LogOperation(&OperationLog{
		UserID:    userID,
		Username:  username,
		Operation: "ssh_connection",
		Resource:  "ssh",
		Action:    action,
		Result:    result,
		Details:   fmt.Sprintf("connection_id=%s, %s", connectionID, details),
		IP:        ip,
	})
}

// LogFileOperation 记录文件操作
func (ol *OperationLogger) LogFileOperation(userID, username, ip, action, connID, path, result, details string) {
	ol.LogOperation(&OperationLog{
		UserID:    userID,
		Username:  username,
		Operation: "file_operation",
		Resource:  "file",
		Action:    action,
		Result:    result,
		Details:   fmt.Sprintf("conn_id=%s, path=%s, %s", connID, path, details),
		IP:        ip,
	})
}
