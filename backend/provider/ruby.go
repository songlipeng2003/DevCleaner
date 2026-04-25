package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// rubyProvider Ruby 提供商
type rubyProvider struct{}

func NewRubyProvider() Provider {
	return &rubyProvider{}
}

func (p *rubyProvider) ID() string   { return "ruby" }
func (p *rubyProvider) Name() string { return "Ruby" }

func (p *rubyProvider) Paths() []PathConfig {
	return []PathConfig{
		// gem 缓存
		{
			Path:        "~/.gem/cache",
			Description: "gem 本地缓存",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.gem/specifications",
			Description: "gem 规格缓存",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Caches/bundler",
			Description: "Bundler 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/bundler",
			Description: "Bundler 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~\\AppData\\Local\\bundler",
			Description: "Bundler 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// Ruby 缓存
		{
			Path:        "~/.ruby缓存",
			Description: "Ruby 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Caches/Ruby",
			Description: "Ruby 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		// rvm / rbenv 缓存
		{
			Path:        "~/.rvm/archives",
			Description: "RVM 归档",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.rbenv/versions/*/cache",
			Description: "rbenv 版本缓存",
			Strategy:    StrategyDirect,
		},
		// Rails 缓存
		{
			Path:        "**/tmp/cache",
			Description: "Rails 临时缓存",
			Strategy:    StrategySafe,
		},
		{
			Path:        "**/log/*.log",
			Description: "Rails 日志文件",
			Strategy:    StrategySafe,
		},
	}
}

func (p *rubyProvider) Scan() ([]ScanResult, error) {
	var results []ScanResult
	
	// 扫描标准路径
	paths := []struct {
		path        string
		description string
	}{
		{expandPath("~/.gem/cache"), "gem 本地缓存"},
		{expandPath("~/.gem/specifications"), "gem 规格缓存"},
		{expandPath("~/Library/Caches/bundler"), "Bundler 缓存"},
		{expandPath("~/.cache/bundler"), "Bundler 缓存 (Linux)"},
		{expandPath("~/Library/Caches/Ruby"), "Ruby 缓存"},
	}
	
	for _, p := range paths {
		if result, ok := p.scanSinglePath(p.path, p.description); ok {
			results = append(results, result)
		}
	}

	return results, nil
}

func (p *rubyProvider) scanSinglePath(path, description string) (ScanResult, bool) {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return ScanResult{}, false
	}
	if err != nil {
		return ScanResult{}, false
	}

	var totalSize int64
	var fileCount int
	var lastMod int64

	filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
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

func (p *rubyProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// 检查是否是安全的清理路径
		if p.isSafeToClean(path) {
			cleaned, failed := cleanPathDirect(path)
			result.Cleaned += cleaned
			result.Failed = append(result.Failed, failed...)
		} else {
			result.Failed = append(result.Failed, path+": 需要用户手动确认清理")
		}
	}

	// 运行 gem cleanup
	p.cleanGem()

	return result, nil
}

func (p *rubyProvider) isSafeToClean(path string) bool {
	unsafePatterns := []string{
		"tmp/cache",
		"log/",
		"rvm/archives",
	}
	
	for _, pattern := range unsafePatterns {
		if strings.Contains(path, pattern) {
			return false
		}
	}
	return true
}

func (p *rubyProvider) cleanGem() {
	// 清理未使用的 gem 版本
	cmd := exec.Command("gem", "cleanup")
	cmd.Run() // 忽略错误
	
	// 清理 bundler 缓存
	cmd = exec.Command("bundle", "clean", "--force")
	cmd.Run()
}

// GetRubyInfo 获取 Ruby 环境信息
func GetRubyInfo() (map[string]string, error) {
	info := make(map[string]string)
	
	// Ruby 版本
	cmd := exec.Command("ruby", "--version")
	if output, err := cmd.Output(); err == nil {
		info["ruby_version"] = strings.TrimSpace(string(output))
	}
	
	// gem 版本
	cmd = exec.Command("gem", "--version")
	if output, err := cmd.Output(); err == nil {
		info["gem_version"] = strings.TrimSpace(string(output))
	}
	
	// Bundler 版本
	cmd = exec.Command("bundle", "--version")
	if output, err := cmd.Output(); err == nil {
		info["bundler_version"] = strings.TrimSpace(string(output))
	}
	
	return info, nil
}
