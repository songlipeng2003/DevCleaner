package provider

import (
	"os"
	"path/filepath"
	"runtime"
)

type vscodeProvider struct{}

func NewVSCodeProvider() Provider {
	return &vscodeProvider{}
}

func (p *vscodeProvider) ID() string   { return "vscode" }
func (p *vscodeProvider) Name() string  { return "VS Code" }

func (p *vscodeProvider) Paths() []PathConfig {
	home, _ := os.UserHomeDir()

	paths := []PathConfig{
		// VS Code 缓存目录
		{
			Path:        "~/.cache/VSCode",
			Description: "VS Code 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/vscode",
			Description: "VS Code 缓存 (Linux 小写)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Caches/Code",
			Description: "VS Code 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\VSCode",
			Description: "VS Code 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// VS Code 扩展缓存
		{
			Path:        "~/.vscode/extensions",
			Description: "VS Code 扩展",
			Strategy:    StrategySafe,
		},
		{
			Path:        "~/.vscode-insiders/extensions",
			Description: "VS Code Insiders 扩展",
			Strategy:    StrategySafe,
		},
		// VS Code Cached Data
		{
			Path:        "~/.config/Code/Cache",
			Description: "VS Code 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.config/Code/CachedData",
			Description: "VS Code 缓存数据 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.config/Code/CachedExtensions",
			Description: "VS Code 扩展缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.config/Code/CachedExtensionVSIXs",
			Description: "VS Code 扩展包缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Application Support/Code/Cache",
			Description: "VS Code 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Application Support/Code/CachedData",
			Description: "VS Code 缓存数据 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Application Support/Code/CachedExtensions",
			Description: "VS Code 扩展缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%APPDATA%\\Code\\Cache",
			Description: "VS Code 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%APPDATA%\\Code\\CachedData",
			Description: "VS Code 缓存数据 (Windows)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%APPDATA%\\Code\\CachedExtensions",
			Description: "VS Code 扩展缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// VS Code Insiders
		{
			Path:        "~/.config/Code - Insiders/Cache",
			Description: "VS Code Insiders 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Application Support/Code - Insiders/Cache",
			Description: "VS Code Insiders 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%APPDATA%\\Code - Insiders\\Cache",
			Description: "VS Code Insiders 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// VS Code GPU Cache
		{
			Path:        "~/.config/Code/GPUCache",
			Description: "VS Code GPU 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Application Support/Code/GPUCache",
			Description: "VS Code GPU 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%APPDATA%\\Code\\GPUCache",
			Description: "VS Code GPU 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// VS Code Logs
		{
			Path:        "~/.config/Code/logs",
			Description: "VS Code 日志 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Application Support/Code/logs",
			Description: "VS Code 日志 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%APPDATA%\\Code\\logs",
			Description: "VS Code 日志 (Windows)",
			Strategy:    StrategyDirect,
		},
		// VS Code IndexedDB
		{
			Path:        "~/.config/Code/IndexedDB",
			Description: "VS Code IndexedDB (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Application Support/Code/IndexedDB",
			Description: "VS Code IndexedDB (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%APPDATA%\\Code\\IndexedDB",
			Description: "VS Code IndexedDB (Windows)",
			Strategy:    StrategyDirect,
		},
		// VS Code Global Storage
		{
			Path:        "~/.config/Code/User/globalStorage",
			Description: "VS Code 全局存储 (Linux)",
			Strategy:    StrategySafe,
		},
		{
			Path:        "~/Library/Application Support/Code/User/globalStorage",
			Description: "VS Code 全局存储 (macOS)",
			Strategy:    StrategySafe,
		},
		{
			Path:        "%APPDATA%\\Code\\User\\globalStorage",
			Description: "VS Code 全局存储 (Windows)",
			Strategy:    StrategySafe,
		},
	}

	// 根据操作系统添加特定路径
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths,
			PathConfig{
				Path:        home + "/Library/Logs/Code",
				Description: "VS Code 日志 (macOS)",
				Strategy:    StrategyDirect,
			},
			PathConfig{
				Path:        home + "/Library/Saved Application State/com.microsoft.VSCode.savedState",
				Description: "VS Code 保存状态 (macOS)",
				Strategy:    StrategySafe,
			},
		)
	case "linux":
		paths = append(paths,
			PathConfig{
				Path:        home + "/.config/Code/storage.json",
				Description: "VS Code 存储 (Linux)",
				Strategy:    StrategySafe,
			},
		)
	case "windows":
		paths = append(paths,
			PathConfig{
				Path:        "%LOCALAPPDATA%\\Programs\\Microsoft VS Code\\Cache",
				Description: "VS Code 程序缓存 (Windows)",
				Strategy:    StrategyDirect,
			},
		)
	}

	return paths
}

func (p *vscodeProvider) Scan() ([]ScanResult, error) {
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

func (p *vscodeProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{}

	for _, path := range paths {
		cleaned, failed := cleanPathDirect(path)
		result.Cleaned += cleaned
		result.Failed = append(result.Failed, failed...)
	}

	return result, nil
}
