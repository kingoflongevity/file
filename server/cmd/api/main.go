package main

import (
	"embed"
	"fmt"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"remote-file-manager/internal/auth"
	"remote-file-manager/internal/config"
	"remote-file-manager/internal/connection"
	"remote-file-manager/internal/database"
	"remote-file-manager/internal/file"
	"remote-file-manager/internal/log"
	"remote-file-manager/internal/middleware"
	"remote-file-manager/internal/task"

	"github.com/gin-gonic/gin"
)

//go:embed client
var clientDist embed.FS

func main() {
	// 加载配置
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化日志
	log.InitLogger(cfg.LogLevel)

	// 初始化数据库
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDatabase()

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

	// 初始化连接管理器
	connManager := connection.NewConnectionManager(cfg)

	// 初始化文件管理器
	fileManager := file.NewFileManager(connManager)

	// 初始化认证管理器
	authManager := auth.NewAuthManager(cfg)

	// 初始化任务管理器
	taskManager := task.NewTaskManager()

	// 注册路由
	registerRoutes(r, authManager, fileManager, connManager, taskManager)

	// 提供静态文件服务
	r.StaticFS("/", http.FS(clientDist))

	// 启动服务
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("Server starting on port %d", cfg.Port)

	// 自动打开浏览器
	go func() {
		time.Sleep(2 * time.Second)
		url := fmt.Sprintf("http://localhost:%d", cfg.Port)
		var err error

		switch runtime.GOOS {
		case "windows":
			err = exec.Command("cmd", "/c", "start", url).Start()
		case "darwin":
			err = exec.Command("open", url).Start()
		case "linux":
			err = exec.Command("xdg-open", url).Start()
		default:
			log.Warn("Unsupported platform, cannot open browser automatically")
		}

		if err != nil {
			log.Warn("Failed to open browser: %v", err)
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func registerRoutes(r *gin.Engine, authManager *auth.AuthManager, fileManager *file.FileManager, sshManager *connection.ConnectionManager, taskManager *task.TaskManager) {
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

	// 连接管理路由
	connRoutes := r.Group("/api/connections")
	connRoutes.Use(authManager.AuthMiddleware())
	{
		connRoutes.POST("", func(c *gin.Context) {
			var conn connection.Connection
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
		connRoutes.GET("", func(c *gin.Context) {
			conns := sshManager.GetConnections()
			c.JSON(http.StatusOK, conns)
		})
		connRoutes.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			conn, exists := sshManager.GetConnection(id)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
				return
			}
			c.JSON(http.StatusOK, conn)
		})
		connRoutes.PUT("/:id", func(c *gin.Context) {
			id := c.Param("id")
			var conn connection.Connection
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
		connRoutes.DELETE("/:id", func(c *gin.Context) {
			id := c.Param("id")
			if err := sshManager.DeleteConnection(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Connection deleted"})
		})
		connRoutes.POST("/:id/test", func(c *gin.Context) {
			id := c.Param("id")
			if err := sshManager.TestConnection(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Connection test successful"})
		})

		// 测试未保存的连接
		connRoutes.POST("/test", func(c *gin.Context) {
			var conn connection.Connection
			if err := c.ShouldBindJSON(&conn); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			// 创建临时连接ID用于测试
			if conn.ID == "" {
				conn.ID = fmt.Sprintf("temp-test-%d", time.Now().UnixNano())
			}

			// 设置默认值
			if conn.Port == 0 {
				switch conn.Type {
				case connection.ConnectionTypeSSH, connection.ConnectionTypeSFTP:
					conn.Port = 22
				case connection.ConnectionTypeFTP:
					conn.Port = 21
				default:
					conn.Port = 22
				}
			}

			// 设置时间字段
			now := time.Now()
			conn.CreatedAt = now
			conn.UpdatedAt = now

			// 加密敏感信息
			conn.Password = sshManager.Encrypt(conn.Password)
			conn.PrivateKey = sshManager.Encrypt(conn.PrivateKey)
			conn.Passphrase = sshManager.Encrypt(conn.Passphrase)

			// 将临时连接添加到内存映射中
			sshManager.Connections()[conn.ID] = &conn

			// 测试连接
			client, err := sshManager.GetClient(conn.ID)
			if err != nil {
				// 无论测试成功与否，都要移除临时连接
				sshManager.RemoveConnection(conn.ID)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			// 关闭连接
			client.Close()

			// 从连接池和内存中移除临时连接
			sshManager.RemoveConnection(conn.ID)

			c.JSON(http.StatusOK, gin.H{"message": "Connection test successful"})
		})
	}

	// 旧API兼容性支持 - 保持与现有客户端兼容
	oldSSHRoutes := r.Group("/api/ssh")
	oldSSHRoutes.Use(authManager.AuthMiddleware())
	{
		// 旧连接管理路由的重定向
		oldSSHRoutes.POST("/connections", func(c *gin.Context) {
			var conn connection.Connection
			if err := c.ShouldBindJSON(&conn); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			// 设置默认连接类型为SSH，保持与旧客户端兼容
			if conn.Type == "" {
				conn.Type = connection.ConnectionTypeSSH
			}
			if err := sshManager.AddConnection(nil, &conn); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, conn)
		})
		oldSSHRoutes.GET("/connections", func(c *gin.Context) {
			conns := sshManager.GetConnections()
			c.JSON(http.StatusOK, conns)
		})
		oldSSHRoutes.GET("/connections/:id", func(c *gin.Context) {
			id := c.Param("id")
			conn, exists := sshManager.GetConnection(id)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
				return
			}
			c.JSON(http.StatusOK, conn)
		})
		oldSSHRoutes.PUT("/connections/:id", func(c *gin.Context) {
			id := c.Param("id")
			var conn connection.Connection
			if err := c.ShouldBindJSON(&conn); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			// 设置默认连接类型为SSH，保持与旧客户端兼容
			if conn.Type == "" {
				conn.Type = connection.ConnectionTypeSSH
			}
			if err := sshManager.UpdateConnection(id, &conn); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, conn)
		})
		oldSSHRoutes.DELETE("/connections/:id", func(c *gin.Context) {
			id := c.Param("id")
			if err := sshManager.DeleteConnection(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Connection deleted"})
		})
		oldSSHRoutes.POST("/connections/:id/test", func(c *gin.Context) {
			id := c.Param("id")
			if err := sshManager.TestConnection(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Connection test successful"})
		})

		// 旧分类管理路由的重定向
		oldSSHRoutes.GET("/categories", func(c *gin.Context) {
			categories := sshManager.GetCategories()
			c.JSON(http.StatusOK, categories)
		})
		oldSSHRoutes.POST("/categories", func(c *gin.Context) {
			var req struct {
				Name string `json:"name" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			if err := sshManager.AddCategory(req.Name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"message": "Category added", "name": req.Name})
		})
		oldSSHRoutes.PUT("/categories/:name", func(c *gin.Context) {
			oldName := c.Param("name")
			var req struct {
				NewName string `json:"new_name" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			if err := sshManager.UpdateCategory(oldName, req.NewName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Category updated", "old_name": oldName, "new_name": req.NewName})
		})
		oldSSHRoutes.DELETE("/categories/:name", func(c *gin.Context) {
			name := c.Param("name")
			if err := sshManager.DeleteCategory(name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Category deleted", "name": name})
		})
	}

	// 分类管理路由
	catRoutes := r.Group("/api/categories")
	catRoutes.Use(authManager.AuthMiddleware())
	{
		catRoutes.GET("", func(c *gin.Context) {
			categories := sshManager.GetCategories()
			c.JSON(http.StatusOK, categories)
		})

		catRoutes.POST("", func(c *gin.Context) {
			var req struct {
				Name string `json:"name" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			if err := sshManager.AddCategory(req.Name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusCreated, gin.H{"message": "Category added", "name": req.Name})
		})

		catRoutes.PUT("/:name", func(c *gin.Context) {
			oldName := c.Param("name")
			var req struct {
				NewName string `json:"new_name" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}
			if err := sshManager.UpdateCategory(oldName, req.NewName); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Category updated", "old_name": oldName, "new_name": req.NewName})
		})

		catRoutes.DELETE("/:name", func(c *gin.Context) {
			name := c.Param("name")
			if err := sshManager.DeleteCategory(name); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Category deleted", "name": name})
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

			// 创建新任务
			newTask := taskManager.CreateTask("download", connID, path)
			log.Info("Created download task: %s, path: %s", newTask.ID, path)

			// 更新任务状态为运行中
			taskManager.UpdateTaskStatus(newTask.ID, task.TaskStatusRunning)
			log.Info("Updated task status: %s to running", newTask.ID)

			// 保存任务ID到局部变量，避免闭包问题
			taskID := newTask.ID
			// 异步执行下载任务
			go func() {
				log.Info("Starting download task: %s", taskID)
				// 更新任务状态为运行中
				taskManager.UpdateTaskStatus(taskID, task.TaskStatusRunning)

				// 模拟下载进度，每500ms更新一次
				go func() {
					log.Info("Starting progress updates for task: %s", taskID)
					for i := 0; i <= 100; i += 10 {
						time.Sleep(500 * time.Millisecond)
						taskManager.UpdateTaskProgress(taskID, i)
						log.Info("Updated task progress: %s, %d%%", taskID, i)
					}
				}()

				// 执行下载操作
				content, err := fileManager.DownloadFile(connID, path)
				if err != nil {
					// 如果是目录，创建压缩任务
					if strings.Contains(err.Error(), "cannot download directory") {
						// 重新创建任务，类型改为zip
						taskManager.DeleteTask(taskID)
						zipTask := taskManager.CreateTask("zip", connID, path)
						taskManager.UpdateTaskStatus(zipTask.ID, task.TaskStatusRunning)
						log.Info("Created zip task: %s for directory download", zipTask.ID)

						// 保存zip任务ID到局部变量
						zipTaskID := zipTask.ID
						// 异步执行压缩任务
						go func() {
							// 执行压缩操作，带进度更新
							zipContent, err := fileManager.ZipDirectoryWithProgress(connID, path, func(progress int) {
								// 更新任务进度
								taskManager.UpdateTaskProgress(zipTaskID, progress)
								log.Info("Updated zip task progress: %s, %d%%", zipTaskID, progress)
							})

							if err != nil {
								// 更新任务状态为失败
								taskManager.UpdateTaskError(zipTaskID, err.Error())
								log.Error("Zip task failed: %s, error: %s", zipTaskID, err.Error())
								return
							}

							// 生成文件名
							fileName := filepath.Base(path) + ".zip"

							// 缓存文件内容
							taskManager.UpdateTaskContent(zipTaskID, zipContent)
							// 更新任务状态和文件名
							taskManager.UpdateTaskFileName(zipTaskID, fileName)
							taskManager.UpdateTaskStatus(zipTaskID, task.TaskStatusCompleted)
							log.Info("Zip task completed: %s", zipTaskID)
						}()

						return
					}

					// 其他错误，更新任务状态为失败
					taskManager.UpdateTaskError(taskID, err.Error())
					log.Error("Download task failed: %s, error: %s", taskID, err.Error())
					return
				}

				// 更新任务进度为100%
				taskManager.UpdateTaskProgress(taskID, 100)
				log.Info("Download task completed: %s", taskID)

				// 生成文件名
				fileName := filepath.Base(path)
				// 检查是否为ZIP文件（ZIP文件以PK开头）
				if len(content) >= 4 && content[0] == 'P' && content[1] == 'K' && content[2] == 0x03 && content[3] == 0x04 {
					// 如果是ZIP文件，确保文件名以.zip结尾
					if !strings.HasSuffix(fileName, ".zip") {
						fileName += ".zip"
					}
				}

				// 缓存文件内容
				taskManager.UpdateTaskContent(taskID, content)
				// 更新任务状态和文件名
				taskManager.UpdateTaskFileName(taskID, fileName)
				taskManager.UpdateTaskStatus(taskID, task.TaskStatusCompleted)
				log.Info("Updated task %s to completed with fileName: %s, content cached: %d bytes", taskID, fileName, len(content))
			}()

			// 等待短暂时间，确保任务已创建
			time.Sleep(100 * time.Millisecond)

			// 设置CORS头，允许前端跨域访问
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// 返回JSON响应，让前端在页面内处理
			c.JSON(http.StatusAccepted, gin.H{
				"taskId":  newTask.ID,
				"message": "Download task created successfully",
				"status":  "accepted",
				"type":    "download_task",
			})
		})
		// 批量下载文件
		fileRoutes.POST("/download/batch", func(c *gin.Context) {
			var req struct {
				ConnID string   `json:"connId" binding:"required"`
				Paths  []string `json:"paths" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid request",
					"type":  "error",
				})
				return
			}

			// 创建新任务，用于批量下载
			newTask := taskManager.CreateTask("batch_zip", req.ConnID, fmt.Sprintf("%v", req.Paths))

			// 异步执行压缩任务
			go func() {
				// 更新任务状态为运行中
				taskManager.UpdateTaskStatus(newTask.ID, task.TaskStatusRunning)

				// 执行批量压缩操作，带进度更新
				zipContent, err := fileManager.ZipFilesWithProgress(req.ConnID, req.Paths, func(progress int) {
					// 更新任务进度
					taskManager.UpdateTaskProgress(newTask.ID, progress)
				})

				if err != nil {
					// 更新任务状态为失败
					taskManager.UpdateTaskError(newTask.ID, err.Error())
					return
				}

				// 生成文件名
				zipFileName := "batch-download.zip"
				if len(req.Paths) == 1 {
					zipFileName = filepath.Base(req.Paths[0]) + ".zip"
				}

				// 缓存文件内容
				taskManager.UpdateTaskContent(newTask.ID, zipContent)
				// 更新任务状态和文件名
				taskManager.UpdateTaskFileName(newTask.ID, zipFileName)
				taskManager.UpdateTaskStatus(newTask.ID, task.TaskStatusCompleted)
			}()

			// 设置CORS头，允许前端跨域访问
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// 返回JSON响应，让前端在页面内处理
			c.JSON(http.StatusAccepted, gin.H{
				"taskId":  newTask.ID,
				"message": "Batch compression task created successfully",
				"status":  "accepted",
				"type":    "batch_zip_task",
			})
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

	// 任务管理路由
	taskRoutes := r.Group("/api/tasks")
	taskRoutes.Use(authManager.AuthMiddleware())
	{
		// 获取所有任务
		taskRoutes.GET("", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"tasks": taskManager.GetAllTasks()})
		})

		// 获取单个任务
		taskRoutes.GET("/:id", func(c *gin.Context) {
			taskID := c.Param("id")
			task, exists := taskManager.GetTask(taskID)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}
			c.JSON(http.StatusOK, task)
		})

		// 创建压缩任务
		taskRoutes.POST("/zip", func(c *gin.Context) {
			var req struct {
				ConnID string `json:"connId" binding:"required"`
				Path   string `json:"path" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			// 创建新任务
			newTask := taskManager.CreateTask("zip", req.ConnID, req.Path)

			// 异步执行压缩任务
			go func() {
				// 更新任务状态为运行中
				taskManager.UpdateTaskStatus(newTask.ID, task.TaskStatusRunning)

				// 执行压缩操作，带进度更新
				_, err := fileManager.ZipDirectoryWithProgress(req.ConnID, req.Path, func(progress int) {
					// 更新任务进度
					taskManager.UpdateTaskProgress(newTask.ID, progress)
				})

				if err != nil {
					// 更新任务状态为失败
					taskManager.UpdateTaskError(newTask.ID, err.Error())
					return
				}

				// 生成文件名
				fileName := filepath.Base(req.Path) + ".zip"

				// 这里应该将生成的文件保存到临时目录或内存中
				// 为了简化，我们暂时只更新任务状态和文件名
				taskManager.UpdateTaskFileName(newTask.ID, fileName)
				taskManager.UpdateTaskStatus(newTask.ID, task.TaskStatusCompleted)
			}()

			// 返回创建的任务
			c.JSON(http.StatusCreated, newTask)
		})

		// 下载任务文件
		taskRoutes.GET("/:id/download", func(c *gin.Context) {
			taskID := c.Param("id")
			task, exists := taskManager.GetTask(taskID)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
				return
			}

			if task.Status != "completed" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Task is not completed"})
				return
			}

			// 直接使用缓存的文件内容
			if task.Content == nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "File content not cached"})
				return
			}

			// 确定文件类型
			contentType := "application/octet-stream"
			if task.Type == "zip" || task.Type == "batch_zip" {
				contentType = "application/zip"
			}

			// 设置Content-Disposition头，触发浏览器下载
			c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", task.FileName))
			c.Data(http.StatusOK, contentType, task.Content)
		})

		// 删除任务
		taskRoutes.DELETE("/:id", func(c *gin.Context) {
			taskID := c.Param("id")
			taskManager.DeleteTask(taskID)
			c.JSON(http.StatusOK, gin.H{"message": "Task deleted"})
		})
	}
}
