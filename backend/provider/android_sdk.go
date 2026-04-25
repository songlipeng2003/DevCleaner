package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"github.com/devcleaner/backend/scanner"
)

// androidSDKProvider Android SDK 提供商
type androidSDKProvider struct{}

func NewAndroidSDKProvider() Provider {
	return &androidSDKProvider{}
}

func (p *androidSDKProvider) ID() string   { return "android_sdk" }
func (p *androidSDKProvider) Name() string { return "Android SDK" }

func (p *androidSDKProvider) Paths() []PathConfig {
	// 根据操作系统确定 Android SDK 默认路径
	var sdkPath string
	switch runtime.GOOS {
	case "darwin":
		sdkPath = "~/Library/Android/sdk"
	case "linux":
		sdkPath = "~/Android/Sdk"
	case "windows":
		sdkPath = "%USERPROFILE%\\AppData\\Local\\Android\\Sdk"
	default:
		sdkPath = "~/Android/Sdk"
	}

	return []PathConfig{
		// Android SDK 构建缓存
		{
			Path:        sdkPath + "/.caches",
			Description: "Android SDK 构建缓存",
			Strategy:    StrategyDirect,
		},
		// Android 模拟器缓存
		{
			Path:        sdkPath + "/.android/avd",
			Description: "Android 虚拟设备缓存",
			Strategy:    StrategySafe, // 需谨慎，可能包含重要数据
		},
		// Gradle 缓存（Android 项目使用）
		{
			Path:        "~/.gradle/caches",
			Description: "Gradle 构建缓存",
			Strategy:    StrategyDirect,
		},
		// Android 构建缓存目录
		{
			Path:        "**/build",
			Description: "Android 项目构建输出目录",
			Strategy:    StrategySafe, // 需谨慎，可能包含重要构建产物
		},
		// Android 依赖缓存
		{
			Path:        sdkPath + "/extras",
			Description: "Android SDK 扩展包缓存",
			Strategy:    StrategySafe,
		},
		// Android 系统镜像缓存
		{
			Path:        sdkPath + "/system-images",
			Description: "Android 系统镜像缓存",
			Strategy:    StrategySafe,
		},
		// Android 平台工具缓存
		{
			Path:        sdkPath + "/platforms",
			Description: "Android 平台 SDK 缓存",
			Strategy:    StrategySafe,
		},
		// Android 构建工具缓存
		{
			Path:        sdkPath + "/build-tools",
			Description: "Android 构建工具缓存",
			Strategy:    StrategySafe,
		},
		// Android NDK 缓存
		{
			Path:        sdkPath + "/ndk",
			Description: "Android NDK 缓存",
			Strategy:    StrategySafe,
		},
		// Android 项目本地属性缓存
		{
			Path:        "**/local.properties",
			Description: "Android 项目本地属性文件",
			Strategy:    StrategySafe,
		},
	}
}

func (p *androidSDKProvider) Scan() ([]ScanResult, error) {
	var results []ScanResult

	// 扫描标准路径
	paths := []struct {
		path        string
		description string
	}{
		{expandPath(p.getSDKPath() + "/.caches"), "Android SDK 构建缓存"},
		{expandPath("~/.gradle/caches"), "Gradle 构建缓存"},
		{expandPath(p.getSDKPath() + "/extras"), "Android SDK 扩展包缓存"},
		{expandPath(p.getSDKPath() + "/system-images"), "Android 系统镜像缓存"},
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
			if strings.Contains(buildPath, "/.android/") || strings.Contains(buildPath, "/.gradle/") {
				continue
			}
			// 检查是否是 Android 项目（包含 AndroidManifest.xml 或 build.gradle）
			dir := filepath.Dir(buildPath)
			if p.isAndroidProject(dir) {
				if result, ok := p.scanSinglePath(buildPath, "Android 项目构建输出目录"); ok {
					results = append(results, result)
				}
			}
		}
	}

	return results, nil
}

func (p *androidSDKProvider) scanSinglePath(path, description string) (ScanResult, bool) {
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

func (p *androidSDKProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// 检查是否是 build 目录，需谨慎
		if strings.Contains(path, "build") && !strings.Contains(path, ".caches") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 build 目录，请使用 `./gradlew clean`")
			continue
		}
		// 检查是否是系统镜像、平台工具、构建工具等重要目录
		if strings.Contains(path, "system-images") ||
			strings.Contains(path, "platforms") ||
			strings.Contains(path, "build-tools") ||
			strings.Contains(path, "ndk") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 Android SDK 组件")
			continue
		}
		// 检查是否是虚拟设备
		if strings.Contains(path, "avd") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 Android 虚拟设备")
			continue
		}
		// 检查是否是 local.properties 文件
		if strings.HasSuffix(path, "local.properties") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 local.properties 文件")
			continue
		}
		// 其他缓存可以安全清理
		cleaned, failed := cleanPathDirect(path)
		result.Cleaned += cleaned
		result.Failed = append(result.Failed, failed...)
	}

	// 运行 gradle 清理命令
	p.cleanGradleCache()

	return result, nil
}

func (p *androidSDKProvider) cleanGradleCache() {
	// 清理 Gradle 缓存
	cmd := exec.Command("gradle", "--stop")
	cmd.Run() // 忽略错误

	cmd = exec.Command("gradle", "clean")
	cmd.Run()

	// 也可以使用 ./gradlew
	cmd = exec.Command("sh", "-c", "find . -name 'gradlew' -type f -exec {} clean \\;")
	cmd.Run()
}

func (p *androidSDKProvider) getSDKPath() string {
	switch runtime.GOOS {
	case "darwin":
		return "~/Library/Android/sdk"
	case "linux":
		return "~/Android/Sdk"
	case "windows":
		return "%USERPROFILE%\\AppData\\Local\\Android\\Sdk"
	default:
		return "~/Android/Sdk"
	}
}

func (p *androidSDKProvider) isAndroidProject(dir string) bool {
	// 检查目录是否包含 AndroidManifest.xml 或 build.gradle 文件
	androidManifest := filepath.Join(dir, "app", "src", "main", "AndroidManifest.xml")
	if _, err := os.Stat(androidManifest); err == nil {
		return true
	}

	buildGradle := filepath.Join(dir, "build.gradle")
	if _, err := os.Stat(buildGradle); err == nil {
		return true
	}

	buildGradleKts := filepath.Join(dir, "build.gradle.kts")
	if _, err := os.Stat(buildGradleKts); err == nil {
		return true
	}

	return false
}

// GetAndroidSDKInfo 获取 Android SDK 环境信息
func GetAndroidSDKInfo() (map[string]string, error) {
	info := make(map[string]string)

	// Android SDK 路径
	sdkPath := expandPath("~/Library/Android/sdk")
	if _, err := os.Stat(sdkPath); os.IsNotExist(err) {
		sdkPath = expandPath("~/Android/Sdk")
	}
	info["android_sdk_path"] = sdkPath

	// 检查 adb 版本
	adbPath := filepath.Join(sdkPath, "platform-tools", "adb")
	if _, err := os.Stat(adbPath); err == nil {
		cmd := exec.Command(adbPath, "version")
		if output, err := cmd.Output(); err == nil {
			lines := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(lines) > 0 {
				info["adb_version"] = lines[0]
			}
		}
	}

	// 检查 Gradle 版本
	cmd := exec.Command("gradle", "--version")
	if output, err := cmd.Output(); err == nil {
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) > 0 {
			info["gradle_version"] = lines[0]
		}
	}

	// 缓存目录大小估算
	gradleCachePath := expandPath("~/.gradle/caches")
	if stat, err := os.Stat(gradleCachePath); err == nil && stat.IsDir() {
		var totalSize int64
		filepath.Walk(gradleCachePath, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
		info["gradle_cache_size"] = scanner.FormatSize(totalSize)
	}

	return info, nil
}

