package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// goProvider Go 提供商
type goProvider struct{}

func NewGoProvider() Provider {
	return &goProvider{}
}

func (p *goProvider) ID() string   { return "go" }
func (p *goProvider) Name() string { return "Go" }

func (p *goProvider) Paths() []PathConfig {
	return []PathConfig{
		// Go 模块缓存
		{
			Path:        "$(go env GOPATH)/pkg/mod",
			Description: "Go 模块缓存",
			Strategy:    StrategyDirect,
		},
		// Go 构建缓存
		{
			Path:        "$(go env GOCACHE)",
			Description: "Go 构建缓存",
			Strategy:    StrategyDirect,
		},
		// Go 测试缓存
		{
			Path:        "$(go env GOCACHE)/test",
			Description: "Go 测试缓存",
			Strategy:    StrategyDirect,
		},
		// go.sum 文件（可选清理）
		{
			Path:        "**/go.sum",
			Description: "go.sum 依赖锁定文件",
			Strategy:    StrategySafe,
		},
	}
}

func (p *goProvider) Scan() ([]ScanResult, error) {
	var results []ScanResult
	
	// 获取 Go 环境变量
	gopath := p.getGoEnv("GOPATH")
	gocache := p.getGoEnv("GOCACHE")
	
	paths := []struct {
		path        string
		description string
	}{
		{filepath.Join(gopath, "pkg", "mod"), "Go 模块缓存"},
		{gocache, "Go 构建缓存"},
	}
	
	for _, pathItem := range paths {
		if result, ok := p.scanSinglePath(pathItem.path, pathItem.description); ok {
			results = append(results, result)
		}
	}

	return results, nil
}

func (p *goProvider) scanSinglePath(path, description string) (ScanResult, bool) {
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

func (p *goProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// Go 模块缓存可以安全清理
		if strings.Contains(path, "pkg/mod") || strings.Contains(path, "GOCACHE") {
			cleaned, failed := cleanPathDirect(path)
			result.Cleaned += cleaned
			result.Failed = append(result.Failed, failed...)
		} else {
			result.Failed = append(result.Failed, path+": 不建议自动清理")
		}
	}

	// 运行 go clean -modcache 清理模块缓存
	p.cleanModCache()

	return result, nil
}

func (p *goProvider) cleanModCache() {
	cmd := exec.Command("go", "clean", "-modcache")
	cmd.Run() // 忽略错误
}

func (p *goProvider) getGoEnv(key string) string {
	cmd := exec.Command("go", "env", key)
	output, err := cmd.Output()
	if err != nil {
		// 返回默认值
		switch key {
		case "GOPATH":
			if runtime.GOOS == "windows" {
				return filepath.Join(os.Getenv("USERPROFILE"), "go")
			}
			return filepath.Join(os.Getenv("HOME"), "go")
		case "GOCACHE":
			if runtime.GOOS == "windows" {
				return filepath.Join(os.Getenv("LOCALAPPDATA"), "go", "build")
			}
			return filepath.Join(os.Getenv("HOME"), ".cache", "go-build")
		}
		return ""
	}
	return strings.TrimSpace(string(output))
}

// GetGoInfo 获取 Go 环境信息
func GetGoInfo() (map[string]string, error) {
	info := make(map[string]string)
	
	// Go 版本
	cmd := exec.Command("go", "version")
	if output, err := cmd.Output(); err == nil {
		info["go_version"] = strings.TrimSpace(string(output))
	}
	
	// GOPATH
	cmd = exec.Command("go", "env", "GOPATH")
	if output, err := cmd.Output(); err == nil {
		info["gopath"] = strings.TrimSpace(string(output))
	}
	
	// GOROOT
	cmd = exec.Command("go", "env", "GOROOT")
	if output, err := cmd.Output(); err == nil {
		info["goroot"] = strings.TrimSpace(string(output))
	}
	
	return info, nil
}
