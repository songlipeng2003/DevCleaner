package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// pythonProvider Python 提供商
type pythonProvider struct{}

func NewPythonProvider() Provider {
	return &pythonProvider{}
}

func (p *pythonProvider) ID() string   { return "python" }
func (p *pythonProvider) Name() string { return "Python" }

func (p *pythonProvider) Paths() []PathConfig {
	return []PathConfig{
		// pip 缓存
		{
			Path:        "~/.cache/pip",
			Description: "pip 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Caches/pip",
			Description: "pip 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\pip\\cache",
			Description: "pip 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// Python 缓存
		{
			Path:        "~/.pycache",
			Description: "Python __pycache__",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "**/__pycache__",
			Description: "项目级 Python 缓存",
			Strategy:    StrategyDirect,
		},
		// venv 缓存
		{
			Path:        "~/.local/share/virtualenvs",
			Description: "virtualenvwrapper 虚拟环境",
			Strategy:    StrategySafe,
		},
		{
			Path:        "~/venv",
			Description: "项目 venv 目录",
			Strategy:    StrategySafe,
		},
		// Conda 缓存
		{
			Path:        "~/conda_pkgs_dir",
			Description: "Conda 下载包缓存",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.conda",
			Description: "Conda 目录 (Linux/macOS)",
			Strategy:    StrategySafe,
		},
		{
			Path:        "C:\\Users\\.conda",
			Description: "Conda 目录 (Windows)",
			Strategy:    StrategySafe,
		},
		// Poetry 缓存
		{
			Path:        "~/.cache/pypoetry",
			Description: "Poetry 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Caches/pypoetry",
			Description: "Poetry 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
	}
}

func (p *pythonProvider) Scan() ([]ScanResult, error) {
	var results []ScanResult
	
	for _, pathConfig := range p.Paths() {
		// 处理通配符路径
		paths := p.expandWildcardPath(pathConfig.Path)
		
		for _, path := range paths {
			if result, ok := p.scanSinglePath(path, pathConfig.Description); ok {
				results = append(results, result)
			}
		}
	}

	return results, nil
}

func (p *pythonProvider) scanSinglePath(path, description string) (ScanResult, bool) {
	expandedPath := expandPath(path)
	
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return ScanResult{}, false
	} else if err != nil {
		return ScanResult{}, false
	}

	var totalSize int64
	var fileCount int
	var lastMod int64

	filepath.Walk(expandedPath, func(walkPath string, fileInfo os.FileInfo, err error) error {
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

func (p *pythonProvider) expandWildcardPath(path string) []string {
	if strings.Contains(path, "**") {
		// 返回基本路径，实际扫描时递归处理
		base := strings.TrimSuffix(path, "/**")
		return []string{expandPath(base)}
	}
	return []string{path}
}

func (p *pythonProvider) Clean(paths []string) (*CleanResult, error) {
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

	// 清理 __pycache__ 目录
	p.cleanPycache()

	return result, nil
}

func (p *pythonProvider) isSafeToClean(path string) bool {
	unsafePatterns := []string{
		"venv",
		".venv",
		"virtualenvs",
		".conda",
		"conda_pkgs_dir",
	}
	
	for _, pattern := range unsafePatterns {
		if strings.Contains(path, pattern) {
			return false
		}
	}
	return true
}

func (p *pythonProvider) cleanPycache() {
	// 清理全局 __pycache__
	home, _ := os.UserHomeDir()
	paths := []string{
		filepath.Join(home, ".pycache"),
	}
	
	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			cleanPathDirect(path)
		}
	}
}

// GetPythonInfo 获取 Python 环境信息
func GetPythonInfo() (map[string]string, error) {
	info := make(map[string]string)
	
	// Python 版本
	cmd := exec.Command("python3", "--version")
	if output, err := cmd.Output(); err == nil {
		info["python_version"] = strings.TrimSpace(string(output))
	}
	
	// pip 版本
	cmd = exec.Command("pip3", "--version")
	if output, err := cmd.Output(); err == nil {
		info["pip_version"] = strings.TrimSpace(string(output))
	}
	
	// 虚拟环境目录
	cmd = exec.Command("python3", "-c", "import site; print(site.getusersitepackages())")
	if output, err := cmd.Output(); err == nil {
		info["user_site"] = strings.TrimSpace(string(output))
	}
	
	return info, nil
}
