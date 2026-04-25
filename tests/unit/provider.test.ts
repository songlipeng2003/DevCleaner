package provider

import (
	"testing"
)

func TestProviderInterface(t *testing.T) {
	providers := GetAllProviders()
	
	if len(providers) == 0 {
		t.Error("GetAllProviders returned empty list")
	}

	for _, p := range providers {
		if p.ID() == "" {
			t.Error("Provider ID is empty")
		}
		if p.Name() == "" {
			t.Error("Provider Name is empty")
		}
		if len(p.Paths()) == 0 {
			t.Errorf("Provider %s has no paths", p.ID())
		}
	}
}

func TestGetProvider(t *testing.T) {
	tests := []struct {
		id       string
		expected string
	}{
		{"npm", "npm"},
		{"yarn", "yarn"},
		{"docker", "docker"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		p := GetProvider(tt.id)
		if tt.expected == "" {
			if p != nil {
				t.Errorf("GetProvider(%q) = %v, want nil", tt.id, p)
			}
		} else {
			if p == nil {
				t.Errorf("GetProvider(%q) = nil, want non-nil", tt.id)
			} else if p.ID() != tt.expected {
				t.Errorf("GetProvider(%q).ID() = %q, want %q", tt.id, p.ID(), tt.expected)
			}
		}
	}
}

func TestNPMProvider(t *testing.T) {
	p := NewNPMProvider()
	
	if p.ID() != "npm" {
		t.Errorf("NPMProvider.ID() = %q, want %q", p.ID(), "npm")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("NPMProvider has no paths")
	}

	// 检查路径包含 npm
	for _, path := range paths {
		if path.Path == "" {
			t.Error("NPMProvider path is empty")
		}
	}
}

func TestYarnProvider(t *testing.T) {
	p := NewYarnProvider()
	
	if p.ID() != "yarn" {
		t.Errorf("YarnProvider.ID() = %q, want %q", p.ID(), "yarn")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("YarnProvider has no paths")
	}
}

func TestDockerProvider(t *testing.T) {
	p := NewDockerProvider()
	
	if p.ID() != "docker" {
		t.Errorf("DockerProvider.ID() = %q, want %q", p.ID(), "docker")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("DockerProvider has no paths")
	}
}

func TestXcodeProvider(t *testing.T) {
	p := NewXcodeProvider()
	
	if p.ID() != "xcode" {
		t.Errorf("XcodeProvider.ID() = %q, want %q", p.ID(), "xcode")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("XcodeProvider has no paths")
	}

	// 验证包含关键路径
	expectedPaths := []string{"DerivedData", "Archives"}
	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("XcodeProvider missing path containing %q", expected)
		}
	}
}

func TestHomebrewProvider(t *testing.T) {
	p := NewHomebrewProvider()
	
	if p.ID() != "homebrew" {
		t.Errorf("HomebrewProvider.ID() = %q, want %q", p.ID(), "homebrew")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("HomebrewProvider has no paths")
	}
}

func TestMavenProvider(t *testing.T) {
	p := NewMavenProvider()

	if p.ID() != "maven" {
		t.Errorf("MavenProvider.ID() = %q, want %q", p.ID(), "maven")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("MavenProvider has no paths")
	}

	// 验证包含关键路径
	expectedPaths := []string{".m2/repository"}
	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("MavenProvider missing path containing %q", expected)
		}
	}
}

func TestGradleProvider(t *testing.T) {
	p := NewGradleProvider()

	if p.ID() != "gradle" {
		t.Errorf("GradleProvider.ID() = %q, want %q", p.ID(), "gradle")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("GradleProvider has no paths")
	}

	// 验证包含关键路径
	expectedPaths := []string{".gradle/caches"}
	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("GradleProvider missing path containing %q", expected)
		}
	}
}

func TestCocoaPodsProvider(t *testing.T) {
	p := NewCocoaPodsProvider()

	if p.ID() != "cocoapods" {
		t.Errorf("CocoaPodsProvider.ID() = %q, want %q", p.ID(), "cocoapods")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("CocoaPodsProvider has no paths")
	}

	// 验证包含关键路径
	expectedPaths := []string{"CocoaPods"}
	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("CocoaPodsProvider missing path containing %q", expected)
		}
	}
}

func TestCarthageProvider(t *testing.T) {
	p := NewCarthageProvider()

	if p.ID() != "carthage" {
		t.Errorf("CarthageProvider.ID() = %q, want %q", p.ID(), "carthage")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("CarthageProvider has no paths")
	}
}

func TestUnityProvider(t *testing.T) {
	p := NewUnityProvider()

	if p.ID() != "unity" {
		t.Errorf("UnityProvider.ID() = %q, want %q", p.ID(), "unity")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("UnityProvider has no paths")
	}

	// 验证包含关键路径
	expectedPaths := []string{"Unity"}
	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("UnityProvider missing path containing %q", expected)
		}
	}
}

func TestCleanStrategy(t *testing.T) {
	if StrategyDirect != 0 {
		t.Errorf("StrategyDirect = %d, want 0", StrategyDirect)
	}
	if StrategyCommand != 1 {
		t.Errorf("StrategyCommand = %d, want 1", StrategyCommand)
	}
	if StrategySafe != 2 {
		t.Errorf("StrategySafe = %d, want 2", StrategySafe)
	}
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
