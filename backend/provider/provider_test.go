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
		{"composer", "composer"},
		{"cargo", "cargo"},
		{"flutter", "flutter"},
		{"nuget", "nuget"},
		{"android_sdk", "android_sdk"},
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
	p := GetProvider("npm")
	if p == nil {
		t.Fatal("NPMProvider is nil")
	}
	
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
	p := GetProvider("yarn")
	if p == nil {
		t.Fatal("YarnProvider is nil")
	}
	
	if p.ID() != "yarn" {
		t.Errorf("YarnProvider.ID() = %q, want %q", p.ID(), "yarn")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("YarnProvider has no paths")
	}
}

func TestDockerProvider(t *testing.T) {
	p := GetProvider("docker")
	if p == nil {
		t.Fatal("DockerProvider is nil")
	}
	
	if p.ID() != "docker" {
		t.Errorf("DockerProvider.ID() = %q, want %q", p.ID(), "docker")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("DockerProvider has no paths")
	}
}

func TestXcodeProvider(t *testing.T) {
	p := GetProvider("xcode")
	if p == nil {
		t.Fatal("XcodeProvider is nil")
	}
	
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
	p := GetProvider("homebrew")
	if p == nil {
		t.Fatal("HomebrewProvider is nil")
	}
	
	if p.ID() != "homebrew" {
		t.Errorf("HomebrewProvider.ID() = %q, want %q", p.ID(), "homebrew")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("HomebrewProvider has no paths")
	}
}

func TestMavenProvider(t *testing.T) {
	p := GetProvider("maven")
	if p == nil {
		t.Fatal("MavenProvider is nil")
	}

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
	p := GetProvider("gradle")
	if p == nil {
		t.Fatal("GradleProvider is nil")
	}

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
	p := GetProvider("cocoapods")
	if p == nil {
		t.Fatal("CocoaPodsProvider is nil")
	}

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
	p := GetProvider("carthage")
	if p == nil {
		t.Fatal("CarthageProvider is nil")
	}

	if p.ID() != "carthage" {
		t.Errorf("CarthageProvider.ID() = %q, want %q", p.ID(), "carthage")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("CarthageProvider has no paths")
	}
}

func TestUnityProvider(t *testing.T) {
	p := GetProvider("unity")
	if p == nil {
		t.Fatal("UnityProvider is nil")
	}

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

func TestComposerProvider(t *testing.T) {
	p := GetProvider("composer")
	if p == nil {
		t.Fatal("ComposerProvider is nil")
	}
	
	if p.ID() != "composer" {
		t.Errorf("ComposerProvider.ID() = %q, want %q", p.ID(), "composer")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("ComposerProvider has no paths")
	}

	// 验证包含关键路径
	expectedPaths := []string{"composer", "cache"}
	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("ComposerProvider missing path containing %q", expected)
		}
	}
}

func TestCargoProvider(t *testing.T) {
	p := GetProvider("cargo")
	if p == nil {
		t.Fatal("CargoProvider is nil")
	}
	
	if p.ID() != "cargo" {
		t.Errorf("CargoProvider.ID() = %q, want %q", p.ID(), "cargo")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("CargoProvider has no paths")
	}

	// 验证包含关键路径
	expectedPaths := []string{"cargo", "registry"}
	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("CargoProvider missing path containing %q", expected)
		}
	}
}

func TestFlutterProvider(t *testing.T) {
	p := GetProvider("flutter")
	if p == nil {
		t.Fatal("FlutterProvider is nil")
	}
	
	if p.ID() != "flutter" {
		t.Errorf("FlutterProvider.ID() = %q, want %q", p.ID(), "flutter")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("FlutterProvider has no paths")
	}

	// 验证包含关键路径
	expectedPaths := []string{"pub", "cache"}
	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("FlutterProvider missing path containing %q", expected)
		}
	}
}

func TestNuGetProvider(t *testing.T) {
	p := GetProvider("nuget")
	if p == nil {
		t.Fatal("NuGetProvider is nil")
	}
	
	if p.ID() != "nuget" {
		t.Errorf("NuGetProvider.ID() = %q, want %q", p.ID(), "nuget")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("NuGetProvider has no paths")
	}

	// 验证包含关键路径
	expectedPaths := []string{"nuget", "packages"}
	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("NuGetProvider missing path containing %q", expected)
		}
	}
}

func TestAndroidSDKProvider(t *testing.T) {
	p := GetProvider("android_sdk")
	if p == nil {
		t.Fatal("AndroidSDKProvider is nil")
	}
	
	if p.ID() != "android_sdk" {
		t.Errorf("AndroidSDKProvider.ID() = %q, want %q", p.ID(), "android_sdk")
	}

	paths := p.Paths()
	if len(paths) == 0 {
		t.Error("AndroidSDKProvider has no paths")
	}

	// 验证包含关键路径
	expectedPaths := []string{"android", "sdk"}
	for _, expected := range expectedPaths {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AndroidSDKProvider missing path containing %q", expected)
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
