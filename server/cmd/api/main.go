package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
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

// 开发环境下使用文件系统，生产环境下嵌入静态文件
// 注意：实际部署时需要确保client目录存在或使用嵌入的静态文件
//
//go:embed client/dist
var clientDist embed.FS

// 在开发环境中，我们从文件系统读取静态文件，而不是使用嵌入的文件
const isDevMode = true

// 读取静态文件的辅助函数，根据isDevMode决定从嵌入文件系统还是实际文件系统读取
func readStaticFile(filePath string) ([]byte, error) {
	if isDevMode {
		// 开发模式：从实际文件系统读取
		return os.ReadFile(fmt.Sprintf("client/dist/%s", filePath))
	} else {
		// 生产模式：从嵌入文件系统读取
		return clientDist.ReadFile(filePath)
	}
}

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

	// 注册路由 - 先注册所有API路由
	registerRoutes(r, authManager, fileManager, connManager, taskManager, cfg)

	// 静态文件服务 - 使用最简单的方法，避免重定向问题

	// 根路径 - 返回index.html
	r.GET("/", func(c *gin.Context) {
		content, err := readStaticFile("index.html")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read index.html"})
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	})

	// 处理assets目录下的静态资源
	r.GET("/assets/*filepath", func(c *gin.Context) {
		filepath := c.Param("filepath")
		// 移除filepath前面的斜杠
		filepath = strings.TrimPrefix(filepath, "/")
		content, err := readStaticFile(fmt.Sprintf("assets/%s", filepath))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		// 设置正确的Content-Type
		contentType := "application/octet-stream"
		if strings.HasSuffix(filepath, ".css") {
			contentType = "text/css; charset=utf-8"
		} else if strings.HasSuffix(filepath, ".js") {
			contentType = "application/javascript; charset=utf-8"
		} else if strings.HasSuffix(filepath, ".svg") {
			contentType = "image/svg+xml"
		}

		c.Data(http.StatusOK, contentType, content)
	})

	// 处理vite.svg文件
	r.GET("/vite.svg", func(c *gin.Context) {
		content, err := readStaticFile("vite.svg")
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}
		c.Data(http.StatusOK, "image/svg+xml", content)
	})

	// 处理index.html直接请求
	r.GET("/index.html", func(c *gin.Context) {
		content, err := readStaticFile("index.html")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read index.html"})
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", content)
	})

	// 兜底路由 - 处理所有其他非API GET请求
	// 用于单页应用路由，所有前端路由请求都返回index.html
	r.NoRoute(func(c *gin.Context) {
		// 检查是否是API请求
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		// 检查是否是健康检查请求
		if c.Request.URL.Path == "/health" {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
			return
		}

		// 所有其他GET请求返回index.html - 使用c.Data避免重定向
		if c.Request.Method == "GET" {
			content, err := readStaticFile("index.html")
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to serve index.html"})
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", content)
			return
		}

		// 其他HTTP方法返回404
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	})

	// 启动服务
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Info("Server starting on port %d", cfg.Port)

	// // 自动打开浏览器 - 已禁用，改为在用户登录后提示
	// go func() {
	// 	time.Sleep(2 * time.Second)
	// 	url := fmt.Sprintf("http://localhost:%d", cfg.Port)
	// 	var err error
	//
	// 	switch runtime.GOOS {
	// 	case "windows":
	// 		err = exec.Command("cmd", "/c", "start", url).Start()
	// 	case "darwin":
	// 		err = exec.Command("open", url).Start()
	// 	case "linux":
	// 		err = exec.Command("xdg-open", url).Start()
	// 	default:
	// 		log.Warn("Unsupported platform, cannot open browser automatically")
	// 	}
	//
	// 	if err != nil {
	// 		log.Warn("Failed to open browser: %v", err)
	// 	}
	// }()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// 保存文件到默认下载目录
func saveFileToDefaultDir(cfg *config.Config, fileName string, content []byte) error {
	// 确保默认下载目录存在
	if err := os.MkdirAll(cfg.DefaultDownloadDir, 0755); err != nil {
		log.Error("Failed to create download directory: %s, error: %s", cfg.DefaultDownloadDir, err.Error())
		return err
	}

	// 构建完整的文件路径
	filePath := filepath.Join(cfg.DefaultDownloadDir, fileName)

	// 检查文件是否已存在，如果存在则添加时间戳
	if _, err := os.Stat(filePath); err == nil {
		// 文件已存在，添加时间戳
		ext := filepath.Ext(fileName)
		baseName := strings.TrimSuffix(fileName, ext)
		timestamp := time.Now().Format("20060102_150405")
		fileName = fmt.Sprintf("%s_%s%s", baseName, timestamp, ext)
		filePath = filepath.Join(cfg.DefaultDownloadDir, fileName)
	}

	// 写入文件
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		log.Error("Failed to write file to download directory: %s, error: %s", filePath, err.Error())
		return err
	}

	log.Info("File saved to download directory: %s, size: %d bytes", filePath, len(content))
	return nil
}

func registerRoutes(r *gin.Engine, authManager *auth.AuthManager, fileManager *file.FileManager, sshManager *connection.ConnectionManager, taskManager *task.TaskManager, cfg *config.Config) {
	// 健康检查路由
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 认证路由
	authRoutes := r.Group("/api/auth")
	{
		authRoutes.POST("/login", authManager.Login)
		authRoutes.GET("/me", authManager.AuthMiddleware(), authManager.GetMe)
	}

	// 连接管理路由
	connRoutes := r.Group("/api/connections")
	connRoutes.Use(authManager.AuthMiddleware())
	{
		// 创建连接 - 只有管理员可以创建
		connRoutes.POST("", authManager.RoleMiddleware("admin"), func(c *gin.Context) {
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

		// 获取连接列表 - 根据用户角色返回相应连接
		connRoutes.GET("", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			userRole, _ := c.Get("role")

			var conns []connection.Connection
			if userRole == "admin" {
				conns = sshManager.GetConnections()
			} else {
				conns = authManager.GetUserConnections(userID.(string), sshManager)
			}

			c.JSON(http.StatusOK, conns)
		})

		// 获取单个连接 - 检查用户是否有权限访问
		connRoutes.GET("/:id", func(c *gin.Context) {
			id := c.Param("id")
			userID, _ := c.Get("user_id")
			userRole, _ := c.Get("role")

			// 检查连接是否存在
			conn, exists := sshManager.GetConnection(id)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
				return
			}

			// 管理员可以访问所有连接，非管理员需要检查权限
			if userRole != "admin" {
				// 获取用户有权限的连接
				userConns := authManager.GetUserConnections(userID.(string), sshManager)
				// 检查用户是否有权限访问该连接
				hasPermission := false
				for _, userConn := range userConns {
					if userConn.ID == id {
						hasPermission = true
						break
					}
				}
				if !hasPermission {
					c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
					return
				}
			}

			c.JSON(http.StatusOK, conn)
		})

		// 更新连接 - 只有管理员可以更新
		connRoutes.PUT("/:id", authManager.RoleMiddleware("admin"), func(c *gin.Context) {
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

		// 删除连接 - 只有管理员可以删除
		connRoutes.DELETE("/:id", authManager.RoleMiddleware("admin"), func(c *gin.Context) {
			id := c.Param("id")
			if err := sshManager.DeleteConnection(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			// 删除与该连接相关的所有权限
			if err := authManager.DeleteUserPermissionsByConnection(id); err != nil {
				log.Error("Failed to delete connection permissions: %v", err)
			}
			c.JSON(http.StatusOK, gin.H{"message": "Connection deleted"})
		})

		// 测试连接 - 检查用户是否有权限访问
		connRoutes.POST("/:id/test", func(c *gin.Context) {
			id := c.Param("id")
			userID, _ := c.Get("user_id")
			userRole, _ := c.Get("role")

			// 管理员可以测试所有连接，非管理员需要检查权限
			if userRole != "admin" {
				// 获取用户有权限的连接
				userConns := authManager.GetUserConnections(userID.(string), sshManager)
				// 检查用户是否有权限访问该连接
				hasPermission := false
				for _, userConn := range userConns {
					if userConn.ID == id {
						hasPermission = true
						break
					}
				}
				if !hasPermission {
					c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
					return
				}
			}

			if err := sshManager.TestConnection(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Connection test successful"})
		})

		// 测试未保存的连接 - 只有管理员可以测试
		connRoutes.POST("/test", authManager.RoleMiddleware("admin"), func(c *gin.Context) {
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
		// 创建连接 - 只有管理员可以创建
		oldSSHRoutes.POST("/connections", authManager.RoleMiddleware("admin"), func(c *gin.Context) {
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
		// 获取连接列表 - 根据用户角色返回相应连接
		oldSSHRoutes.GET("/connections", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			userRole, _ := c.Get("role")

			var conns []connection.Connection
			if userRole == "admin" {
				conns = sshManager.GetConnections()
			} else {
				conns = authManager.GetUserConnections(userID.(string), sshManager)
			}
			c.JSON(http.StatusOK, conns)
		})
		// 获取单个连接 - 检查用户是否有权限访问
		oldSSHRoutes.GET("/connections/:id", func(c *gin.Context) {
			id := c.Param("id")
			userID, _ := c.Get("user_id")
			userRole, _ := c.Get("role")

			// 检查连接是否存在
			conn, exists := sshManager.GetConnection(id)
			if !exists {
				c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
				return
			}

			// 管理员可以访问所有连接，非管理员需要检查权限
			if userRole != "admin" {
				// 获取用户有权限的连接
				userConns := authManager.GetUserConnections(userID.(string), sshManager)
				// 检查用户是否有权限访问该连接
				hasPermission := false
				for _, userConn := range userConns {
					if userConn.ID == id {
						hasPermission = true
						break
					}
				}
				if !hasPermission {
					c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
					return
				}
			}
			c.JSON(http.StatusOK, conn)
		})
		// 更新连接 - 只有管理员可以更新
		oldSSHRoutes.PUT("/connections/:id", authManager.RoleMiddleware("admin"), func(c *gin.Context) {
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
		// 删除连接 - 只有管理员可以删除
		oldSSHRoutes.DELETE("/connections/:id", authManager.RoleMiddleware("admin"), func(c *gin.Context) {
			id := c.Param("id")
			if err := sshManager.DeleteConnection(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			// 删除与该连接相关的所有权限
			if err := authManager.DeleteUserPermissionsByConnection(id); err != nil {
				log.Error("Failed to delete connection permissions: %v", err)
			}
			c.JSON(http.StatusOK, gin.H{"message": "Connection deleted"})
		})
		// 测试连接 - 检查用户是否有权限访问
		oldSSHRoutes.POST("/connections/:id/test", func(c *gin.Context) {
			id := c.Param("id")
			userID, _ := c.Get("user_id")
			userRole, _ := c.Get("role")

			// 管理员可以测试所有连接，非管理员需要检查权限
			if userRole != "admin" {
				// 获取用户有权限的连接
				userConns := authManager.GetUserConnections(userID.(string), sshManager)
				// 检查用户是否有权限访问该连接
				hasPermission := false
				for _, userConn := range userConns {
					if userConn.ID == id {
						hasPermission = true
						break
					}
				}
				if !hasPermission {
					c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
					return
				}
			}

			if err := sshManager.TestConnection(id); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Connection test successful"})
		})

		// 旧分类管理路由的重定向 - 只有管理员可以操作
		oldSSHRoutes.GET("/categories", func(c *gin.Context) {
			categories := sshManager.GetCategories()
			c.JSON(http.StatusOK, categories)
		})
		oldSSHRoutes.POST("/categories", authManager.RoleMiddleware("admin"), func(c *gin.Context) {
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
		oldSSHRoutes.PUT("/categories/:name", authManager.RoleMiddleware("admin"), func(c *gin.Context) {
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
		oldSSHRoutes.DELETE("/categories/:name", authManager.RoleMiddleware("admin"), func(c *gin.Context) {
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
							// 保存文件到默认下载目录
							if err := saveFileToDefaultDir(cfg, fileName, zipContent); err != nil {
								log.Error("Failed to save file to default directory: %s, error: %s", fileName, err.Error())
							}
							// 更新任务状态、文件名和下载路径
							taskManager.UpdateTaskFileName(zipTaskID, fileName)
							taskManager.UpdateTaskDownloadPath(zipTaskID, cfg.DefaultDownloadDir)
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
				// 保存文件到默认下载目录
				if err := saveFileToDefaultDir(cfg, fileName, content); err != nil {
					log.Error("Failed to save file to default directory: %s, error: %s", fileName, err.Error())
				}
				// 更新任务状态、文件名和下载路径
				taskManager.UpdateTaskFileName(taskID, fileName)
				taskManager.UpdateTaskDownloadPath(taskID, cfg.DefaultDownloadDir)
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
				// 保存文件到默认下载目录
				if err := saveFileToDefaultDir(cfg, zipFileName, zipContent); err != nil {
					log.Error("Failed to save file to default directory: %s, error: %s", zipFileName, err.Error())
				}
				// 更新任务状态、文件名和下载路径
				taskManager.UpdateTaskFileName(newTask.ID, zipFileName)
				taskManager.UpdateTaskDownloadPath(newTask.ID, cfg.DefaultDownloadDir)
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

	// 配置管理路由 - 公开路由
	publicConfigRoutes := r.Group("/api/config")
	{
		// 获取网站名称 - 公开访问
		publicConfigRoutes.GET("/site-name", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"siteName": cfg.SiteName,
			})
		})
	}

	// 配置管理路由 - 需要认证
	configRoutes := r.Group("/api/config")
	configRoutes.Use(authManager.AuthMiddleware())
	{
		// 获取默认下载目录
		configRoutes.GET("/download-dir", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"downloadDir": cfg.DefaultDownloadDir,
			})
		})

		// 设置默认下载目录
		configRoutes.POST("/download-dir", func(c *gin.Context) {
			var req struct {
				DownloadDir string `json:"downloadDir" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			// 更新配置
			cfg.DefaultDownloadDir = req.DownloadDir
			c.JSON(http.StatusOK, gin.H{
				"downloadDir": cfg.DefaultDownloadDir,
				"message":     "Download directory updated successfully",
			})
		})

		// 设置网站名称
		configRoutes.POST("/site-name", func(c *gin.Context) {
			var req struct {
				SiteName string `json:"siteName" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			// 更新配置
			cfg.SiteName = req.SiteName
			c.JSON(http.StatusOK, gin.H{
				"siteName": cfg.SiteName,
				"message":  "Site name updated successfully",
			})
		})
	}

	// 用户管理路由
	userRoutes := r.Group("/api/users")
	userRoutes.Use(authManager.AuthMiddleware(), authManager.RoleMiddleware("admin"))
	{
		// 获取所有用户
		userRoutes.GET("", func(c *gin.Context) {
			users := authManager.GetAllUsers()
			c.JSON(http.StatusOK, gin.H{"users": users})
		})

		// 创建用户
		userRoutes.POST("", func(c *gin.Context) {
			var userReq struct {
				Username string `json:"username" binding:"required"`
				Password string `json:"password" binding:"required,min=6"`
				Email    string `json:"email" binding:"required,email"`
				Role     string `json:"role"`
				Active   bool   `json:"active"`
			}
			if err := c.ShouldBindJSON(&userReq); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			// 创建用户对象
			user := &auth.User{
				ID:        authManager.GenerateID(),
				Username:  userReq.Username,
				Password:  authManager.HashPassword(userReq.Password),
				Email:     userReq.Email,
				Role:      userReq.Role,
				Active:    userReq.Active,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			// 设置默认值
			if user.Role == "" {
				user.Role = "user"
			}
			if user.Active == false {
				user.Active = true
			}

			// 保存用户
			if err := authManager.SaveUser(user); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, user)
		})

		// 获取单个用户
		userRoutes.GET("/:id", func(c *gin.Context) {
			userID := c.Param("id")
			users := authManager.GetAllUsers()
			for _, user := range users {
				if user.ID == userID {
					c.JSON(http.StatusOK, user)
					return
				}
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		})

		// 更新用户
		userRoutes.PUT("/:id", func(c *gin.Context) {
			userID := c.Param("id")
			var updatedUser auth.User
			if err := c.ShouldBindJSON(&updatedUser); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			if err := authManager.UpdateUser(userID, &updatedUser); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "User updated successfully"})
		})

		// 删除用户
		userRoutes.DELETE("/:id", func(c *gin.Context) {
			userID := c.Param("id")
			if err := authManager.DeleteUser(userID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
		})
	}

	// 个人中心路由
	profileRoutes := r.Group("/api/profile")
	profileRoutes.Use(authManager.AuthMiddleware())
	{
		// 获取当前用户信息
		profileRoutes.GET("", authManager.GetMe)

		// 更新当前用户信息
		profileRoutes.PUT("", func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			var updatedUser auth.User
			if err := c.ShouldBindJSON(&updatedUser); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			if err := authManager.UpdateUser(userID.(string), &updatedUser); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
		})
	}

	// 连接权限管理路由
	permissionRoutes := r.Group("/api/permissions")
	permissionRoutes.Use(authManager.AuthMiddleware(), authManager.RoleMiddleware("admin"))
	{
		// 获取用户的连接权限
		permissionRoutes.GET("/user/:userId", func(c *gin.Context) {
			userID := c.Param("userId")
			permissions := authManager.GetUserPermissions(userID)
			c.JSON(http.StatusOK, gin.H{"permissions": permissions})
		})

		// 授予用户连接权限
		permissionRoutes.POST("", func(c *gin.Context) {
			var permission auth.UserConnectionPermission
			if err := c.ShouldBindJSON(&permission); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
				return
			}

			if err := authManager.AddUserPermission(&permission); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusCreated, gin.H{"message": "Permission granted successfully", "permission": permission})
		})

		// 撤销用户连接权限
		permissionRoutes.DELETE("/:id", func(c *gin.Context) {
			permissionID := c.Param("id")
			if err := authManager.DeleteUserPermission(permissionID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"message": "Permission revoked successfully"})
		})
	}

	// 获取用户有权限的连接
	connRoutes.GET("/authorized", authManager.AuthMiddleware(), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		connections := authManager.GetUserConnections(userID.(string), sshManager)
		c.JSON(http.StatusOK, connections)
	})
}
