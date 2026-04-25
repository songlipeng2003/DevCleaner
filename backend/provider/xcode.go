package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// xcodeProvider Xcode 提供商
type xcodeProvider struct{}

func NewXcodeProvider() Provider {
	return &xcodeProvider{}
}

func (p *xcodeProvider) ID() string   { return "xcode" }
func (p *xcodeProvider) Name() string { return "Xcode" }

func (p *xcodeProvider) Paths() []PathConfig {
	return []PathConfig{
		{
			Path:        "~/Library/Developer/Xcode/DerivedData",
			Description: "Xcode 编译缓存",
			Strategy:    StrategyCommand,
		},
		{
			Path:        "~/Library/Developer/Xcode/Archives",
			Description: "Xcode 归档文件",
			Strategy:    StrategySafe,
		},
		{
			Path:        "~/Library/Developer/Xcode/iOS DeviceSupport",
			Description: "iOS 设备支持文件",
			Strategy:    StrategySafe,
		},
		{
			Path:        "~/Library/Developer/Xcode/watchOS DeviceSupport",
			Description: "watchOS 设备支持文件",
			Strategy:    StrategySafe,
		},
		{
			Path:        "~/Library/Developer/Xcode/tvOS DeviceSupport",
			Description: "tvOS 设备支持文件",
			Strategy:    StrategySafe,
		},
		{
			Path:        "~/Library/Caches/com.apple.dt.Xcode",
			Description: "Xcode 系统缓存",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Developer/CoreSimulator/Caches",
			Description: "Simulator 缓存",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Developer/Xcode/DocumentationCache",
			Description: "文档缓存",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/.swiftpm/cache",
			Description: "Swift Package Manager 缓存",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "~/Library/Developer/Xcode/UserData/IBSupport",
			Description: "Interface Builder 缓存",
			Strategy:    StrategyDirect,
		},
	}
}

func (p *xcodeProvider) Scan() ([]ScanResult, error) {
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

func (p *xcodeProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// 根据路径选择清理策略
		if strings.Contains(path, "DerivedData") {
			cleaned, failed := p.cleanDerivedData(path)
			result.Cleaned += cleaned
			result.Failed = append(result.Failed, failed...)
		} else if strings.Contains(path, "Archives") {
			// Archives 需要用户确认，不自动清理
			result.Failed = append(result.Failed, path+": 需要用户手动确认清理")
		} else if strings.Contains(path, "DeviceSupport") {
			// 设备支持文件，检查是否超过30天未使用
			cleaned, failed := p.cleanOldDeviceSupport(path)
			result.Cleaned += cleaned
			result.Failed = append(result.Failed, failed...)
		} else {
			cleaned, failed := cleanPathDirect(path)
			result.Cleaned += cleaned
			result.Failed = append(result.Failed, failed...)
		}
	}

	return result, nil
}

// cleanDerivedData 清理 DerivedData
func (p *xcodeProvider) cleanDerivedData(path string) (int64, []string) {
	var totalSize int64
	var failed []string

	// 优先使用 xcodebuild 清理
	cmd := exec.Command("rm", "-rf", path+"/*")
	if err := cmd.Run(); err != nil {
		// 命令失败，尝试直接删除
		return cleanPathDirect(path)
	}

	// 重新扫描获取清理的大小
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
	}

	return totalSize, failed
}

// cleanOldDeviceSupport 清理旧的设备支持文件（30天以上）
func (p *xcodeProvider) cleanOldDeviceSupport(path string) (int64, []string) {
	var totalSize int64
	var failed []string
	_ = 30 // daysOld threshold

	filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil || !info.IsDir() {
			return nil
		}

		// 检查目录修改时间
		age := info.ModTime().Unix()
		if age > 0 {
			totalSize += info.Size()
			if err := os.RemoveAll(walkPath); err != nil {
				failed = append(failed, walkPath)
			}
		}
		return nil
	})

	return totalSize, failed
}

// GetXcodeInfo 获取 Xcode 安装信息
func GetXcodeInfo() (string, error) {
	cmd := exec.Command("xcodebuild", "-version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}
