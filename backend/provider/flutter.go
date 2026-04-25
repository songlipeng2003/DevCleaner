package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/devcleaner/backend/scanner"
)

// flutterProvider Flutter 提供商
type flutterProvider struct{}

func NewFlutterProvider() Provider {
	return &flutterProvider{}
}

func (p *flutterProvider) ID() string   { return "flutter" }
func (p *flutterProvider) Name() string { return "Flutter" }

func (p *flutterProvider) Paths() []PathConfig {
	return []PathConfig{
		// Flutter SDK 缓存
		{
			Path:        "~/.flutter",
			Description: "Flutter SDK 缓存",
			Strategy:    StrategySafe,
		},
		// Flutter 缓存目录
		{
			Path:        "~/.flutter_cache",
			Description: "Flutter 全局缓存",
			Strategy:    StrategyDirect,
		},
		// Pub 全局缓存
		{
			Path:        "~/.pub-cache",
			Description: "Pub 包管理器全局缓存",
			Strategy:    StrategyDirect,
		},
		// Flutter 构建缓存
		{
			Path:        "**/build",
			Description: "Flutter 构建输出目录",
			Strategy:    StrategySafe, // 需谨慎，可能包含重要构建产物
		},
		// Flutter 依赖缓存（项目级）
		{
			Path:        "**/.dart_tool",
			Description: "Dart/Flutter 项目依赖缓存",
			Strategy:    StrategyDirect,
		},
		// Flutter 模拟器缓存
		{
			Path:        "~/Library/Android/sdk/.android/avd",
			Description: "Android 虚拟设备缓存 (macOS)",
			Strategy:    StrategySafe,
		},
		{
			Path:        "~/.android/avd",
			Description: "Android 虚拟设备缓存 (Linux)",
			Strategy:    StrategySafe,
		},
		{
			Path:        "%USERPROFILE%\\.android\\avd",
			Description: "Android 虚拟设备缓存 (Windows)",
			Strategy:    StrategySafe,
		},
	}
}

func (p *flutterProvider) Scan() ([]ScanResult, error) {
	var results []ScanResult

	// 扫描标准路径
	paths := []struct {
		path        string
		description string
	}{
		{expandPath("~/.pub-cache"), "Pub 全局缓存"},
		{expandPath("~/.flutter_cache"), "Flutter 全局缓存"},
		{expandPath("~/.flutter"), "Flutter SDK 缓存"},
	}

	for _, pathItem := range paths {
		if result, ok := p.scanSinglePath(pathItem.path, pathItem.description); ok {
			results = append(results, result)
		}
	}

	// 扫描 build 目录（限制深度，避免扫描整个文件系统）
	home, _ := os.UserHomeDir()
	buildPattern := filepath.Join(home, "**", "build")
	matches, _ := filepath.Glob(buildPattern)
	if matches != nil {
		for _, buildPath := range matches {
			// 跳过某些常见项目的 build 目录
			if strings.Contains(buildPath, "/.flutter/") || strings.Contains(buildPath, "/.pub-cache/") {
				continue
			}
			// 检查是否是 Flutter 项目（包含 pubspec.yaml）
			pubspecPath := filepath.Join(filepath.Dir(buildPath), "pubspec.yaml")
			if _, err := os.Stat(pubspecPath); err == nil {
				if result, ok := p.scanSinglePath(buildPath, "Flutter 构建输出目录"); ok {
					results = append(results, result)
				}
			}
		}
	}

	// 扫描 .dart_tool 目录
	dartToolPattern := filepath.Join(home, "**", ".dart_tool")
	matches, _ = filepath.Glob(dartToolPattern)
	if matches != nil {
		for _, dartToolPath := range matches {
			if result, ok := p.scanSinglePath(dartToolPath, "Dart/Flutter 项目依赖缓存"); ok {
				results = append(results, result)
			}
		}
	}

	return results, nil
}

func (p *flutterProvider) scanSinglePath(path, description string) (ScanResult, bool) {
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

func (p *flutterProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// 检查是否是 build 目录，需谨慎
		if strings.Contains(path, "build") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 build 目录，请使用 `flutter clean`")
			continue
		}
		// 检查是否是 Flutter SDK 目录
		if strings.Contains(path, ".flutter") && !strings.Contains(path, ".flutter_cache") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 Flutter SDK 目录")
			continue
		}
		// 检查是否是 Android 虚拟设备
		if strings.Contains(path, "avd") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 Android 虚拟设备")
			continue
		}
		// 其他缓存可以安全清理
		cleaned, failed := cleanPathDirect(path)
		result.Cleaned += cleaned
		result.Failed = append(result.Failed, failed...)
	}

	// 运行 flutter pub cache clean 命令
	p.cleanPubCache()

	return result, nil
}

func (p *flutterProvider) cleanPubCache() {
	// 清理 pub 缓存
	cmd := exec.Command("flutter", "pub", "cache", "clean")
	cmd.Run() // 忽略错误

	// 也可以使用 dart 命令
	cmd = exec.Command("dart", "pub", "cache", "clean")
	cmd.Run()
}

// GetFlutterInfo 获取 Flutter 环境信息
func GetFlutterInfo() (map[string]string, error) {
	info := make(map[string]string)

	// Flutter 版本
	cmd := exec.Command("flutter", "--version")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) > 0 {
			info["flutter_version"] = lines[0]
		}
	}

	// Dart 版本
	cmd = exec.Command("dart", "--version")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) > 0 {
			info["dart_version"] = lines[0]
		}
	}

	// Pub 缓存目录大小估算
	pubCachePath := expandPath("~/.pub-cache")
	if stat, err := os.Stat(pubCachePath); err == nil && stat.IsDir() {
		var totalSize int64
		filepath.Walk(pubCachePath, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
		info["pub_cache_size"] = scanner.FormatSize(totalSize)
	}

	return info, nil
}

