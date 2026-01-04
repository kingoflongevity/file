package middleware

import (
	"context"
	"net/http"
	"sync"
	"time"

	"remote-file-manager/internal/log"

	"github.com/gin-gonic/gin"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	visitors map[string][2]int64 // IP -> [count, lastResetTime]
	mutex    sync.Mutex
	rate     int           // 允许的请求数量
	period   time.Duration // 时间窗口
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(rate int, period time.Duration) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string][2]int64),
		rate:     rate,
		period:   period,
	}
}

// Limit 速率限制中间件
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now().Unix()

		rl.mutex.Lock()
		defer rl.mutex.Unlock()

		// 获取或初始化IP记录
		record, exists := rl.visitors[ip]
		if !exists {
			// 新IP，初始化计数为1，重置时间为当前时间
			rl.visitors[ip] = [2]int64{1, now}
			c.Next()
			return
		}

		// 检查时间窗口是否已过期
		if now-record[1] > int64(rl.period.Seconds()) {
			// 时间窗口过期，重置计数和时间
			rl.visitors[ip] = [2]int64{1, now}
			c.Next()
			return
		}

		// 检查请求数量是否超过限制
		if record[0] >= int64(rl.rate) {
			// 请求过多，返回429状态码
			log.Warn("Rate limit exceeded for IP: %s", ip)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "请求频率过高，请稍后重试",
				"retry_after": rl.period.Seconds(),
			})
			c.Abort()
			return
		}

		// 请求数量加1
		rl.visitors[ip] = [2]int64{record[0] + 1, record[1]}
		c.Next()
	}
}

// Timeout 超时控制中间件
func Timeout(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建超时上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 替换请求上下文
		c.Request = c.Request.WithContext(ctx)

		// 使用done channel来通知请求完成
		done := make(chan struct{})

		// 使用goroutine执行请求处理
		go func() {
			c.Next()
			done <- struct{}{}
		}()

		// 等待请求完成或超时
		select {
		case <-done:
			// 请求正常完成
			return
		case <-ctx.Done():
			// 请求超时
			log.Warn("Request timed out: %s %s", c.Request.Method, c.Request.URL.Path)
			c.AbortWithStatusJSON(http.StatusRequestTimeout, gin.H{
				"error": "请求超时",
			})
		}
	}
}

// RequestLogger 请求日志中间件
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()

		// 请求路径
		path := c.Request.URL.Path
		method := c.Request.Method
		ip := c.ClientIP()

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)

		// 状态码
		status := c.Writer.Status()

		// 记录日志
		log.Info("%s %s %d %v %s", method, path, status, latency, ip)
	}
}

// Recovery 恢复中间件
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录错误日志
				log.Error("Panic recovered: %v", err)

				// 返回500状态码
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "服务器内部错误",
				})
			}
		}()

		c.Next()
	}
}

// CORSMiddleware CORS中间件
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
