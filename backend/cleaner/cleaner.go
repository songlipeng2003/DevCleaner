package cleaner

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CleanResult 清理结果
type CleanResult struct {
	ToolID   string   `json:"tool_id"`
	Cleaned  int64    `json:"cleaned"`
	Failed   []string `json:"failed"`
	FileNum  int      `json:"file_num"`
}

// Cleaner 清理引擎
type Cleaner struct {
	whitelist []string
}

// NewCleaner 创建清理器
func NewCleaner(whitelist []string) *Cleaner {
	return &Cleaner{
		whitelist: whitelist,
	}
}

// Clean 清理指定路径
func (c *Cleaner) Clean(toolID, path string) (*CleanResult, error) {
	result := &CleanResult{
		ToolID:  toolID,
		Failed:  []string{},
	}

	// 检查白名单
	for _, w := range c.whitelist {
		if strings.HasPrefix(path, w) {
			return nil, fmt.Errorf("path %s is in whitelist", path)
		}
	}

	// 统计要删除的文件
	var totalSize int64
	var fileCount int

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			result.Failed = append(result.Failed, fmt.Sprintf("access error: %s", path))
			return nil
		}

		// 检查是否在白名单中
		for _, w := range c.whitelist {
			if strings.HasPrefix(path, w) {
				return nil
			}
		}

		if !info.IsDir() {
			if err := os.Remove(path); err != nil {
				result.Failed = append(result.Failed, fmt.Sprintf("delete failed: %s", path))
			} else {
				totalSize += info.Size()
				fileCount++
			}
		}
		return nil
	})

	// 删除空目录
	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			os.Remove(path) // 忽略错误，继续清理
		}
		return nil
	})

	result.Cleaned = totalSize
	result.FileNum = fileCount
	return result, nil
}
