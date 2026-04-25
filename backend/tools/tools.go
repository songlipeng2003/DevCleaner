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
			{Path: "~/Library/Caches/npm", Description: "npm 全局缓存"},
		},
	},
	{
		ID:   "yarn",
		Name: "Yarn",
		Paths: []PathConfig{
			{Path: "~/.yarn-cache", Description: "Yarn 缓存"},
			{Path: "~/Library/Caches/Yarn", Description: "Yarn 全局缓存"},
		},
	},
	{
		ID:   "pnpm",
		Name: "pnpm",
		Paths: []PathConfig{
			{Path: "~/.pnpm-store", Description: "pnpm 存储"},
			{Path: "~/Library/Caches/pnpm", Description: "pnpm 全局缓存"},
		},
	},
	{
		ID:   "docker",
		Name: "Docker",
		Paths: []PathConfig{
			{Path: "~/Library/Containers/com.docker.docker/Data/vms", Description: "Docker VM"},
			{Path: "/var/lib/docker", Description: "Docker 数据"},
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
			{Path: "$(brew --cache)", Description: "Homebrew 缓存"},
			{Path: "/usr/local/Cellar", Description: "Homebrew Cellar"},
			{Path: "~/Library/Caches/Homebrew", Description: "Homebrew 下载缓存"},
		},
	},
	{
		ID:   "python",
		Name: "Python",
		Paths: []PathConfig{
			{Path: "~/.cache/pip", Description: "pip 缓存"},
			{Path: "~/Library/Caches/pip", Description: "pip 全局缓存"},
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
			{Path: "~/.gem/cache", Description: "gem 缓存"},
			{Path: "~/Library/Caches bundler", Description: "bundler 缓存"},
		},
	},
	{
		ID:   "maven",
		Name: "Maven",
		Paths: []PathConfig{
			{Path: "~/.m2/repository", Description: "Maven 本地仓库"},
		},
	},
	{
		ID:   "gradle",
		Name: "Gradle",
		Paths: []PathConfig{
			{Path: "~/.gradle/caches", Description: "Gradle 缓存"},
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
			{Path: "~/Library/Unity/Cache", Description: "Unity 缓存"},
			{Path: "~/Library/Caches/Unity", Description: "Unity 下载缓存"},
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
