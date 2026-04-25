package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// homebrewProvider Homebrew 提供商
type homebrewProvider struct{}

func NewHomebrewProvider() Provider {
	return &homebrewProvider{}
}

func (p *homebrewProvider) ID() string   { return "homebrew" }
func (p *homebrewProvider) Name() string { return "Homebrew" }

func (p *homebrewProvider) Paths() []PathConfig {
	var paths []PathConfig
	
	// 根据系统确定路径
	if runtime.GOOS == "darwin" {
		paths = []PathConfig{
			{
				Path:        "$(brew --cache)",
				Description: "Homebrew 下载缓存",
				Strategy:    StrategyDirect,
			},
			{
				Path:        "~/Library/Caches/Homebrew",
				Description: "Homebrew 缓存（新版）",
				Strategy:    StrategyDirect,
			},
			{
				Path:        "/usr/local/Cellar",
				Description: "Homebrew Cellar (Intel)",
				Strategy:    StrategyCommand,
			},
			{
				Path:        "/opt/homebrew/Cellar",
				Description: "Homebrew Cellar (Apple Silicon)",
				Strategy:    StrategyCommand,
			},
			{
				Path:        "/usr/local/bin",
				Description: "Homebrew bin (Intel)",
				Strategy:    StrategySafe,
			},
			{
				Path:        "/opt/homebrew/bin",
				Description: "Homebrew bin (Apple Silicon)",
				Strategy:    StrategySafe,
			},
		}
	} else if runtime.GOOS == "linux" {
		paths = []PathConfig{
			{
				Path:        "$(brew --cache)",
				Description: "Homebrew 下载缓存",
				Strategy:    StrategyDirect,
			},
			{
				Path:        "/home/linuxbrew/.linuxbrew/Cellar",
				Description: "Homebrew Cellar",
				Strategy:    StrategyCommand,
			},
		}
	}

	return paths
}

func (p *homebrewProvider) Scan() ([]ScanResult, error) {
	var results []ScanResult
	
	// 获取 brew cache 路径
	cachePath := p.getBrewCache()
	if cachePath != "" {
		if size, ok := p.scanPath(cachePath, "Homebrew 下载缓存"); ok {
			results = append(results, size)
		}
	}

	// 扫描其他路径
	for _, pathConfig := range p.Paths() {
		if pathConfig.Path == "$(brew --cache)" {
			continue // 已处理
		}
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

func (p *homebrewProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// 缓存目录直接删除
		if strings.Contains(path, "cache") || strings.Contains(path, "Cache") {
			cleaned, failed := cleanPathDirect(path)
			result.Cleaned += cleaned
			result.Failed = append(result.Failed, failed...)
			continue
		}

		// Cellar 使用 brew uninstall 清理
		if strings.Contains(path, "Cellar") {
			cleaned, failed := p.cleanCellar(path)
			result.Cleaned += cleaned
			result.Failed = append(result.Failed, failed...)
			continue
		}

		// bin 目录不清理
		result.Failed = append(result.Failed, path+": bin 目录不清理")
	}

	return result, nil
}

// getBrewCache 获取 brew cache 路径
func (p *homebrewProvider) getBrewCache() string {
	cmd := exec.Command("brew", "--cache")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// scanPath 扫描单个路径
func (p *homebrewProvider) scanPath(path, description string) (ScanResult, bool) {
	expandedPath := expandPath(path)
	
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return ScanResult{}, false
	} else if err != nil {
		return ScanResult{}, false
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
		return ScanResult{
			Path:        expandedPath,
			Size:        totalSize,
			FileNum:     fileCount,
			LastMod:     lastMod,
			Description: description,
		}, true
	}

	return ScanResult{}, false
}

// cleanCellar 清理 Cellar（使用 brew cleanup）
func (p *homebrewProvider) cleanCellar(path string) (int64, []string) {
	var failed []string
	var cleaned int64

	// 先获取清理前的缓存大小
	cachePath := p.getBrewCache()
	if cachePath != "" {
		info, err := os.Stat(cachePath)
		if err == nil {
			cleaned = info.Size()
		}
	}

	// 运行 brew cleanup
	cmd := exec.Command("brew", "cleanup", "--prune=all")
	output, err := cmd.CombinedOutput()
	if err != nil {
		failed = append(failed, string(output))
	}

	return cleaned, failed
}

// GetHomebrewInfo 获取 Homebrew 安装信息
func GetHomebrewInfo() (string, error) {
	cmd := exec.Command("brew", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
