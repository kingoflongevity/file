package log

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var (
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger

	logLevel int
)

const (
	DEBUG = iota
	INFO
	WARN
	ERROR
)

// InitLogger 初始化日志系统
func InitLogger(level string) {
	// 设置日志级别
	switch strings.ToLower(level) {
	case "debug":
		logLevel = DEBUG
	case "info":
		logLevel = INFO
	case "warn":
		logLevel = WARN
	case "error":
		logLevel = ERROR
	default:
		logLevel = INFO
	}

	// 创建日志文件
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	logFile := fmt.Sprintf("%s/%s.log", logDir, time.Now().Format("2006-01-02"))
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// 创建多写入器：同时写入文件和控制台
	multiWriter := io.MultiWriter(file, os.Stdout)

	// 初始化日志记录器
	infoLogger = log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warnLogger = log.New(multiWriter, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger = log.New(multiWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
	debugLogger = log.New(multiWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

	Info("Logger initialized with level: %s", level)
}

// Debug 记录调试日志
func Debug(format string, v ...interface{}) {
	if logLevel <= DEBUG {
		debugLogger.Printf(format, v...)
	}
}

// Info 记录信息日志
func Info(format string, v ...interface{}) {
	if logLevel <= INFO {
		infoLogger.Printf(format, v...)
	}
}

// Warn 记录警告日志
func Warn(format string, v ...interface{}) {
	if logLevel <= WARN {
		warnLogger.Printf(format, v...)
	}
}

// Error 记录错误日志
func Error(format string, v ...interface{}) {
	if logLevel <= ERROR {
		errorLogger.Printf(format, v...)
	}
}

// Fatalf 记录致命错误并退出程序
func Fatalf(format string, v ...interface{}) {
	errorLogger.Printf(format, v...)
	os.Exit(1)
}
