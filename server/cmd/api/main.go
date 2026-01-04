package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"file-manager/internal/auth"
	"file-manager/internal/config"
	"file-manager/internal/file"
	"file-manager/internal/log"
	"file-manager/internal/middleware"
	"file-manager/internal/ssh"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	log.InitLogger(cfg.LogLevel)

	// 初始化Gin引擎
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 注册中间件
	r.Use(gin.Recovery())                       // 恢复中间件
	r.Use(middleware.RequestLogger())           // 请求日志中间件
	r.Use(middleware.CORSMiddleware())          // CORS中间件
	r.Use(middleware.Timeout(30 * time.Second)) // 超时控制中间件

	// 创建速率限制器 (100个请求/分钟)
	rateLimiter := middleware.NewRateLimiter(100, 1*time.Minute)
	r.Use(rateLimiter.Limit()) // 速率限制中间件

	// 初始化SSH管理器
	sshManager := ssh.NewSSHManager(cfg)

	// 初始化文件管理器
	fileManager := file.NewFileManager(sshManager)

	// 初始化认证管理器
	authManager := auth.NewAuthManager(cfg)

	// 注册路由
	registerRoutes(r, authManager, fileManager, sshManager)

	// 启动服务
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("Server starting on port %d", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func registerRoutes(r *gin.Engine, authManager *auth.AuthManager, fileManager *file.FileManager, sshManager *ssh.SSHManager) {
	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 认证路由
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.POST("/login", authManager.Login)
		authRoutes.POST("/register", authManager.Register)
		authRoutes.GET("/me", authManager.AuthMiddleware(), authManager.GetMe)
	}

	// SSH连接管理路由
	sshRoutes := r.Group("/api/ssh")
	sshRoutes.Use(authManager.AuthMiddleware())
	{
		sshRoutes.POST("/connections", func(c *gin.Context) {
			var conn ssh.SSHConnection
			if err := c.ShouldBindJSON(&conn); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			if err := sshManager.AddConnection(nil, &conn); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, conn)
		})
		sshRoutes.GET("/connections", func(c *gin.Context) {
			conns := sshManager.GetConnections()
			c.JSON(http.StatusOK, conns)
		})
		sshRoutes.GET("/connections/:id", func(c *gin.Context) {
			id := c.Param("id")
			conn, exists := sshManager.GetConnection(id)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
				return
			}
			c.JSON(http.StatusOK, conn)
		})
		sshRoutes.PUT("/connections/:id", func(c *gin.Context) {
			id := c.Param("id")
			var conn ssh.SSHConnection
			if err := c.ShouldBindJSON(&conn); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			if err := sshManager.UpdateConnection(id, &conn); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, conn)
		})
		sshRoutes.DELETE("/connections/:id", func(c *gin.Context) {
			id := c.Param("id")
			if err := sshManager.DeleteConnection(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Connection deleted"})
		})
		sshRoutes.POST("/connections/:id/test", func(c *gin.Context) {
			id := c.Param("id")
			if err := sshManager.TestConnection(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Connection test successful"})
		})
	}

	// 文件操作路由
	fileRoutes := r.Group("/api/files")
	fileRoutes.Use(authManager.AuthMiddleware())
	{
		// 文件浏览
		fileRoutes.GET("/list", func(c *gin.Context) {
			connID := c.Query("connId")
			path := c.Query("path")
			files, err := fileManager.ListFiles(connID, path)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, files)
		})
		fileRoutes.GET("/tree", func(c *gin.Context) {
			connID := c.Query("connId")
			path := c.Query("path")
			depthStr := c.DefaultQuery("depth", "1")
			depth, _ := strconv.Atoi(depthStr)
			tree, err := fileManager.GetFileTree(connID, path, depth)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, tree)
		})

		// 文件操作
		fileRoutes.POST("/mkdir", func(c *gin.Context) {
			var req struct {
				ConnID string `json:"connId" binding:"required"`
				Path   string `json:"path" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			if err := fileManager.CreateDirectory(req.ConnID, req.Path); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Directory created"})
		})
		fileRoutes.POST("/upload", func(c *gin.Context) {
			// 简单的上传处理，实际应用中应该支持文件流
			c.JSON(http.StatusNotImplemented, gin.H{"error": "Not implemented"})
		})
		fileRoutes.GET("/download/:connId/*path", func(c *gin.Context) {
			connID := c.Param("connId")
			path := c.Param("path")
			content, err := fileManager.DownloadFile(connID, path)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.Data(http.StatusOK, "application/octet-stream", content)
		})
		fileRoutes.DELETE("/delete", func(c *gin.Context) {
			var req struct {
				ConnID string   `json:"connId" binding:"required"`
				Paths  []string `json:"paths" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			if err := fileManager.DeleteFiles(req.ConnID, req.Paths); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Files deleted"})
		})
		fileRoutes.PUT("/rename", func(c *gin.Context) {
			var req struct {
				ConnID  string `json:"connId" binding:"required"`
				OldPath string `json:"oldPath" binding:"required"`
				NewPath string `json:"newPath" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			if err := fileManager.RenameFile(req.ConnID, req.OldPath, req.NewPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "File renamed"})
		})
		fileRoutes.POST("/copy", func(c *gin.Context) {
			var req struct {
				ConnID   string   `json:"connId" binding:"required"`
				SrcPaths []string `json:"srcPaths" binding:"required"`
				DestPath string   `json:"destPath" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			if err := fileManager.CopyFiles(req.ConnID, req.SrcPaths, req.DestPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Files copied"})
		})
		fileRoutes.POST("/move", func(c *gin.Context) {
			var req struct {
				ConnID   string   `json:"connId" binding:"required"`
				SrcPaths []string `json:"srcPaths" binding:"required"`
				DestPath string   `json:"destPath" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			if err := fileManager.MoveFiles(req.ConnID, req.SrcPaths, req.DestPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Files moved"})
		})

		// 文件内容操作
		fileRoutes.GET("/content", func(c *gin.Context) {
			connID := c.Query("connId")
			path := c.Query("path")
			content, err := fileManager.GetFileContent(connID, path)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"content": content})
		})
		fileRoutes.PUT("/content", func(c *gin.Context) {
			var req struct {
				ConnID  string `json:"connId" binding:"required"`
				Path    string `json:"path" binding:"required"`
				Content string `json:"content" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			if err := fileManager.SaveFileContent(req.ConnID, req.Path, req.Content); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "File content saved"})
		})
	}
}
