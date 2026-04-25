package provider

import (
	"os"
	"os/exec"
	"path/filepath"
)

// mavenProvider Maven 提供商
type mavenProvider struct{}

func NewMavenProvider() Provider {
	return &mavenProvider{}
}

func (p *mavenProvider) ID() string   { return "maven" }
func (p *mavenProvider) Name() string { return "Maven" }

func (p *mavenProvider) Paths() []PathConfig {
	return []PathConfig{
		{
			Path:        "~/.m2/repository",
			Description: "Maven 本地仓库",
			Strategy:    StrategySafe, // 安全删除，需确认
		},
		{
			Path:        "~/.m2/wrapper/dists",
			Description: "Maven Wrapper 发行版",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%USERPROFILE%\\.m2\\repository",
			Description: "Maven 本地仓库 (Windows)",
			Strategy:    StrategySafe,
		},
		{
			Path:        "/root/.m2/repository",
			Description: "Maven 本地仓库 (Linux)",
			Strategy:    StrategySafe,
		},
	}
}

func (p *mavenProvider) Scan() ([]ScanResult, error) {
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

func (p *mavenProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// Maven 仓库使用 mvn 命令清理是最安全的方式
		cleaned, err := p.cleanByCommand()
		if err != nil {
			// 命令失败，尝试直接删除
			cleaned, failed := cleanPathDirect(path)
			result.Cleaned += cleaned
			result.Failed = append(result.Failed, failed...)
		} else {
			result.Cleaned += cleaned
		}
	}

	return result, nil
}

func (p *mavenProvider) cleanByCommand() (int64, error) {
	// 记录清理前的大小
	home, _ := os.UserHomeDir()
	m2Path := filepath.Join(home, ".m2", "repository")
	
	var beforeSize int64
	filepath.Walk(m2Path, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			beforeSize += info.Size()
		}
		return nil
	})

	// 执行 maven dependency:purge-local-repository
	cmd := exec.Command("mvn", "dependency:purge-local-repository", "-Drecursive=true", "-DactTransitively=false")
	cmd.Run()

	// 记录清理后的大小
	var afterSize int64
	filepath.Walk(m2Path, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			afterSize += info.Size()
		}
		return nil
	})

	return beforeSize - afterSize, nil
}
