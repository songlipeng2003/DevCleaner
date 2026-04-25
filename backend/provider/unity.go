package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// unityProvider Unity 提供商
type unityProvider struct{}

func NewUnityProvider() Provider {
	return &unityProvider{}
}

func (p *unityProvider) ID() string   { return "unity" }
func (p *unityProvider) Name() string { return "Unity" }

func (p *unityProvider) Paths() []PathConfig {
	return []PathConfig{
		// macOS 路径
		{
			Path:        "~/Library/Unity/Cache",
			Description: "Unity 编辑器缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Caches/Unity",
			Description: "Unity 下载和构建缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Logs/Unity",
			Description: "Unity 日志文件 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Application Support/Unity",
			Description: "Unity 应用数据 (macOS)",
			Strategy:    StrategySafe,
		},
		// Windows 路径
		{
			Path:        "%USERPROFILE%\\AppData\\Local\\Unity",
			Description: "Unity 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%USERPROFILE%\\AppData\\LocalLow\\Unity",
			Description: "Unity 低完整性缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// Linux 路径
		{
			Path:        "~/.config/unity3d",
			Description: "Unity 配置目录 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/Unity",
			Description: "Unity 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.local/share/unity3d",
			Description: "Unity 数据目录 (Linux)",
			Strategy:    StrategySafe,
		},
	}
}

func (p *unityProvider) Scan() ([]ScanResult, error) {
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

func (p *unityProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// Unity 缓存可以安全删除
		if isUnitySafeToClean(path) {
			cleaned, failed := cleanPathDirect(path)
			result.Cleaned += cleaned
			result.Failed = append(result.Failed, failed...)
		} else {
			// 需要用户确认的路径
			result.Failed = append(result.Failed, path+": 需要用户手动确认清理")
		}
	}

	// 清理 Unity 日志文件（如果存在）
	p.cleanUnityLogs()

	return result, nil
}

func (p *unityProvider) cleanUnityLogs() {
	home, _ := os.UserHomeDir()
	logPaths := []string{
		filepath.Join(home, "Library", "Logs", "Unity"),
		filepath.Join(home, "AppData", "Local", "Unity", "Editor"),
	}

	for _, logPath := range logPaths {
		if _, err := os.Stat(logPath); err == nil {
			cleanPathDirect(logPath)
		}
	}
}

func isUnitySafeToClean(path string) bool {
	// 这些路径可以安全清理
	safePatterns := []string{
		"Library/Unity/Cache",
		"Caches/Unity",
		"Logs/Unity",
	}

	for _, pattern := range safePatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

// 使用 Unity Hub 或命令行清理缓存
func (p *unityProvider) cleanViaUnityHub() {
	// Unity Hub 命令行工具可能不存在，静默失败
	cmd := exec.Command("unity-hub", "--batchmode", "-quit")
	cmd.Run()
}
