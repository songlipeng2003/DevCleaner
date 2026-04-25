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
	// 优先从配置文件加载
	provider := GetProviderFromConfig(id)
	if provider != nil {
		return provider
	}
	// 配置中没有找到，返回nil
	return nil
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
	return GetAllProvidersFromConfig()
}

// GetAllProvidersFromConfig 获取所有配置中的提供商（包括未硬编码的）
func GetAllProvidersFromConfig() []Provider {
	cfg, err := config.LoadConfig()
	if err != nil || cfg == nil {
		return []Provider{} // 配置加载失败，返回空列表
	}

	var providers []Provider
	for _, providerConfig := range cfg.Providers {
		providers = append(providers, NewConfigProviderFromConfig(providerConfig))
	}

	return providers
}
