package tools

import "os"

// Tool 定义支持的开发工具
type Tool struct {
	ID    string
	Name  string
	Paths []PathConfig
}

// PathConfig 路径配置
type PathConfig struct {
	Path        string
	Description string
}

// AllTools 所有支持的开发工具
var AllTools = []Tool{
	{
		ID:   "npm",
		Name: "npm",
		Paths: []PathConfig{
			{Path: "~/.npm", Description: "npm 缓存"},
			{Path: "~/Library/Caches/npm", Description: "npm 全局缓存 (macOS)"},
			{Path: "%APPDATA%\\npm-cache", Description: "npm 缓存 (Windows)"},
			{Path: "%LOCALAPPDATA%\\npm-cache", Description: "npm 本地缓存 (Windows)"},
			{Path: "~/.cache/npm", Description: "npm 缓存 (Linux)"},
		},
	},
	{
		ID:   "yarn",
		Name: "Yarn",
		Paths: []PathConfig{
			{Path: "~/.yarn-cache", Description: "Yarn 缓存"},
			{Path: "~/Library/Caches/Yarn", Description: "Yarn 全局缓存 (macOS)"},
			{Path: "%LOCALAPPDATA%\\Yarn", Description: "Yarn 缓存 (Windows)"},
			{Path: "%APPDATA%\\Yarn", Description: "Yarn 全局缓存 (Windows)"},
			{Path: "~/.cache/yarn", Description: "Yarn 缓存 (Linux)"},
		},
	},
	{
		ID:   "pnpm",
		Name: "pnpm",
		Paths: []PathConfig{
			{Path: "~/.pnpm-store", Description: "pnpm 存储"},
			{Path: "~/Library/Caches/pnpm", Description: "pnpm 全局缓存 (macOS)"},
			{Path: "%LOCALAPPDATA%\\pnpm", Description: "pnpm 缓存 (Windows)"},
			{Path: "%APPDATA%\\pnpm", Description: "pnpm 全局缓存 (Windows)"},
			{Path: "~/.cache/pnpm", Description: "pnpm 缓存 (Linux)"},
		},
	},
	{
		ID:   "docker",
		Name: "Docker",
		Paths: []PathConfig{
			{Path: "~/Library/Containers/com.docker.docker/Data/vms", Description: "Docker VM (macOS)"},
			{Path: "/var/lib/docker", Description: "Docker 数据目录 (Linux)"},
			{Path: "C:\\ProgramData\\docker", Description: "Docker 数据目录 (Windows)"},
			{Path: "%USERPROFILE%\\AppData\\Local\\Docker", Description: "Docker 本地数据 (Windows)"},
		},
	},
	{
		ID:   "xcode",
		Name: "Xcode",
		Paths: []PathConfig{
			{Path: "~/Library/Developer/Xcode/DerivedData", Description: "编译缓存"},
			{Path: "~/Library/Developer/Xcode/Archives", Description: "归档文件"},
			{Path: "~/Library/Developer/Xcode/iOS DeviceSupport", Description: "设备支持"},
			{Path: "~/Library/Caches/com.apple.dt.Xcode", Description: "Xcode 缓存"},
		},
	},
	{
		ID:   "homebrew",
		Name: "Homebrew",
		Paths: []PathConfig{
			{Path: "$(brew --cache)", Description: "Homebrew 下载缓存"},
			{Path: "/usr/local/Cellar", Description: "Homebrew Cellar (Intel macOS)"},
			{Path: "/opt/homebrew/Cellar", Description: "Homebrew Cellar (Apple Silicon macOS)"},
			{Path: "~/Library/Caches/Homebrew", Description: "Homebrew 缓存 (macOS)"},
			{Path: "/home/linuxbrew/.linuxbrew/Cellar", Description: "Homebrew Cellar (Linux)"},
			{Path: "~/.cache/Homebrew", Description: "Homebrew 缓存 (Linux)"},
		},
	},
	{
		ID:   "python",
		Name: "Python",
		Paths: []PathConfig{
			{Path: "~/.cache/pip", Description: "pip 缓存 (Linux/macOS)"},
			{Path: "~/Library/Caches/pip", Description: "pip 全局缓存 (macOS)"},
			{Path: "%APPDATA%\\pip\\cache", Description: "pip 缓存 (Windows)"},
			{Path: "%LOCALAPPDATA%\\pip\\cache", Description: "pip 本地缓存 (Windows)"},
		},
	},
	{
		ID:   "go",
		Name: "Go",
		Paths: []PathConfig{
			{Path: "$(go env GOPATH)/pkg/mod", Description: "Go 模块缓存"},
		},
	},
	{
		ID:   "ruby",
		Name: "Ruby",
		Paths: []PathConfig{
			{Path: "~/.gem/cache", Description: "gem 本地缓存"},
			{Path: "~/Library/Caches/bundler", Description: "Bundler 缓存 (macOS)"},
			{Path: "~/.cache/bundler", Description: "Bundler 缓存 (Linux)"},
			{Path: "%APPDATA%\\bundler", Description: "Bundler 缓存 (Windows)"},
			{Path: "%LOCALAPPDATA%\\bundler", Description: "Bundler 本地缓存 (Windows)"},
		},
	},
	{
		ID:   "maven",
		Name: "Maven",
		Paths: []PathConfig{
			{Path: "~/.m2/repository", Description: "Maven 本地仓库"},
			{Path: "%USERPROFILE%\\.m2\\repository", Description: "Maven 本地仓库 (Windows)"},
			{Path: "/root/.m2/repository", Description: "Maven 本地仓库 (Linux)"},
		},
	},
	{
		ID:   "gradle",
		Name: "Gradle",
		Paths: []PathConfig{
			{Path: "~/.gradle/caches", Description: "Gradle 缓存目录"},
			{Path: "%USERPROFILE%\\.gradle\\caches", Description: "Gradle 缓存目录 (Windows)"},
			{Path: "/root/.gradle/caches", Description: "Gradle 缓存目录 (Linux)"},
			{Path: "~/Library/Caches/Gradle", Description: "Gradle 缓存 (macOS)"},
		},
	},
	{
		ID:   "cocoapods",
		Name: "CocoaPods",
		Paths: []PathConfig{
			{Path: "~/Library/Caches/CocoaPods", Description: "CocoaPods 缓存"},
			{Path: "~/Library/Developer/Xcode/DerivedData", Description: "Pods 构建缓存"},
		},
	},
	{
		ID:   "unity",
		Name: "Unity",
		Paths: []PathConfig{
			{Path: "~/Library/Unity/Cache", Description: "Unity 缓存 (macOS)"},
			{Path: "~/Library/Caches/Unity", Description: "Unity 下载缓存 (macOS)"},
			{Path: "%USERPROFILE%\\AppData\\Local\\Unity", Description: "Unity 缓存 (Windows)"},
			{Path: "%USERPROFILE%\\AppData\\LocalLow\\Unity", Description: "Unity 低完整性缓存 (Windows)"},
			{Path: "~/.cache/Unity", Description: "Unity 缓存 (Linux)"},
			{Path: "~/.config/unity3d", Description: "Unity 配置目录 (Linux)"},
		},
	},
	{
		ID:   "composer",
		Name: "Composer",
		Paths: []PathConfig{
			{Path: "~/.cache/composer", Description: "Composer 全局缓存 (Linux/macOS)"},
			{Path: "~/Library/Caches/composer", Description: "Composer 全局缓存 (macOS)"},
			{Path: "%APPDATA%\\Composer", Description: "Composer 全局缓存 (Windows)"},
			{Path: "~/.composer/cache", Description: "Composer 本地缓存（旧版本）"},
		},
	},
	{
		ID:   "cargo",
		Name: "Cargo",
		Paths: []PathConfig{
			{Path: "~/.cargo/registry", Description: "Cargo 注册表缓存 (Linux/macOS)"},
			{Path: "~/.cargo/git", Description: "Cargo Git 依赖缓存 (Linux/macOS)"},
			{Path: "%USERPROFILE%\\.cargo\\registry", Description: "Cargo 注册表缓存 (Windows)"},
			{Path: "%USERPROFILE%\\.cargo\\git", Description: "Cargo Git 依赖缓存 (Windows)"},
		},
	},
	{
		ID:   "flutter",
		Name: "Flutter",
		Paths: []PathConfig{
			{Path: "~/.pub-cache", Description: "Pub 全局缓存 (Linux/macOS)"},
			{Path: "~/.flutter_cache", Description: "Flutter 全局缓存 (Linux/macOS)"},
			{Path: "%USERPROFILE%\\.pub-cache", Description: "Pub 全局缓存 (Windows)"},
			{Path: "%USERPROFILE%\\.flutter_cache", Description: "Flutter 全局缓存 (Windows)"},
		},
	},
	{
		ID:   "nuget",
		Name: "NuGet",
		Paths: []PathConfig{
			{Path: "~/.nuget/packages", Description: "NuGet 全局包缓存"},
			{Path: "%USERPROFILE%\\.nuget\\packages", Description: "NuGet 全局包缓存 (Windows)"},
			{Path: "~/.local/share/NuGet/Cache", Description: "NuGet HTTP 缓存 (Linux/macOS)"},
			{Path: "%APPDATA%\\NuGet\\Cache", Description: "NuGet HTTP 缓存 (Windows)"},
		},
	},
	{
		ID:   "android_sdk",
		Name: "Android SDK",
		Paths: []PathConfig{
			{Path: "~/Library/Android/sdk/.caches", Description: "Android SDK 构建缓存 (macOS)"},
			{Path: "%USERPROFILE%\\AppData\\Local\\Android\\Sdk\\.caches", Description: "Android SDK 构建缓存 (Windows)"},
			{Path: "~/Android/Sdk/.caches", Description: "Android SDK 构建缓存 (Linux)"},
			{Path: "~/.gradle/caches", Description: "Gradle 构建缓存 (Linux/macOS)"},
			{Path: "%USERPROFILE%\\.gradle\\caches", Description: "Gradle 构建缓存 (Windows)"},
		},
	},
}

// GetToolByID 根据 ID 获取工具
func GetToolByID(id string) *Tool {
	for i := range AllTools {
		if AllTools[i].ID == id {
			return &AllTools[i]
		}
	}
	return nil
}

// ExpandPath 展开路径中的特殊字符
func ExpandPath(path string) string {
	home, _ := os.UserHomeDir()
	path = os.ExpandEnv(path)
	path = os.Expand(path, func(key string) string {
		if key == "~" {
			return home
		}
		return ""
	})
	return path
}
