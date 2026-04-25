package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// pnpmProvider pnpm 包管理器缓存清理
type pnpmProvider struct{}

func NewPnpmProvider() Provider {
	return &pnpmProvider{}
}

func (p *pnpmProvider) ID() string   { return "pnpm" }
func (p *pnpmProvider) Name() string { return "pnpm" }

func (p *pnpmProvider) Paths() []PathConfig {
	home, _ := os.UserHomeDir()

	paths := []PathConfig{
		// pnpm store 路径（跨平台）
		{
			Path:        "~/.pnpm-store",
			Description: "pnpm 全局 store",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.local/share/pnpm/store",
			Description: "pnpm store (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/pnpm/store",
			Description: "pnpm store (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\pnpm\\store",
			Description: "pnpm store (Windows)",
			Strategy:    StrategyDirect,
		},
		// pnpm 缓存目录
		{
			Path:        "~/.pnpm-cache",
			Description: "pnpm 缓存",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/pnpm",
			Description: "pnpm 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
	}

	// 根据操作系统添加特定路径
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths,
			PathConfig{
				Path:        home + "/Library/Caches/pnpm",
				Description: "pnpm 缓存 (macOS)",
				Strategy:    StrategyDirect,
			},
		)
	case "linux":
		paths = append(paths,
			PathConfig{
				Path:        home + "/.cache/nodejs/pnpm",
				Description: "pnpm 缓存 (Linux via nodejs)",
				Strategy:    StrategyDirect,
			},
		)
	case "windows":
		paths = append(paths,
			PathConfig{
				Path:        "%APPDATA%\\pnpm\\cache",
				Description: "pnpm 缓存 (Windows)",
				Strategy:    StrategyDirect,
			},
			PathConfig{
				Path:        "%TEMP%\\pnpm",
				Description: "pnpm 临时文件 (Windows)",
				Strategy:    StrategyDirect,
			},
		)
	}

	return paths
}

func (p *pnpmProvider) Scan() ([]ScanResult, error) {
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

		filepath.Walk(expandedPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() {
				totalSize += info.Size()
				fileCount++
				if mod := info.ModTime().Unix(); mod > lastMod {
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

func (p *pnpmProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	// 尝试使用 pnpm 命令清理
	cmd := exec.Command("pnpm", "store", "prune")
	if output, err := cmd.CombinedOutput(); err != nil {
		result.Failed = append(result.Failed, string(output))
	}

	// 清理所有缓存
	cmd = exec.Command("pnpm", "store", "clear")
	if output, err := cmd.CombinedOutput(); err != nil {
		result.Failed = append(result.Failed, string(output))
	}

	// 直接清理各个路径
	for _, path := range paths {
		cleaned, failed := cleanPathDirect(path)
		result.Cleaned += cleaned
		result.Failed = append(result.Failed, failed...)
	}

	return result, nil
}

func isPnpmCache(path string) bool {
	return strings.Contains(path, "pnpm")
}
