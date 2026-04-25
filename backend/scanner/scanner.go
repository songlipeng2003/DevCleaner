package scanner

import (
	"os"
	"path/filepath"
	"sync"
)

// ScanResult 扫描结果
type ScanResult struct {
	ToolID   string `json:"tool_id"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	FileNum  int    `json:"file_num"`
	LastMod  int64  `json:"last_modified"`
}

// Scanner 扫描引擎
type Scanner struct {
	results chan ScanResult
	wg      sync.WaitGroup
}

// NewScanner 创建扫描器
func NewScanner() *Scanner {
	return &Scanner{
		results: make(chan ScanResult, 100),
	}
}

// ScanPath 扫描指定路径
func (s *Scanner) ScanPath(toolID, path string) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		var totalSize int64
		var fileCount int
		var lastMod int64

		filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				totalSize += info.Size()
				fileCount++
				if info.ModTime().UnixNano() > lastMod {
					lastMod = info.ModTime().UnixNano()
				}
			}
			return nil
		})

		s.results <- ScanResult{
			ToolID:  toolID,
			Path:    path,
			Size:    totalSize,
			FileNum: fileCount,
			LastMod: lastMod,
		}
	}()
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
