package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// composerProvider Composer 提供商
type composerProvider struct{}

func NewComposerProvider() Provider {
	return &composerProvider{}
}

func (p *composerProvider) ID() string   { return "composer" }
func (p *composerProvider) Name() string { return "Composer" }

func (p *composerProvider) Paths() []PathConfig {
	return []PathConfig{
		// Composer 全局缓存
		{
			Path:        "~/.cache/composer",
			Description: "Composer 全局缓存 (Linux/macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Caches/composer",
			Description: "Composer 全局缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%APPDATA%\\Composer",
			Description: "Composer 全局缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// Composer 本地缓存（旧版本）
		{
			Path:        "~/.composer/cache",
			Description: "Composer 本地缓存（旧版本）",
			Strategy:    StrategyDirect,
		},
		// Composer vendor 目录（项目依赖）
		{
			Path:        "**/vendor",
			Description: "Composer vendor 目录（项目依赖）",
			Strategy:    StrategySafe, // 需谨慎清理
		},
		// Composer 锁文件
		{
			Path:        "**/composer.lock",
			Description: "Composer 锁文件",
			Strategy:    StrategySafe,
		},
	}
}

func (p *composerProvider) Scan() ([]ScanResult, error) {
	var results []ScanResult

	// 扫描标准路径
	paths := []struct {
		path        string
		description string
	}{
		{expandPath("~/.cache/composer"), "Composer 全局缓存"},
		{expandPath("~/Library/Caches/composer"), "Composer 全局缓存 (macOS)"},
		{expandPath("~/.composer/cache"), "Composer 本地缓存（旧版本）"},
	}

	for _, pathItem := range paths {
		if result, ok := p.scanSinglePath(pathItem.path, pathItem.description); ok {
			results = append(results, result)
		}
	}

	return results, nil
}

func (p *composerProvider) scanSinglePath(path, description string) (ScanResult, bool) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ScanResult{}, false
	} else if err != nil {
		return ScanResult{}, false
	}

	var totalSize int64
	var fileCount int
	var lastMod int64

	filepath.Walk(path, func(walkPath string, fileInfo os.FileInfo, err error) error {
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
			Path:        path,
			Size:        totalSize,
			FileNum:     fileCount,
			LastMod:     lastMod,
			Description: description,
		}, true
	}

	return ScanResult{}, false
}

func (p *composerProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// 检查是否是 vendor 目录，需谨慎
		if strings.Contains(path, "vendor") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 vendor 目录")
			continue
		}
		// 检查是否是 composer.lock 文件
		if strings.HasSuffix(path, "composer.lock") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 composer.lock 文件")
			continue
		}
		// 其他缓存可以安全清理
		cleaned, failed := cleanPathDirect(path)
		result.Cleaned += cleaned
		result.Failed = append(result.Failed, failed...)
	}

	// 运行 composer clear-cache 命令
	p.clearCache()

	return result, nil
}

func (p *composerProvider) clearCache() {
	// 尝试使用 composer clear-cache 命令
	cmd := exec.Command("composer", "clear-cache")
	cmd.Run() // 忽略错误
}

// GetComposerInfo 获取 Composer 环境信息
func GetComposerInfo() (map[string]string, error) {
	info := make(map[string]string)

	// Composer 版本
	cmd := exec.Command("composer", "--version")
	if output, err := cmd.Output(); err == nil {
		info["composer_version"] = strings.TrimSpace(string(output))
	}

	// 全局配置路径
	cmd = exec.Command("composer", "config", "--global", "home")
	if output, err := cmd.Output(); err == nil {
		info["composer_home"] = strings.TrimSpace(string(output))
	}

	return info, nil
}