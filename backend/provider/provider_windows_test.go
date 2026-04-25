//go:build windows
// +build windows

package provider

import (
	"testing"
)

// TestWindowsOnlyProviders 测试 Windows 专用或主要支持的 Provider
func TestWindows_OnlyProviders(t *testing.T) {
	// NuGet 主要在 Windows 上使用，但也支持其他平台
	t.Run("NuGet", func(t *testing.T) {
		p := NewNuGetProvider()
		if p == nil {
			t.Fatal("NuGetProvider is nil")
		}

		paths := p.Paths()
		if len(paths) == 0 {
			t.Error("NuGetProvider has no paths")
		}

		// 验证包含 Windows 特定路径
		foundWindowsPath := false
		for _, path := range paths {
			if contains(path.Path, "USERPROFILE") ||
				contains(path.Path, "APPDATA") ||
				contains(path.Path, "nuget") {
				foundWindowsPath = true
				break
			}
		}
		if !foundWindowsPath {
			t.Error("NuGetProvider should contain Windows-specific paths")
		}
	})
}

// TestWindowsPathFormats 测试 Windows 路径格式
func TestWindows_PathFormats(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "NuGet packages",
			path:     "%USERPROFILE%\\.nuget\\packages",
			expected: []string{"nuget", "packages"},
		},
		{
			name:     "NuGet cache",
			path:     "%APPDATA%\\NuGet\\Cache",
			expected: []string{"NuGet", "Cache"},
		},
		{
			name:     "pip cache",
			path:     "%LOCALAPPDATA%\\pip\\cache",
			expected: []string{"pip", "cache"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandPath(tc.path)
			for _, exp := range tc.expected {
				if !contains(expanded, exp) {
					t.Errorf("Expanded path %q should contain %q", expanded, exp)
				}
			}
		})
	}
}

// TestWindowsEnvironmentVariables 测试 Windows 环境变量展开
func TestWindows_EnvironmentVariables(t *testing.T) {
	testCases := []string{
		"%USERPROFILE%",
		"%APPDATA%",
		"%LOCALAPPDATA%",
		"%TEMP%",
		"%TMP%",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			expanded := expandPath(tc)
			if expanded == "" {
				t.Errorf("expandPath(%q) returned empty string", tc)
			}
			// 验证环境变量已被展开
			if contains(expanded, "%") {
				t.Errorf("expandPath(%q) did not expand environment variable, got %q", tc, expanded)
			}
		})
	}
}

// TestWindowsProviderPaths 测试 Windows Provider 路径
func TestWindows_ProviderPaths(t *testing.T) {
	providers := []struct {
		name     string
		provider Provider
	}{
		{"npm", NewNPMProvider()},
		{"yarn", NewYarnProvider()},
		{"nuget", NewNuGetProvider()},
		{"python", NewPythonProvider()},
	}

	for _, tt := range providers {
		t.Run(tt.name, func(t *testing.T) {
			paths := tt.provider.Paths()
			if len(paths) == 0 {
				t.Skip("No paths configured")
			}

			// 验证至少有 Windows 兼容的路径
			foundWindowsPath := false
			for _, path := range paths {
				if contains(path.Path, "%USERPROFILE%") ||
					contains(path.Path, "%APPDATA%") ||
					contains(path.Path, "%LOCALAPPDATA%") ||
					contains(path.Path, "~") {
					foundWindowsPath = true
					break
				}
			}
			if !foundWindowsPath {
				t.Logf("Warning: Provider %s may not have Windows-specific paths", tt.name)
			}
		})
	}
}

// TestGradleWindowsPaths 测试 Gradle Windows 路径
func TestWindows_GradlePaths(t *testing.T) {
	p := NewGradleProvider()
	paths := p.Paths()

	// 验证包含 Windows 路径
	foundWindowsPath := false
	for _, path := range paths {
		if contains(path.Path, "AppData") || contains(path.Path, "Users") {
			foundWindowsPath = true
			break
		}
	}

	if !foundWindowsPath {
		t.Error("GradleProvider should contain Windows-specific paths")
	}
}

// TestMavenWindowsPaths 测试 Maven Windows 路径
func TestWindows_MavenPaths(t *testing.T) {
	p := NewMavenProvider()
	paths := p.Paths()

	// 验证包含 Windows 路径
	foundWindowsPath := false
	for _, path := range paths {
		if contains(path.Path, "AppData") || contains(path.Path, "Users") {
			foundWindowsPath = true
			break
		}
	}

	if !foundWindowsPath {
		t.Error("MavenProvider should contain Windows-specific paths")
	}
}

// TestUnityWindowsPaths 测试 Unity Windows 路径
func TestWindows_UnityPaths(t *testing.T) {
	p := NewUnityProvider()
	paths := p.Paths()

	// 验证包含 Windows 路径
	foundWindowsPath := false
	for _, path := range paths {
		if contains(path.Path, "AppData") || contains(path.Path, "Unity") {
			foundWindowsPath = true
			break
		}
	}

	if !foundWindowsPath {
		t.Error("UnityProvider should contain Windows-specific paths")
	}
}
