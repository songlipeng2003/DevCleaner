//go:build darwin
// +build darwin

package provider

import (
	"testing"
)

// TestMacOSOnlyProviders 测试 macOS 专用的 Provider
func TestMacOS_OnlyProviders(t *testing.T) {
	// Xcode Provider - macOS 专用
	t.Run("Xcode", func(t *testing.T) {
		p := NewXcodeProvider()
		if p == nil {
			t.Fatal("XcodeProvider is nil")
		}

		paths := p.Paths()
		if len(paths) == 0 {
			t.Error("XcodeProvider has no paths")
		}

		// 验证包含 macOS 特定路径
		foundMacOSPath := false
		for _, path := range paths {
			if contains(path.Path, "Library") || contains(path.Path, "Xcode") {
				foundMacOSPath = true
				break
			}
		}
		if !foundMacOSPath {
			t.Error("XcodeProvider should contain macOS Library paths")
		}
	})

	// CocoaPods Provider - macOS 专用
	t.Run("CocoaPods", func(t *testing.T) {
		p := NewCocoaPodsProvider()
		if p == nil {
			t.Fatal("CocoaPodsProvider is nil")
		}

		paths := p.Paths()
		if len(paths) == 0 {
			t.Error("CocoaPodsProvider has no paths")
		}
	})

	// Carthage Provider - macOS 专用
	t.Run("Carthage", func(t *testing.T) {
		p := NewCarthageProvider()
		if p == nil {
			t.Fatal("CarthageProvider is nil")
		}

		paths := p.Paths()
		if len(paths) == 0 {
			t.Error("CarthageProvider has no paths")
		}
	})
}

// TestHomebrewMacOSPaths 测试 Homebrew macOS 路径
func TestMacOS_HomebrewPaths(t *testing.T) {
	p := NewHomebrewProvider()
	paths := p.Paths()

	if len(paths) == 0 {
		t.Skip("HomebrewProvider has no paths on this system")
	}

	// macOS 上应该有 Homebrew 缓存路径
	foundCachePath := false
	for _, path := range paths {
		if contains(path.Path, "Homebrew") || contains(path.Path, "brew") || contains(path.Path, "Cellar") {
			foundCachePath = true
			break
		}
	}

	if !foundCachePath {
		t.Log("Note: Homebrew may not be installed or configured")
	}
}

// TestXcodeCachePaths 测试 Xcode 缓存路径
func TestMacOS_XcodeCachePaths(t *testing.T) {
	p := NewXcodeProvider()
	paths := p.Paths()

	expectedPaths := []string{
		"DerivedData",
		"Archives",
		"Xcode",
	}

	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) || contains(path.Description, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("XcodeProvider missing expected path containing %q", expected)
		}
	}
}

// TestMacOSSpecificPaths 测试 macOS 特定路径格式
func TestMacOS_SpecificPaths(t *testing.T) {
	// 验证 ~/Library 路径格式
	testPaths := []struct {
		path        string
		shouldExist string
	}{
		{"~/Library/Caches/npm", "Library"},
		{"~/Library/Caches/Yarn", "Library"},
		{"~/Library/Developer/Xcode", "Xcode"},
	}

	for _, tt := range testPaths {
		expanded := expandPath(tt.path)
		if !contains(expanded, tt.shouldExist) {
			t.Errorf("Path %s should contain %s after expansion", tt.path, tt.shouldExist)
		}
	}
}

// TestPlatformSpecificCachePaths 测试平台特定缓存路径
func TestMacOS_PlatformCachePaths(t *testing.T) {
	// 在 macOS 上测试各种 Provider 的路径
	providers := []struct {
		name     string
		provider Provider
	}{
		{"npm", NewNPMProvider()},
		{"yarn", NewYarnProvider()},
		{"homebrew", NewHomebrewProvider()},
		{"cocoapods", NewCocoaPodsProvider()},
	}

	for _, tt := range providers {
		t.Run(tt.name, func(t *testing.T) {
			paths := tt.provider.Paths()
			if len(paths) == 0 {
				t.Skip("No paths configured")
			}
		})
	}
}
