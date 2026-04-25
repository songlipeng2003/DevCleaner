package provider

import (
	"os"
	"path/filepath"
	"runtime"
)

type jetbrainsProvider struct{}

func NewJetBrainsProvider() Provider {
	return &jetbrainsProvider{}
}

func (p *jetbrainsProvider) ID() string   { return "jetbrains" }
func (p *jetbrainsProvider) Name() string { return "JetBrains IDEs" }

func (p *jetbrainsProvider) Paths() []PathConfig {
	home, _ := os.UserHomeDir()

	paths := []PathConfig{
		// JetBrains 通用配置目录
		{
			Path:        "~/.JetBrains",
			Description: "JetBrains 配置 (旧版本)",
			Strategy:    StrategySafe,
		},
		{
			Path:        "~/.config/JetBrains",
			Description: "JetBrains 配置 (Linux)",
			Strategy:    StrategySafe,
		},
		{
			Path:        "~/Library/Application Support/JetBrains",
			Description: "JetBrains 配置 (macOS)",
			Strategy:    StrategySafe,
		},
		{
			Path:        "%APPDATA%\\JetBrains",
			Description: "JetBrains 配置 (Windows)",
			Strategy:    StrategySafe,
		},
		// 各 IDE 缓存目录
		// IntelliJ IDEA
		{
			Path:        "~/Library/Caches/JetBrains/IntelliJIdea",
			Description: "IntelliJ IDEA 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/JetBrains/IntelliJIdea",
			Description: "IntelliJ IDEA 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\JetBrains\\IntelliJIdea",
			Description: "IntelliJ IDEA 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// WebStorm
		{
			Path:        "~/Library/Caches/JetBrains/WebStorm",
			Description: "WebStorm 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/JetBrains/WebStorm",
			Description: "WebStorm 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\JetBrains\\WebStorm",
			Description: "WebStorm 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// PyCharm
		{
			Path:        "~/Library/Caches/JetBrains/PyCharm",
			Description: "PyCharm 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/JetBrains/PyCharm",
			Description: "PyCharm 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\JetBrains\\PyCharm",
			Description: "PyCharm 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// GoLand
		{
			Path:        "~/Library/Caches/JetBrains/GoLand",
			Description: "GoLand 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/JetBrains/GoLand",
			Description: "GoLand 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\JetBrains\\GoLand",
			Description: "GoLand 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// DataGrip
		{
			Path:        "~/Library/Caches/JetBrains/DataGrip",
			Description: "DataGrip 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/JetBrains/DataGrip",
			Description: "DataGrip 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\JetBrains\\DataGrip",
			Description: "DataGrip 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// Rider
		{
			Path:        "~/Library/Caches/JetBrains/Rider",
			Description: "Rider 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/JetBrains/Rider",
			Description: "Rider 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\JetBrains\\Rider",
			Description: "Rider 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// CLion
		{
			Path:        "~/Library/Caches/JetBrains/CLion",
			Description: "CLion 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/JetBrains/CLion",
			Description: "CLion 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\JetBrains\\CLion",
			Description: "CLion 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// RubyMine
		{
			Path:        "~/Library/Caches/JetBrains/RubyMine",
			Description: "RubyMine 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/JetBrains/RubyMine",
			Description: "RubyMine 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\JetBrains\\RubyMine",
			Description: "RubyMine 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// PhpStorm
		{
			Path:        "~/Library/Caches/JetBrains/PhpStorm",
			Description: "PhpStorm 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/JetBrains/PhpStorm",
			Description: "PhpStorm 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\JetBrains\\PhpStorm",
			Description: "PhpStorm 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// Android Studio (基于 IntelliJ)
		{
			Path:        "~/Library/Caches/AndroidStudio",
			Description: "Android Studio 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/AndroidStudio",
			Description: "Android Studio 缓存 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\Google\\AndroidStudio",
			Description: "Android Studio 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
	}

	// 根据操作系统添加特定路径
	switch runtime.GOOS {
	case "darwin":
		paths = append(paths,
			PathConfig{
				Path:        home + "/Library/Logs/JetBrains",
				Description: "JetBrains 日志 (macOS)",
				Strategy:    StrategyDirect,
			},
		)
	case "linux":
		paths = append(paths,
			PathConfig{
				Path:        home + "/.local/share/JetBrains",
				Description: "JetBrains 数据 (Linux)",
				Strategy:    StrategySafe,
			},
		)
	case "windows":
		paths = append(paths,
			PathConfig{
				Path:        home + "\\.JetBrains",
				Description: "JetBrains 旧配置 (Windows)",
				Strategy:    StrategySafe,
			},
		)
	}

	return paths
}

func (p *jetbrainsProvider) Scan() ([]ScanResult, error) {
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

func (p *jetbrainsProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{}

	for _, path := range paths {
		cleaned, failed := cleanPathDirect(path)
		result.Cleaned += cleaned
		result.Failed = append(result.Failed, failed...)
	}

	return result, nil
}
