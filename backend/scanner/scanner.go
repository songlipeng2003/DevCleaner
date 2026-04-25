package scanner

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ScanResult 扫描结果
type ScanResult struct {
	ToolID      string `json:"tool_id"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	FileNum     int    `json:"file_num"`
	LastMod     int64  `json:"last_modified"`
	Description string `json:"description,omitempty"`
}

// Scanner 扫描引擎
type Scanner struct {
	results   chan ScanResult
	wg        sync.WaitGroup
	cancel    chan struct{}
	whitelist []string
}

// NewScanner 创建扫描器
func NewScanner(whitelist []string) *Scanner {
	return &Scanner{
		results:   make(chan ScanResult, 100),
		cancel:   make(chan struct{}),
		whitelist: whitelist,
	}
}

// Cancel 取消扫描
func (s *Scanner) Cancel() {
	close(s.cancel)
}

// ScanPath 扫描指定路径
func (s *Scanner) ScanPath(toolID, path, description string) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		select {
		case <-s.cancel:
			return
		default:
		}

		// 展开路径
		expandedPath := expandPath(path)
		
		// 检查路径是否存在
		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			return
		}

		var totalSize int64
		var fileCount int
		var lastMod int64

		filepath.Walk(expandedPath, func(walkPath string, info os.FileInfo, err error) error {
			select {
			case <-s.cancel:
				return filepath.SkipAll
			default:
			}

			if err != nil {
				return nil
			}

			// 检查是否在白名单中
			if isWhitelisted(walkPath, s.whitelist) {
				return nil
			}

			if !info.IsDir() {
				totalSize += info.Size()
				fileCount++
				modTime := info.ModTime().Unix()
				if modTime > lastMod {
					lastMod = modTime
				}
			}
			return nil
		})

		if totalSize > 0 {
			s.results <- ScanResult{
				ToolID:      toolID,
				Path:        expandedPath,
				Size:        totalSize,
				FileNum:     fileCount,
				LastMod:     lastMod,
				Description: description,
			}
		}
	}()
}

// ScanTool 扫描指定工具的所有路径
func (s *Scanner) ScanTool(toolID string, paths []PathInfo) {
	for _, p := range paths {
		s.ScanPath(toolID, p.Path, p.Description)
	}
}

// CollectResults 收集扫描结果
func (s *Scanner) CollectResults() []ScanResult {
	go func() {
		s.wg.Wait()
		close(s.results)
	}()

	var results []ScanResult
	for r := range s.results {
		results = append(results, r)
	}
	return results
}

// PathInfo 路径信息
type PathInfo struct {
	Path        string
	Description string
}

// expandPath 展开路径中的特殊字符
func expandPath(path string) string {
	usr, _ := user.Current()
	home := usr.HomeDir
	
	path = strings.ReplaceAll(path, "~", home)
	path = os.ExpandEnv(path)
	
	return path
}

// isWhitelisted 检查路径是否在白名单中
func isWhitelisted(path string, whitelist []string) bool {
	for _, w := range whitelist {
		if strings.HasPrefix(path, w) {
			return true
		}
	}
	return false
}

// QuickScan 快速扫描（不遍历子目录）
func QuickScan(path string) (int64, error) {
	expandedPath := expandPath(path)
	
	info, err := os.Stat(expandedPath)
	if err != nil {
		return 0, err
	}

	if !info.IsDir() {
		return info.Size(), nil
	}

	var totalSize int64
	entries, err := os.ReadDir(expandedPath)
	if err != nil {
		return 0, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		totalSize += info.Size()
	}

	return totalSize, nil
}

// ScanStats 扫描统计
type ScanStats struct {
	TotalSize    int64
	TotalFiles   int
	ScanDuration time.Duration
	ScannedPaths int
}

// FormatSize 格式化大小
func FormatSize(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	}
	const unit = 1024
	sizes := []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for bytes >= unit && i < len(sizes)-1 {
		bytes /= unit
		i++
	}
	return string(rune('0'+byte(bytes%10))) + " " + sizes[i]
}
