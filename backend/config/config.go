package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
)

// Config 配置结构
type Config struct {
	Version   string             `json:"version"`
	Providers []ProviderConfig  `json:"providers"`
}

// ProviderConfig Provider 配置
type ProviderConfig struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Platforms  []string          `json:"platforms"`
	Paths       []PathConfig      `json:"paths,omitempty"`
	IDEs        []IDEConfig       `json:"ides,omitempty"`
	CleanItems  []CleanItemConfig `json:"清理项,omitempty"`
}

// PathConfig 路径配置
type PathConfig struct {
	Path        string `json:"path"`
	Description string `json:"description"`
	Strategy    string `json:"strategy"`
	Command     string `json:"command,omitempty"`
}

// IDEConfig IDE 配置
type IDEConfig struct {
	ID    string        `json:"id"`
	Name  string        `json:"name"`
	Paths []PathConfig  `json:"paths"`
}

// CleanItemConfig 清理项配置
type CleanItemConfig struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Paths       []PathConfig `json:"paths"`
}

// GetCurrentPlatform 返回当前平台
func GetCurrentPlatform() string {
	switch runtime.GOOS {
	case "darwin":
		return "darwin"
	case "linux":
		return "linux"
	case "windows":
		return "windows"
	default:
		return ""
	}
}

// LoadConfig 加载配置文件
func LoadConfig() (*Config, error) {
	// 获取可执行文件所在目录
	execPath, err := os.Executable()
	if err != nil {
		execPath = os.Args[0]
	}
	execDir := filepath.Dir(execPath)

	// 可能的配置文件路径
	configPaths := []string{
		filepath.Join(execDir, "providers.json"),           // 与可执行文件同目录
		filepath.Join(execDir, "..", "providers.json"),    // 上级目录
		filepath.Join(execDir, "backend", "providers.json"), // backend 目录
		"providers.json",                                  // 当前目录
		"./backend/providers.json",                        // 相对路径
	}

	var config *Config
	for _, path := range configPaths {
		data, err := os.ReadFile(path)
		if err == nil {
			config = &Config{}
			if err := json.Unmarshal(data, config); err == nil {
				return config, nil
			}
		}
	}

	return nil, nil
}

// GetProvidersForCurrentPlatform 获取当前平台可用的 Providers
func (c *Config) GetProvidersForCurrentPlatform() []ProviderConfig {
	currentPlatform := GetCurrentPlatform()
	var result []ProviderConfig

	for _, p := range c.Providers {
		for _, platform := range p.Platforms {
			if platform == currentPlatform {
				result = append(result, p)
				break
			}
		}
	}

	return result
}

// GetProviderByID 根据 ID 获取 Provider 配置
func (c *Config) GetProviderByID(id string) *ProviderConfig {
	for i := range c.Providers {
		if c.Providers[i].ID == id {
			return &c.Providers[i]
		}
	}
	return nil
}

// IsPlatformSupported 检查平台是否支持
func (p *ProviderConfig) IsPlatformSupported(platform string) bool {
	for _, p := range p.Platforms {
		if p == platform {
			return true
		}
	}
	return false
}
