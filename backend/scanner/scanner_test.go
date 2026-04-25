package scanner

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExpandPath(t *testing.T) {
	// 测试路径展开
	tests := []struct {
		input    string
		contains string
	}{
		{"~/.npm", ".npm"},
		{"~/Library/Caches/npm", "Library/Caches/npm"},
	}

	for _, tt := range tests {
		result := expandPath(tt.input)
		if !contains(result, tt.contains) {
			t.Errorf("expandPath(%q) = %q, want to contain %q", tt.input, result, tt.contains)
		}
	}
}

func TestIsWhitelisted(t *testing.T) {
	whitelist := []string{
		"/Users/test/Documents",
		"/Users/test/Desktop",
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"/Users/test/Documents/important.txt", true},
		{"/Users/test/Desktop/file.txt", true},
		{"/Users/test/Downloads/file.txt", false},
		{"/Users/test/.npm/cache", false},
	}

	for _, tt := range tests {
		result := isWhitelisted(tt.path, whitelist)
		if result != tt.expected {
			t.Errorf("isWhitelisted(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestQuickScan(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	
	// 创建测试文件
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 扫描
	size, err := QuickScan(tmpDir)
	if err != nil {
		t.Fatalf("QuickScan failed: %v", err)
	}

	if size == 0 {
		t.Error("QuickScan returned 0, expected file size")
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1024, "1 KB"},
		{1048576, "1 MB"},
		{1073741824, "1 GB"},
	}

	for _, tt := range tests {
		result := FormatSize(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatSize(%d) = %q, want %q", tt.bytes, result, tt.expected)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
