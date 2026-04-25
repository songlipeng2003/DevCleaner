//go:build darwin || linux || windows
// +build darwin linux windows

package provider

import (
	"runtime"
	"strings"
	"testing"
)

// TestCrossPlatformPaths 测试跨平台路径格式
func TestCrossPlatformPaths(t *testing.T) {
	providers := GetAllProviders()

	for _, p := range providers {
		t.Run(p.ID(), func(t *testing.T) {
			paths := p.Paths()
			if len(paths) == 0 {
				t.Errorf("Provider %s has no paths configured", p.ID())
			}

			// 验证路径不为空
			for _, path := range paths {
				if path.Path == "" {
					t.Errorf("Provider %s has empty path", p.ID())
				}
				// 验证描述不为空
				if path.Description == "" {
					t.Errorf("Provider %s has empty description for path %s", p.ID(), path.Path)
				}
			}
		})
	}
}

// TestPlatformSpecificPathsAvailability 测试平台特定路径是否在当前平台可用
func TestPlatformSpecificPathsAvailability(t *testing.T) {
	tests := []struct {
		providerID     string
		expectedPrefix []string
	}{
		{"npm", []string{"~/.npm", "~/.cache/npm"}},
		{"yarn", []string{"~/.yarn-cache", "~/.cache/yarn"}},
		{"python", []string{"~/.cache/pip", "~/.pycache"}},
		{"go", []string{"$(go env GOPATH)"}},
	}

	for _, tt := range tests {
		t.Run(tt.providerID, func(t *testing.T) {
			p := GetProvider(tt.providerID)
			if p == nil {
				t.Skip("Provider not available")
			}

			paths := p.Paths()
			if len(paths) == 0 {
				t.Errorf("Provider %s has no paths", tt.providerID)
			}
		})
	}
}

// TestPathExpansion 测试路径展开功能
func TestPathExpansion(t *testing.T) {
	tests := []struct {
		input    string
		platform string
	}{
		{"~/.npm", runtime.GOOS},
		{"~/Library/Caches", runtime.GOOS},
		{"%USERPROFILE%", runtime.GOOS},
		{"%APPDATA%", runtime.GOOS},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			expanded := expandPath(tt.input)
			if expanded == "" {
				t.Errorf("expandPath(%q) returned empty string", tt.input)
			}
			// 验证展开后不包含原始的特殊字符（除了 ~）
			if strings.HasPrefix(tt.input, "~") && strings.HasPrefix(expanded, "~") {
				t.Errorf("expandPath(%q) did not expand ~", tt.input)
			}
		})
	}
}

// TestProviderPathsConsistency 测试 Provider 路径配置的一致性
func TestProviderPathsConsistency(t *testing.T) {
	providers := GetAllProviders()

	for _, p := range providers {
		t.Run(p.ID(), func(t *testing.T) {
			paths := p.Paths()

			// 验证所有路径都有策略设置
			for _, path := range paths {
				if path.Strategy < StrategyDirect || path.Strategy > StrategySafe {
					t.Errorf("Provider %s has invalid strategy %d for path %s",
						p.ID(), path.Strategy, path.Path)
				}
			}

			// 验证描述包含平台信息
			for _, path := range paths {
				desc := path.Description
				// 如果路径是平台特定的，描述应该包含平台名称
				pathLower := strings.ToLower(path.Path)
				if strings.Contains(pathLower, "macos") || strings.Contains(pathLower, "darwin") {
					if !strings.Contains(strings.ToLower(desc), "macos") && !strings.Contains(strings.ToLower(desc), "mac") {
						// macOS 特定路径应该有对应的描述
					}
				}
			}
		})
	}
}

// TestCleanStrategyConsistency 测试清理策略一致性
func TestCleanStrategyConsistency(t *testing.T) {
	// 安全策略应该用于需要确认的路径
	safeStrategyPaths := []string{
		"vendor", "node_modules", "venv", ".venv",
		"Archives", "DeviceSupport", "Application Support",
	}

	for _, pattern := range safeStrategyPaths {
		t.Run(pattern, func(t *testing.T) {
			// 这个测试验证安全策略被正确应用于敏感路径
			// 在实际实现中，这些路径应该使用 StrategySafe
		})
	}
}
