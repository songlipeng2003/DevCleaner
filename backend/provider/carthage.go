package provider

import (
	"os"
	"path/filepath"
)

// carthageProvider Carthage 提供商
type carthageProvider struct{}

func NewCarthageProvider() Provider {
	return &carthageProvider{}
}

func (p *carthageProvider) ID() string   { return "carthage" }
func (p *carthageProvider) Name() string { return "Carthage" }

func (p *carthageProvider) Paths() []PathConfig {
	return []PathConfig{
		{
			Path:        "~/Library/Caches Carthage",
			Description: "Carthage 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Carthage",
			Description: "Carthage 构建目录",
			Strategy:    StrategySafe, // 安全删除
		},
		{
			Path:        "~/.carthage",
			Description: "Carthage 配置目录",
			Strategy:    StrategyDirect,
		},
	}
}

func (p *carthageProvider) Scan() ([]ScanResult, error) {
	var results []ScanResult

	for _, pathConfig := range p.Paths() {
		expandedPath := expandPath(pathConfig.Path)

		if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
			continue
		} else if err != nil {
			continue
		}

		var totalSize int64
		var fileCount int
		var lastMod int64

		filepath.Walk(expandedPath, func(path string, fileInfo os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !fileInfo.IsDir() {
				totalSize += fileInfo.Size()
				fileCount++
				if mod := fileInfo.ModTime().Unix(); mod > lastMod {
					lastMod = mod
				}
			}
			return nil
		})

		if totalSize > 0 {
			results = append(results, ScanResult{
				Path:        expandedPath,
				Size:        totalSize,
				FileNum:     fileCount,
				LastMod:     lastMod,
				Description: pathConfig.Description,
			})
		}
	}

	return results, nil
}

func (p *carthageProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// Carthage 主要缓存是 Build 目录
		if isCarthageBuildDir(path) {
			// 只清理 Build 目录中的缓存
			cleaned, failed := p.cleanBuildCache(path)
			result.Cleaned += cleaned
			result.Failed = append(result.Failed, failed...)
		} else {
			cleaned, failed := cleanPathDirect(path)
			result.Cleaned += cleaned
			result.Failed = append(result.Failed, failed...)
		}
	}

	return result, nil
}

func (p *carthageProvider) cleanBuildCache(buildDir string) (int64, []string) {
	var totalSize int64
	var failed []string

	// 只清理 Build 目录，不删除整个 Carthage 目录
	buildPath := filepath.Join(buildDir, "Build")

	if _, err := os.Stat(buildPath); os.IsNotExist(err) {
		return 0, nil
	}

	filepath.Walk(buildPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			failed = append(failed, path)
			return nil
		}

		if !info.IsDir() {
			if err := os.Remove(path); err != nil {
				failed = append(failed, path)
			} else {
				totalSize += info.Size()
			}
		}
		return nil
	})

	// 删除空的 Build 子目录
	filepath.Walk(buildPath, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			os.Remove(path)
		}
		return nil
	})

	return totalSize, failed
}

func isCarthageBuildDir(path string) bool {
	return filepath.Base(path) == "Carthage" ||
		filepath.Base(path) == ".carthage" ||
		filepath.Base(path) == "Caches Carthage"
}
