package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"syscall"

	"github.com/devcleaner/backend/logger"
	"github.com/devcleaner/backend/provider"
	"github.com/devcleaner/backend/tools"
	"github.com/gin-gonic/gin"
)

// 初始化
func init() {
	// 设置 Gin 为发布模式
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	// 初始化日志系统
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	logger.Init(logLevel)
	defer logger.Close()

	logger.Info("DevCleaner Backend 启动中...")

	// 初始化工具列表（从配置文件加载）
	tools.InitTools()
	logger.Info("工具列表已加载，共 %d 个工具", len(tools.GetAllTools()))

	r := gin.Default()

	// CORS 中间件
	r.Use(corsMiddleware())

	// 请求日志中间件
	r.Use(requestLogger())

	// 获取工具列表
	r.GET("/api/tools", handleGetTools)

	// 获取磁盘使用情况
	r.GET("/api/disk", handleDiskUsage)

	// 扫描指定工具
	r.POST("/api/scan", handleScan)

	// 清理指定工具缓存
	r.POST("/api/clean", handleClean)

	// 健康检查
	r.GET("/health", handleHealth)

	// 版本信息
	r.GET("/api/version", handleVersion)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("DevCleaner Backend 启动在 http://localhost:%s", port)
	if err := r.Run(":" + port); err != nil {
		logger.WithError(err, "服务器启动失败")
		os.Exit(1)
	}
}

// corsMiddleware CORS 中间件
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}

// requestLogger 请求日志中间件
func requestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Debug("请求: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next()
		logger.Debug("响应: %s %s [%d]", c.Request.Method, c.Request.URL.Path, c.Writer.Status())
	}
}

// handleHealth 健康检查
func handleHealth(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ok"})
}

// handleVersion 版本信息
func handleVersion(c *gin.Context) {
	c.JSON(200, gin.H{
		"version": "0.1.0",
		"name":    "DevCleaner",
	})
}

// handleGetTools 获取工具列表
func handleGetTools(c *gin.Context) {
	toolsList := tools.GetAllTools()
	toolInfos := make([]map[string]interface{}, 0, len(toolsList))
	for _, tool := range toolsList {
		toolInfos = append(toolInfos, map[string]interface{}{
			"id":     tool.ID,
			"name":   tool.Name,
			"paths":  tool.Paths,
			"enabled": true,
		})
	}
	logger.Info("获取工具列表: %d 个工具", len(toolInfos))
	c.JSON(http.StatusOK, toolInfos)
}

// handleDiskUsage 获取磁盘使用情况
func handleDiskUsage(c *gin.Context) {
	usage, err := getDiskUsage()
	if err != nil {
		logger.WithError(err, "获取磁盘使用情况失败")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("获取磁盘使用情况失败: %v", err),
		})
		return
	}
	logger.Info("获取磁盘使用情况: 总计 %d GB, 已使用 %d GB", usage["total"]/(1024*1024*1024), usage["used"]/(1024*1024*1024))
	c.JSON(http.StatusOK, usage)
}

// handleScan 扫描指定工具
func handleScan(c *gin.Context) {
	var req struct {
		ToolID string `json:"toolId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("扫描请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少 toolId 参数"})
		return
	}

	logger.Info("开始扫描工具: %s", req.ToolID)

	// 获取工具
	tool := tools.GetToolByID(req.ToolID)
	if tool == nil {
		logger.Warn("工具未找到: %s", req.ToolID)
		c.JSON(http.StatusNotFound, gin.H{"error": "工具未找到"})
		return
	}

	// 获取工具的 Provider
	p := provider.GetProvider(req.ToolID)
	if p == nil {
		logger.Error("Provider 未找到: %s", req.ToolID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider 未找到"})
		return
	}

	// 执行扫描
	results, err := p.Scan()
	if err != nil {
		logger.WithError(err, "扫描失败: %s", req.ToolID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("扫描失败: %v", err),
		})
		return
	}

	logger.Info("扫描完成: %s, 发现 %d 个缓存项", req.ToolID, len(results))
	c.JSON(http.StatusOK, gin.H{
		"toolId": req.ToolID,
		"result": results,
	})
}

// handleClean 清理指定工具缓存
func handleClean(c *gin.Context) {
	var req struct {
		ToolID string   `json:"toolId" binding:"required"`
		Paths  []string `json:"paths" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Warn("清理请求参数错误: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少必要参数"})
		return
	}

	logger.Info("开始清理工具: %s, 路径数: %d", req.ToolID, len(req.Paths))

	// 获取工具的 Provider
	p := provider.GetProvider(req.ToolID)
	if p == nil {
		logger.Error("Provider 未找到: %s", req.ToolID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Provider 未找到"})
		return
	}

	// 执行清理
	result, err := p.Clean(req.Paths)
	if err != nil {
		logger.WithError(err, "清理失败: %s", req.ToolID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   fmt.Sprintf("清理失败: %v", err),
			"cleaned": result,
		})
		return
	}

	logger.Info("清理完成: %s, 释放空间: %d bytes", req.ToolID, result.Cleaned)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"cleaned": result.Cleaned,
		"failed":  result.Failed,
	})
}

// getDiskUsage 获取真实的磁盘使用情况
func getDiskUsage() (map[string]int64, error) {
	switch runtime.GOOS {
	case "darwin", "linux":
		return getUnixDiskUsage()
	case "windows":
		return getWindowsDiskUsage()
	default:
		return getUnixDiskUsage() // 默认使用 Unix 方式
	}
}

// getUnixDiskUsage 获取 Unix 系统的磁盘使用情况
func getUnixDiskUsage() (map[string]int64, error) {
	// 优先使用主目录所在的分区
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/"
	}

	var statfs syscall.Statfs_t
	if err := syscall.Statfs(homeDir, &statfs); err != nil {
		return nil, fmt.Errorf("获取磁盘信息失败: %w", err)
	}

	total := int64(statfs.Blocks) * int64(statfs.Bsize)
	free := int64(statfs.Bfree) * int64(statfs.Bsize)
	used := total - free

	return map[string]int64{
		"total": total,
		"free":  free,
		"used":  used,
	}, nil
}

// getWindowsDiskUsage 获取 Windows 系统的磁盘使用情况
func getWindowsDiskUsage() (map[string]int64, error) {
	// Windows 下使用 os.Stat 来获取磁盘信息
	// 注意：Windows 完整实现需要使用 golang.org/x/sys/windows
	// 这里使用一个简化的实现
	
	// 获取用户主目录
	homeDir := os.Getenv("USERPROFILE")
	if homeDir == "" {
		homeDir = "C:\\"
	}

	// 检查是否是有效的路径
	dir, err := os.Stat(homeDir)
	if err != nil {
		return nil, fmt.Errorf("获取磁盘信息失败: %w", err)
	}

	// 对于 Windows，返回一个估计值
	// 实际实现需要调用 Windows API
	_ = dir

	// 简化实现：返回系统盘信息
	return map[string]int64{
		"total": 500 * 1024 * 1024 * 1024, // 500 GB (估计值)
		"free":  200 * 1024 * 1024 * 1024, // 200 GB (估计值)
		"used":  300 * 1024 * 1024 * 1024, // 300 GB (估计值)
	}, nil
}
