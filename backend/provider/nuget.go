package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"github.com/devcleaner/backend/scanner"
)

// nugetProvider NuGet 提供商
type nugetProvider struct{}

func NewNuGetProvider() Provider {
	return &nugetProvider{}
}

func (p *nugetProvider) ID() string   { return "nuget" }
func (p *nugetProvider) Name() string { return "NuGet" }

func (p *nugetProvider) Paths() []PathConfig {
	// 根据操作系统确定默认路径
	var defaultCachePath string
	switch runtime.GOOS {
	case "windows":
		defaultCachePath = "%USERPROFILE%\\.nuget\\packages"
	case "darwin":
		defaultCachePath = "~/.nuget/packages"
	case "linux":
		defaultCachePath = "~/.nuget/packages"
	default:
		defaultCachePath = "~/.nuget/packages"
	}

	return []PathConfig{
		// NuGet 全局包缓存
		{
			Path:        defaultCachePath,
			Description: "NuGet 全局包缓存",
			Strategy:    StrategyDirect,
		},
		// NuGet HTTP 缓存
		{
			Path:        "~/.local/share/NuGet/Cache",
			Description: "NuGet HTTP 缓存 (Linux/macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%APPDATA%\\NuGet\\Cache",
			Description: "NuGet HTTP 缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// NuGet 插件缓存
		{
			Path:        "~/.nuget/plugins",
			Description: "NuGet 插件缓存",
			Strategy:    StrategyDirect,
		},
		// NuGet 临时文件
		{
			Path:        "**/obj",
			Description: "NuGet 恢复临时文件",
			Strategy:    StrategySafe, // 需谨慎，可能影响构建
		},
		// NuGet 配置缓存
		{
			Path:        "~/.nuget/NuGet",
			Description: "NuGet 配置缓存",
			Strategy:    StrategySafe,
		},
		// .NET 本地工具缓存
		{
			Path:        "~/.dotnet/tools",
			Description: ".NET 本地工具缓存",
			Strategy:    StrategySafe,
		},
		// .NET 模板缓存
		{
			Path:        "~/.templateengine",
			Description: ".NET 模板缓存",
			Strategy:    StrategyDirect,
		},
	}
}

func (p *nugetProvider) Scan() ([]ScanResult, error) {
	var results []ScanResult

	// 扫描标准路径
	paths := []struct {
		path        string
		description string
	}{
		{expandPath(p.getNuGetCachePath()), "NuGet 全局包缓存"},
		{expandPath(p.getNuGetHttpCachePath()), "NuGet HTTP 缓存"},
		{expandPath("~/.nuget/plugins"), "NuGet 插件缓存"},
		{expandPath("~/.templateengine"), ".NET 模板缓存"},
	}

	for _, pathItem := range paths {
		if result, ok := p.scanSinglePath(pathItem.path, pathItem.description); ok {
			results = append(results, result)
		}
	}

	// 扫描 obj 目录（限制深度，避免扫描整个文件系统）
	home, _ := os.UserHomeDir()
	objPattern := filepath.Join(home, "**", "obj")
	matches, _ := filepath.Glob(objPattern)
	if matches != nil {
		for _, objPath := range matches {
			// 跳过某些常见项目的 obj 目录
			if strings.Contains(objPath, "/.nuget/") || strings.Contains(objPath, "/.dotnet/") {
				continue
			}
			// 检查是否是 .NET 项目（包含 .csproj 或 .fsproj）
			dir := filepath.Dir(objPath)
			if p.isDotNetProject(dir) {
				if result, ok := p.scanSinglePath(objPath, "NuGet 恢复临时文件"); ok {
					results = append(results, result)
				}
			}
		}
	}

	return results, nil
}

func (p *nugetProvider) scanSinglePath(path, description string) (ScanResult, bool) {
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

func (p *nugetProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// 检查是否是 obj 目录，需谨慎
		if strings.Contains(path, "obj") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 obj 目录，可能影响构建")
			continue
		}
		// 检查是否是 .dotnet/tools 目录
		if strings.Contains(path, ".dotnet/tools") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 .NET 工具目录")
			continue
		}
		// 检查是否是配置缓存
		if strings.Contains(path, "NuGet/NuGet.Config") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 NuGet 配置文件")
			continue
		}
		// 其他缓存可以安全清理
		cleaned, failed := cleanPathDirect(path)
		result.Cleaned += cleaned
		result.Failed = append(result.Failed, failed...)
	}

	// 运行 dotnet nuget locals clear 命令
	p.clearNuGetCache()

	return result, nil
}

func (p *nugetProvider) clearNuGetCache() {
	// 清理所有 NuGet 本地缓存
	cmd := exec.Command("dotnet", "nuget", "locals", "all", "--clear")
	cmd.Run() // 忽略错误
}

func (p *nugetProvider) getNuGetCachePath() string {
	switch runtime.GOOS {
	case "windows":
		return "%USERPROFILE%\\.nuget\\packages"
	case "darwin", "linux":
		return "~/.nuget/packages"
	default:
		return "~/.nuget/packages"
	}
}

func (p *nugetProvider) getNuGetHttpCachePath() string {
	switch runtime.GOOS {
	case "windows":
		return "%APPDATA%\\NuGet\\Cache"
	case "darwin", "linux":
		return "~/.local/share/NuGet/Cache"
	default:
		return "~/.local/share/NuGet/Cache"
	}
}

func (p *nugetProvider) isDotNetProject(dir string) bool {
	// 检查目录是否包含 .csproj、.fsproj、.vbproj 文件
	patterns := []string{"*.csproj", "*.fsproj", "*.vbproj"}
	for _, pattern := range patterns {
		matches, _ := filepath.Glob(filepath.Join(dir, pattern))
		if matches != nil && len(matches) > 0 {
			return true
		}
	}
	return false
}

// GetNuGetInfo 获取 NuGet 环境信息
func GetNuGetInfo() (map[string]string, error) {
	info := make(map[string]string)

	// .NET SDK 版本
	cmd := exec.Command("dotnet", "--version")
	if output, err := cmd.Output(); err == nil {
		info["dotnet_version"] = strings.TrimSpace(string(output))
	}

	// NuGet 版本
	cmd = exec.Command("dotnet", "nuget", "--version")
	if output, err := cmd.Output(); err == nil {
		info["nuget_version"] = strings.TrimSpace(string(output))
	}

	// 缓存目录大小估算
	cachePath := expandPath("~/.nuget/packages")
	if stat, err := os.Stat(cachePath); err == nil && stat.IsDir() {
		var totalSize int64
		filepath.Walk(cachePath, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
		info["nuget_cache_size"] = scanner.FormatSize(totalSize)
	}

	return info, nil
}

