package provider

import (
	"os"
	"path/filepath"
	"testing"
)

// TestAndroidSDKProviderPaths 测试 Android SDK Provider 路径配置
func TestAndroidSDKProviderPaths(t *testing.T) {
	p := NewAndroidSDKProvider()
	paths := p.Paths()

	if len(paths) == 0 {
		t.Error("AndroidSDKProvider has no paths")
	}

	// 验证关键路径存在
	expectedContains := []string{"android", "sdk", "gradle", "caches"}
	for _, expected := range expectedContains {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) || contains(path.Description, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("AndroidSDKProvider missing path containing %q", expected)
		}
	}
}

// TestCargoProviderPaths 测试 Cargo Provider 路径配置
func TestCargoProviderPaths(t *testing.T) {
	p := NewCargoProvider()
	paths := p.Paths()

	if len(paths) == 0 {
		t.Error("CargoProvider has no paths")
	}

	// 验证关键路径
	expectedContains := []string{"cargo", "registry", "git"}
	for _, expected := range expectedContains {
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

// TestFlutterProviderPaths 测试 Flutter Provider 路径配置
func TestFlutterProviderPaths(t *testing.T) {
	p := NewFlutterProvider()
	paths := p.Paths()

	if len(paths) == 0 {
		t.Error("FlutterProvider has no paths")
	}

	// 验证关键路径
	expectedContains := []string{"pub", "flutter", "dart", "cache"}
	for _, expected := range expectedContains {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) || contains(path.Description, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("FlutterProvider missing path containing %q", expected)
		}
	}
}

// TestNuGetProviderPaths 测试 NuGet Provider 路径配置
func TestNuGetProviderPaths(t *testing.T) {
	p := NewNuGetProvider()
	paths := p.Paths()

	if len(paths) == 0 {
		t.Error("NuGetProvider has no paths")
	}

	// 验证关键路径
	expectedContains := []string{"nuget", "dotnet", "packages"}
	for _, expected := range expectedContains {
		found := false
		for _, path := range paths {
			if contains(path.Path, expected) || contains(path.Description, expected) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("NuGetProvider missing path containing %q", expected)
		}
	}
}

// TestScanMethods 测试各 Provider 的 Scan 方法
func TestScanMethods(t *testing.T) {
	// 创建临时缓存目录进行测试
	tempDir := t.TempDir()

	// 创建模拟缓存文件
	cacheFile := filepath.Join(tempDir, "test-cache.bin")
	if err := os.WriteFile(cacheFile, []byte("test data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		provider Provider
		wantErr  bool
	}{
		{"NPM", NewNPMProvider(), false},
		{"Yarn", NewYarnProvider(), false},
		{"Docker", NewDockerProvider(), false},
		{"Xcode", NewXcodeProvider(), false},
		{"Homebrew", NewHomebrewProvider(), false},
		{"Python", NewPythonProvider(), false},
		{"Go", NewGoProvider(), false},
		{"Ruby", NewRubyProvider(), false},
		{"Maven", NewMavenProvider(), false},
		{"Gradle", NewGradleProvider(), false},
		{"CocoaPods", NewCocoaPodsProvider(), false},
		{"Carthage", NewCarthageProvider(), false},
		{"Unity", NewUnityProvider(), false},
		{"Composer", NewComposerProvider(), false},
		{"Cargo", NewCargoProvider(), false},
		{"Flutter", NewFlutterProvider(), false},
		{"NuGet", NewNuGetProvider(), false},
		{"AndroidSDK", NewAndroidSDKProvider(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := tt.provider.Scan()
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Scan 可能返回 nil 或空切片
			if results == nil {
				t.Logf("Scan() returned nil (no cache found)")
			}
		})
	}
}

// TestCleanMethods 测试各 Provider 的 Clean 方法
func TestCleanMethods(t *testing.T) {
	// 创建临时目录进行清理测试
	tempDir := t.TempDir()

	// 创建测试文件
	testFile := filepath.Join(tempDir, "test-file.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name     string
		provider Provider
	}{
		{"NPM", NewNPMProvider()},
		{"Yarn", NewYarnProvider()},
		{"Docker", NewDockerProvider()},
		{"Xcode", NewXcodeProvider()},
		{"Homebrew", NewHomebrewProvider()},
		{"Python", NewPythonProvider()},
		{"Go", NewGoProvider()},
		{"Ruby", NewRubyProvider()},
		{"Maven", NewMavenProvider()},
		{"Gradle", NewGradleProvider()},
		{"CocoaPods", NewCocoaPodsProvider()},
		{"Carthage", NewCarthageProvider()},
		{"Unity", NewUnityProvider()},
		{"Composer", NewComposerProvider()},
		{"Cargo", NewCargoProvider()},
		{"Flutter", NewFlutterProvider()},
		{"NuGet", NewNuGetProvider()},
		{"AndroidSDK", NewAndroidSDKProvider()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 测试清理功能
			result, err := tt.provider.Clean([]string{tempDir})
			if err != nil {
				t.Logf("Clean() error (may be expected): %v", err)
			}
			if result == nil {
				t.Error("Clean() returned nil")
			}
		})
	}
}

// TestExpandPath 测试路径展开功能
func TestExpandPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input    string
		expected string
	}{
		{"~/.npm", home + "/.npm"},
		{"~/Library/Caches", home + "/Library/Caches"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := expandPath(tt.input)
			if result != tt.expected {
				t.Errorf("expandPath(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestCleanPathDirect 测试直接清理功能
func TestCleanPathDirect(t *testing.T) {
	// 创建临时文件和目录
	tempDir := t.TempDir()

	// 创建测试文件和目录结构
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	files := []string{
		filepath.Join(tempDir, "file1.txt"),
		filepath.Join(tempDir, "file2.txt"),
		filepath.Join(subDir, "nested.txt"),
	}

	var totalSize int64
	for _, f := range files {
		data := []byte("test data")
		if err := os.WriteFile(f, data, 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", f, err)
		}
		totalSize += int64(len(data))
	}

	// 测试清理
	cleaned, failed := cleanPathDirect(tempDir)

	if cleaned == 0 {
		t.Error("Expected some bytes cleaned")
	}

	if len(failed) > 0 {
		t.Errorf("Unexpected failures: %v", failed)
	}

	// 验证文件已删除（目录可能仍存在但内容为空）
	for _, f := range files {
		if _, err := os.Stat(f); !os.IsNotExist(err) {
			t.Errorf("Expected file %s to be removed", f)
		}
	}
}

// TestGetAllProvidersCount 测试所有 Provider 数量
func TestGetAllProvidersCount(t *testing.T) {
	providers := GetAllProviders()

	// 应该有 18 个 Provider
	expected := 18
	if len(providers) != expected {
		t.Errorf("GetAllProviders() returned %d providers, want %d", len(providers), expected)
	}
}

// TestProviderUniqueness 测试 Provider ID 唯一性
func TestProviderUniqueness(t *testing.T) {
	providers := GetAllProviders()
	ids := make(map[string]bool)

	for _, p := range providers {
		id := p.ID()
		if ids[id] {
			t.Errorf("Duplicate provider ID: %s", id)
		}
		ids[id] = true
	}
}

// TestScanResultFields 测试扫描结果字段
func TestScanResultFields(t *testing.T) {
	result := ScanResult{
		Path:        "/test/path",
		Size:        1024,
		FileNum:     10,
		LastMod:     1234567890,
		Description: "Test cache",
	}

	if result.Path != "/test/path" {
		t.Errorf("Path = %q, want %q", result.Path, "/test/path")
	}
	if result.Size != 1024 {
		t.Errorf("Size = %d, want %d", result.Size, 1024)
	}
	if result.FileNum != 10 {
		t.Errorf("FileNum = %d, want %d", result.FileNum, 10)
	}
}

// TestCleanResultFields 测试清理结果字段
func TestCleanResultFields(t *testing.T) {
	result := CleanResult{
		Cleaned:  2048,
		Failed:   []string{},
		FileNum:  5,
	}

	if result.Cleaned != 2048 {
		t.Errorf("Cleaned = %d, want %d", result.Cleaned, 2048)
	}
	if result.FileNum != 5 {
		t.Errorf("FileNum = %d, want %d", result.FileNum, 5)
	}
}
