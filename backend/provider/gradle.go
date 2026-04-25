package provider

import (
	"os"
	"os/exec"
	"path/filepath"
)

// gradleProvider Gradle 提供商
type gradleProvider struct{}

func NewGradleProvider() Provider {
	return &gradleProvider{}
}

func (p *gradleProvider) ID() string   { return "gradle" }
func (p *gradleProvider) Name() string { return "Gradle" }

func (p *gradleProvider) Paths() []PathConfig {
	return []PathConfig{
		{
			Path:        "~/.gradle/caches",
			Description: "Gradle 缓存目录",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.gradle/wrapper/dists",
			Description: "Gradle Wrapper 发行版",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%USERPROFILE%\\.gradle\\caches",
			Description: "Gradle 缓存目录 (Windows)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "/root/.gradle/caches",
			Description: "Gradle 缓存目录 (Linux)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Caches/Gradle",
			Description: "Gradle 缓存 (macOS)",
			Strategy:    StrategyDirect,
		},
	}
}

func (p *gradleProvider) Scan() ([]ScanResult, error) {
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

func (p *gradleProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	// 优先使用 gradle 命令清理
	cleaned, err := p.cleanByCommand()
	if err != nil {
		// 命令失败，使用直接删除
		for _, path := range paths {
			cleaned, failed := cleanPathDirect(path)
			result.Cleaned += cleaned
			result.Failed = append(result.Failed, failed...)
		}
	} else {
		result.Cleaned = cleaned
	}

	return result, nil
}

func (p *gradleProvider) cleanByCommand() (int64, error) {
	// 记录清理前的大小
	home, _ := os.UserHomeDir()
	gradlePath := filepath.Join(home, ".gradle")

	var beforeSize int64
	filepath.Walk(gradlePath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			beforeSize += info.Size()
		}
		return nil
	})

	// 使用 gradle cleanBuildCache 命令
	cmd := exec.Command("gradle", "cleanBuildCache", "--no-daemon")
	cmd.Run()

	// 清理 caches 目录（保留 wrapper）
	cleanCachePaths := []string{
		filepath.Join(home, ".gradle", "caches"),
		filepath.Join(home, ".gradle", "daemon"),
	}

	for _, cachePath := range cleanCachePaths {
		if _, err := os.Stat(cachePath); err == nil {
			cleanPathDirect(cachePath)
		}
	}

	// 记录清理后的大小
	var afterSize int64
	filepath.Walk(gradlePath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			afterSize += info.Size()
		}
		return nil
	})

	return beforeSize - afterSize, nil
}
