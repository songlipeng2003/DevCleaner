package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/devcleaner/backend/config"
)

// Provider 提供商接口
type Provider interface {
	ID() string
	Name() string
	Paths() []PathConfig
	Scan() ([]ScanResult, error)
	Clean(paths []string) (*CleanResult, error)
}

// PathConfig 路径配置
type PathConfig struct {
	Path        string
	Description string
	Strategy    CleanStrategy
	Command     string
}

// CleanStrategy 清理策略
type CleanStrategy int

const (
	StrategyDirect CleanStrategy = iota // 直接删除
	StrategyCommand                     // 使用命令清理
	StrategySafe                        // 安全删除（需确认）
)

// ScanResult 扫描结果
type ScanResult struct {
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	FileNum     int    `json:"file_num"`
	LastMod     int64  `json:"last_modified"`
	Description string `json:"description"`
}

// CleanResult 清理结果
type CleanResult struct {
	Cleaned  int64    `json:"cleaned"`
	Failed   []string `json:"failed"`
	FileNum  int      `json:"file_num"`
}

// ConfigProvider 基于配置的 Provider
type ConfigProvider struct {
	id          string
	name        string
	description string
	paths       []PathConfig
}

// NewConfigProviderFromConfig 从配置创建 Provider
func NewConfigProviderFromConfig(providerConfig config.ProviderConfig) *ConfigProvider {
	paths := make([]PathConfig, 0)

	// 转换主路径配置
	for _, pathConfig := range providerConfig.Paths {
		strategy := StrategyDirect
		if pathConfig.Strategy == "command" {
			strategy = StrategyCommand
		} else if pathConfig.Strategy == "safe" {
			strategy = StrategySafe
		}
		paths = append(paths, PathConfig{
			Path:        pathConfig.Path,
			Description: pathConfig.Description,
			Strategy:    strategy,
			Command:     pathConfig.Command,
		})
	}

	// 转换 IDEs 的路径配置
	for _, ide := range providerConfig.IDEs {
		for _, pathConfig := range ide.Paths {
			strategy := StrategyDirect
			if pathConfig.Strategy == "command" {
				strategy = StrategyCommand
			} else if pathConfig.Strategy == "safe" {
				strategy = StrategySafe
			}
			paths = append(paths, PathConfig{
				Path:        pathConfig.Path,
				Description: ide.Name + " " + pathConfig.Description,
				Strategy:    strategy,
				Command:     pathConfig.Command,
			})
		}
	}

	// 转换 CleanItems 的路径配置
	for _, item := range providerConfig.CleanItems {
		for _, pathConfig := range item.Paths {
			strategy := StrategyDirect
			if pathConfig.Strategy == "command" {
				strategy = StrategyCommand
			} else if pathConfig.Strategy == "safe" {
				strategy = StrategySafe
			}
			paths = append(paths, PathConfig{
				Path:        pathConfig.Path,
				Description: item.Name + " " + pathConfig.Description,
				Strategy:    strategy,
				Command:     pathConfig.Command,
			})
		}
	}

	return &ConfigProvider{
		id:          providerConfig.ID,
		name:        providerConfig.Name,
		description: providerConfig.Description,
		paths:       paths,
	}
}

func (p *ConfigProvider) ID() string   { return p.id }
func (p *ConfigProvider) Name() string { return p.name }

func (p *ConfigProvider) Paths() []PathConfig {
	return p.paths
}

func (p *ConfigProvider) Scan() ([]ScanResult, error) {
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

func (p *ConfigProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// 查找对应的路径配置
		var pathConfig PathConfig
		for _, pc := range p.Paths() {
			expanded := expandPath(pc.Path)
			if expanded == path {
				pathConfig = pc
				break
			}
		}

		// 根据策略清理
		if pathConfig.Strategy == StrategyCommand && pathConfig.Command != "" {
			cleaned, err := p.cleanByCommand(pathConfig.Command)
			if err == nil {
				result.Cleaned += cleaned
				continue
			}
		}

		// 直接删除
		cleaned, failed := cleanPathDirect(path)
		result.Cleaned += cleaned
		result.FileNum += len(failed)
		result.Failed = append(result.Failed, failed...)
	}

	return result, nil
}

func (p *ConfigProvider) cleanByCommand(command string) (int64, error) {
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return 0, nil
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	if _, err := cmd.CombinedOutput(); err != nil {
		return 0, err
	}

	return 0, nil
}

// npmProvider npm 提供商
type npmProvider struct{}

func NewNPMProvider() Provider {
	return &npmProvider{}
}

func (p *npmProvider) ID() string   { return "npm" }
func (p *npmProvider) Name() string { return "npm" }

func (p *npmProvider) Paths() []PathConfig {
	return []PathConfig{
		{
			Path:        "~/.npm",
			Description: "npm 全局缓存",
			Strategy:    StrategyCommand,
			Command:     "npm cache clean --force",
		},
		{
			Path:        "~/Library/Caches/npm",
			Description: "npm 缓存（macOS）",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.npm/_cacache",
			Description: "npm 内容缓存",
			Strategy:    StrategyCommand,
			Command:     "npm cache clean --force",
		},
		{
			Path:        "%APPDATA%\\npm-cache",
			Description: "npm 缓存（Windows）",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\npm-cache",
			Description: "npm 本地缓存（Windows）",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/npm",
			Description: "npm 缓存（Linux）",
			Strategy:    StrategyDirect,
		},
	}
}

func (p *npmProvider) Scan() ([]ScanResult, error) {
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

func (p *npmProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// 尝试使用 npm 命令清理
		if isNPMPackageCache(path) {
			cleaned, err := p.cleanByCommand()
			if err == nil {
				result.Cleaned += cleaned
				continue
			}
		}

		// 直接删除
		cleaned, failed := p.cleanDirect(path)
		result.Cleaned += cleaned
		result.FileNum += len(failed)
		result.Failed = append(result.Failed, failed...)
	}

	return result, nil
}

func (p *npmProvider) cleanByCommand() (int64, error) {
	cmd := exec.Command("npm", "cache", "clean", "--force")
	if _, err := cmd.CombinedOutput(); err != nil {
		return 0, err
	}

	// 尝试获取清理的大小（通过扫描清理前的缓存大小估算）
	// 这里简化处理，实际可记录清理前的状态
	return 0, nil
}

func (p *npmProvider) cleanDirect(path string) (int64, []string) {
	var totalSize int64
	var failed []string

	filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			failed = append(failed, walkPath)
			return nil
		}

		if !info.IsDir() {
			if err := os.Remove(walkPath); err != nil {
				failed = append(failed, walkPath)
			} else {
				totalSize += info.Size()
			}
		}
		return nil
	})

	// 删除空目录
	filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			os.Remove(walkPath)
		}
		return nil
	})

	return totalSize, failed
}

// yarnProvider Yarn 提供商
type yarnProvider struct{}

func NewYarnProvider() Provider {
	return &yarnProvider{}
}

func (p *yarnProvider) ID() string   { return "yarn" }
func (p *yarnProvider) Name() string { return "Yarn" }

func (p *yarnProvider) Paths() []PathConfig {
	return []PathConfig{
		{
			Path:        "~/.yarn-cache",
			Description: "Yarn 缓存",
			Strategy:    StrategyCommand,
			Command:     "yarn cache clean",
		},
		{
			Path:        "~/Library/Caches/Yarn",
			Description: "Yarn 缓存（macOS）",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.cache/yarn",
			Description: "Yarn 缓存（Linux）",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%LOCALAPPDATA%\\Yarn",
			Description: "Yarn 缓存（Windows）",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%APPDATA%\\Yarn",
			Description: "Yarn 全局缓存（Windows）",
			Strategy:    StrategyDirect,
		},
	}
}

func (p *yarnProvider) Scan() ([]ScanResult, error) {
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

func (p *yarnProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// 尝试使用 yarn 命令清理
		cmd := exec.Command("yarn", "cache", "clean")
		if err := cmd.Run(); err != nil {
			// 命令失败，直接删除
			cleaned, failed := cleanPathDirect(path)
			result.Cleaned += cleaned
			result.FileNum += len(failed)
			result.Failed = append(result.Failed, failed...)
		}
	}

	return result, nil
}

// dockerProvider Docker 提供商
type dockerProvider struct{}

func NewDockerProvider() Provider {
	return &dockerProvider{}
}

func (p *dockerProvider) ID() string   { return "docker" }
func (p *dockerProvider) Name() string { return "Docker" }

func (p *dockerProvider) Paths() []PathConfig {
	return []PathConfig{
		{
			Path:        "~/Library/Containers/com.docker.docker/Data/vms",
			Description: "Docker VM 数据",
			Strategy:    StrategyCommand,
			Command:     "docker system prune -a -f",
		},
		{
			Path:        "/var/lib/docker",
			Description: "Docker 数据目录（Linux）",
			Strategy:    StrategyCommand,
			Command:     "docker system prune -a -f",
		},
		{
			Path:        "C:\\ProgramData\\docker",
			Description: "Docker 数据目录（Windows）",
			Strategy:    StrategyCommand,
			Command:     "docker system prune -a -f",
		},
	}
}

func (p *dockerProvider) Scan() ([]ScanResult, error) {
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

func (p *dockerProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	// Docker 必须使用命令清理
	commands := [][]string{
		{"docker", "system", "df"},
		{"docker", "system", "prune", "-a", "-f"},
	}

	for _, args := range commands {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Run() // 忽略错误，继续尝试
	}

	// Docker 使用全局命令，不需要遍历 paths
	for range paths {
		result.Failed = append(result.Failed, "Docker cleanup attempted via docker CLI")
	}

	return result, nil
}

// 辅助函数
func expandPath(path string) string {
	home, _ := os.UserHomeDir()
	path = strings.ReplaceAll(path, "~", home)
	return os.ExpandEnv(path)
}

func isNPMPackageCache(path string) bool {
	return strings.Contains(path, ".npm") || strings.Contains(path, "npm-cache")
}

func cleanPathDirect(path string) (int64, []string) {
	var totalSize int64
	var failed []string

	filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			failed = append(failed, walkPath)
			return nil
		}

		if !info.IsDir() {
			if err := os.Remove(walkPath); err != nil {
				failed = append(failed, walkPath)
			} else {
				totalSize += info.Size()
			}
		}
		return nil
	})

	// 删除空目录
	filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() {
			os.Remove(walkPath)
		}
		return nil
	})

	return totalSize, failed
}

// GetProvider 获取指定类型的提供商
func GetProvider(id string) Provider {
	switch id {
	case "npm":
		return NewNPMProvider()
	case "yarn":
		return NewYarnProvider()
	case "pnpm":
		return NewPnpmProvider()
	case "docker":
		return NewDockerProvider()
	case "xcode":
		return NewXcodeProvider()
	case "homebrew":
		return NewHomebrewProvider()
	case "python":
		return NewPythonProvider()
	case "go":
		return NewGoProvider()
	case "ruby":
		return NewRubyProvider()
	case "maven":
		return NewMavenProvider()
	case "gradle":
		return NewGradleProvider()
	case "cocoapods":
		return NewCocoaPodsProvider()
	case "carthage":
		return NewCarthageProvider()
	case "unity":
		return NewUnityProvider()
	case "composer":
		return NewComposerProvider()
	case "cargo":
		return NewCargoProvider()
	case "flutter":
		return NewFlutterProvider()
	case "nuget":
		return NewNuGetProvider()
	case "android_sdk":
		return NewAndroidSDKProvider()
	case "jetbrains":
		return NewJetBrainsProvider()
	case "vscode":
		return NewVSCodeProvider()
	default:
		// 尝试从配置文件加载
		return GetProviderFromConfig(id)
	}
}

// GetProviderFromConfig 从配置文件获取 Provider
func GetProviderFromConfig(id string) Provider {
	cfg, err := config.LoadConfig()
	if err != nil || cfg == nil {
		return nil
	}

	providerConfig := cfg.GetProviderByID(id)
	if providerConfig == nil {
		return nil
	}

	return NewConfigProviderFromConfig(*providerConfig)
}

// GetAllProviders 获取所有提供商
func GetAllProviders() []Provider {
	return []Provider{
		NewNPMProvider(),
		NewYarnProvider(),
		NewPnpmProvider(),
		NewDockerProvider(),
		NewXcodeProvider(),
		NewHomebrewProvider(),
		NewPythonProvider(),
		NewGoProvider(),
		NewRubyProvider(),
		NewMavenProvider(),
		NewGradleProvider(),
		NewCocoaPodsProvider(),
		NewCarthageProvider(),
		NewUnityProvider(),
		NewComposerProvider(),
		NewCargoProvider(),
		NewFlutterProvider(),
		NewNuGetProvider(),
		NewAndroidSDKProvider(),
		NewJetBrainsProvider(),
		NewVSCodeProvider(),
	}
}

// GetAllProvidersFromConfig 获取所有配置中的提供商（包括未硬编码的）
func GetAllProvidersFromConfig() []Provider {
	cfg, err := config.LoadConfig()
	if err != nil || cfg == nil {
		return GetAllProviders()
	}

	var providers []Provider
	for _, providerConfig := range cfg.Providers {
		providers = append(providers, NewConfigProviderFromConfig(providerConfig))
	}

	return providers
}
