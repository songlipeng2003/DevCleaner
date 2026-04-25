package provider

import (
	"os"
	"os/exec"
	"path/filepath"
)

// cocoapodsProvider CocoaPods 提供商
type cocoapodsProvider struct{}

func NewCocoaPodsProvider() Provider {
	return &cocoapodsProvider{}
}

func (p *cocoapodsProvider) ID() string   { return "cocoapods" }
func (p *cocoapodsProvider) Name() string { return "CocoaPods" }

func (p *cocoapodsProvider) Paths() []PathConfig {
	return []PathConfig{
		{
			Path:        "~/Library/Caches/CocoaPods",
			Description: "CocoaPods 缓存 (macOS)",
			Strategy:    StrategyCommand, // 使用 pod 命令清理
		},
		{
			Path:        "~/.cocoapods",
			Description: "CocoaPods 本地仓库",
			Strategy:    StrategySafe, // 安全删除
		},
		{
			Path:        "~/Library/Developer/Xcode/DerivedData",
			Description: "Xcode DerivedData (Pods 构建产物)",
			Strategy:    StrategySafe,
		},
	}
}

func (p *cocoapodsProvider) Scan() ([]ScanResult, error) {
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

func (p *cocoapodsProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	// 尝试使用 pod 命令清理
	cleaned, err := p.cleanByCommand()
	if err != nil {
		// 命令失败，直接删除缓存目录
		cachePath := expandPath("~/Library/Caches/CocoaPods")
		cleaned, failed := cleanPathDirect(cachePath)
		result.Cleaned += cleaned
		result.Failed = append(result.Failed, failed...)
	} else {
		result.Cleaned += cleaned
	}

	// CocoaPods 本地仓库使用 pod 命令清理
	home, _ := os.UserHomeDir()
	trunkRepo := filepath.Join(home, ".cocoapods", "repos", "trunk")
	if _, err := os.Stat(trunkRepo); err == nil {
		// 不直接删除 repos，仅清理临时文件
		p.cleanRepos(trunkRepo)
	}

	return result, nil
}

func (p *cocoapodsProvider) cleanByCommand() (int64, error) {
	// 记录清理前的大小
	cachePath := expandPath("~/Library/Caches/CocoaPods")

	var beforeSize int64
	filepath.Walk(cachePath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			beforeSize += info.Size()
		}
		return nil
	})

	// 执行 pod cache clean --all
	cmd := exec.Command("pod", "cache", "clean", "--all")
	cmd.Run()

	// 记录清理后的大小
	var afterSize int64
	filepath.Walk(cachePath, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			afterSize += info.Size()
		}
		return nil
	})

	return beforeSize - afterSize, nil
}

func (p *cocoapodsProvider) cleanRepos(repoPath string) {
	// 清理 repos 中的临时文件和过期的缓存
	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// 清理 Checksums 等临时文件
		if info.Name() == "Checksums" || info.Name() == ".tmp" {
			os.RemoveAll(path)
		}

		return nil
	})
}
