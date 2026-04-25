//go:build linux
// +build linux

package provider

import (
	"testing"
)

// TestLinuxSpecificProviders 测试 Linux 特定的 Provider 行为
func TestLinux_SpecificProviders(t *testing.T) {
	// Homebrew on Linux
	t.Run("Homebrew", func(t *testing.T) {
		p := NewHomebrewProvider()
		if p == nil {
			t.Fatal("HomebrewProvider is nil")
		}

		paths := p.Paths()
		if len(paths) == 0 {
			t.Skip("Homebrew not installed on this Linux system")
		}

		// 验证包含 Linux Homebrew 路径
		foundLinuxPath := false
		for _, path := range paths {
			if contains(path.Path, "linuxbrew") || contains(path.Path, ".linuxbrew") {
				foundLinuxPath = true
				break
			}
		}
		if !foundLinuxPath {
			t.Log("Note: Linuxbrew may not be installed")
		}
	})
}

// TestLinuxPathFormats 测试 Linux 路径格式
func TestLinux_PathFormats(t *testing.T) {
	testCases := []struct {
		name     string
		path     string
		expected string
	}{
		{"Home directory", "~", "/"},
		{".cache directory", "~/.cache", ".cache"},
		{".config directory", "~/.config", ".config"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expanded := expandPath(tc.path)
			if expanded == "" {
				t.Errorf("expandPath(%q) returned empty string", tc.path)
			}
			if !contains(expanded, tc.expected) {
				t.Errorf("Expanded path %q should contain %q", expanded, tc.expected)
			}
		})
	}
}

// TestLinuxCachePaths 测试 Linux 缓存路径
func TestLinux_CachePaths(t *testing.T) {
	providers := []struct {
		name     string
		provider Provider
	}{
		{"npm", NewNPMProvider()},
		{"yarn", NewYarnProvider()},
		{"python", NewPythonProvider()},
		{"go", NewGoProvider()},
	}

	for _, tt := range providers {
		t.Run(tt.name, func(t *testing.T) {
			paths := tt.provider.Paths()
			if len(paths) == 0 {
				t.Skip("No paths configured")
			}

			// 验证至少有 Linux 兼容的路径
			foundLinuxPath := false
			for _, path := range paths {
				if contains(path.Path, "~/.cache") ||
					contains(path.Path, "~") ||
					contains(path.Path, "/root") {
					foundLinuxPath = true
					break
				}
			}
			if !foundLinuxPath {
				t.Logf("Warning: Provider %s may not have Linux-specific paths", tt.name)
			}
		})
	}
}

// TestGradleLinuxPaths 测试 Gradle Linux 路径
func TestLinux_GradlePaths(t *testing.T) {
	p := NewGradleProvider()
	paths := p.Paths()

	// 验证包含 Linux 路径
	foundLinuxPath := false
	for _, path := range paths {
		if contains(path.Path, "/root/.gradle") ||
			contains(path.Path, "~/.gradle") {
			foundLinuxPath = true
			break
		}
	}

	if !foundLinuxPath {
		t.Error("GradleProvider should contain Linux-specific paths")
	}
}

// TestMavenLinuxPaths 测试 Maven Linux 路径
func TestLinux_MavenPaths(t *testing.T) {
	p := NewMavenProvider()
	paths := p.Paths()

	// 验证包含 Linux 路径
	foundLinuxPath := false
	for _, path := range paths {
		if contains(path.Path, "/root/.m2") ||
			contains(path.Path, "~/.m2") {
			foundLinuxPath = true
			break
		}
	}

	if !foundLinuxPath {
		t.Error("MavenProvider should contain Linux-specific paths")
	}
}

// TestUnityLinuxPaths 测试 Unity Linux 路径
func TestLinux_UnityPaths(t *testing.T) {
	p := NewUnityProvider()
	paths := p.Paths()

	// 验证包含 Linux 路径
	foundLinuxPath := false
	for _, path := range paths {
		if contains(path.Path, ".config/unity3d") ||
			contains(path.Path, ".cache/Unity") ||
			contains(path.Path, ".local/share/unity3d") {
			foundLinuxPath = true
			break
		}
	}

	if !foundLinuxPath {
		t.Error("UnityProvider should contain Linux-specific paths")
	}
}

// TestLinuxFlutterPaths 测试 Flutter Linux 路径
func TestLinux_FlutterPaths(t *testing.T) {
	p := NewFlutterProvider()
	paths := p.Paths()

	// 验证包含 Linux Android SDK 路径
	foundLinuxPath := false
	for _, path := range paths {
		if contains(path.Path, ".android/avd") ||
			contains(path.Path, "~/.android") {
			foundLinuxPath = true
			break
		}
	}

	if !foundLinuxPath {
		t.Log("Note: Flutter may not be installed with Android SDK support")
	}
}

// TestLinuxCargoPaths 测试 Cargo Linux 路径
func TestLinux_CargoPaths(t *testing.T) {
	p := NewCargoProvider()
	paths := p.Paths()

	expectedPaths := []string{"cargo", "registry", "git"}

	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("CargoProvider missing expected path containing %q", expected)
		}
	}
}

// TestLinuxRubyPaths 测试 Ruby Linux 路径
func TestLinux_RubyPaths(t *testing.T) {
	p := NewRubyProvider()
	paths := p.Paths()

	// 验证包含 Linux Ruby 路径
	foundLinuxPath := false
	for _, path := range paths {
		if contains(path.Path, ".gem") ||
			contains(path.Path, ".cache/bundler") ||
			contains(path.Path, ".rvm") ||
			contains(path.Path, ".rbenv") {
			foundLinuxPath = true
			break
		}
	}

	if !foundLinuxPath {
		t.Error("RubyProvider should contain Linux-specific paths")
	}
}
