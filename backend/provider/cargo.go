package provider

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"github.com/devcleaner/backend/scanner"
)

// cargoProvider Cargo 提供商
type cargoProvider struct{}

func NewCargoProvider() Provider {
	return &cargoProvider{}
}

func (p *cargoProvider) ID() string   { return "cargo" }
func (p *cargoProvider) Name() string { return "Cargo" }

func (p *cargoProvider) Paths() []PathConfig {
	return []PathConfig{
		// Cargo 全局缓存（registry）
		{
			Path:        "~/.cargo/registry",
			Description: "Cargo 注册表缓存 (Linux/macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%USERPROFILE%\\.cargo\\registry",
			Description: "Cargo 注册表缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// Cargo git 依赖缓存
		{
			Path:        "~/.cargo/git",
			Description: "Cargo Git 依赖缓存 (Linux/macOS)",
			Strategy:    StrategyDirect,
		},
		{
			Path:        "%USERPROFILE%\\.cargo\\git",
			Description: "Cargo Git 依赖缓存 (Windows)",
			Strategy:    StrategyDirect,
		},
		// Cargo 构建缓存
		{
			Path:        "**/target",
			Description: "Cargo 构建输出目录",
			Strategy:    StrategySafe, // 需谨慎，可能包含重要构建产物
		},
		// Cargo 本地安装缓存
		{
			Path:        "~/.cargo/bin",
			Description: "Cargo 安装的可执行文件",
			Strategy:    StrategySafe,
		},
		// Cargo 配置缓存
		{
			Path:        "~/.cargo/.crates.toml",
			Description: "Cargo 已安装 crate 记录",
			Strategy:    StrategySafe,
		},
		// Cargo 日志
		{
			Path:        "~/.cargo/.crates2.json",
			Description: "Cargo 已安装 crate 记录（新格式）",
			Strategy:    StrategySafe,
		},
	}
}

func (p *cargoProvider) Scan() ([]ScanResult, error) {
	var results []ScanResult

	// 扫描标准路径
	paths := []struct {
		path        string
		description string
	}{
		{expandPath("~/.cargo/registry"), "Cargo 注册表缓存"},
		{expandPath("~/.cargo/git"), "Cargo Git 依赖缓存"},
	}

	for _, pathItem := range paths {
		if result, ok := p.scanSinglePath(pathItem.path, pathItem.description); ok {
			results = append(results, result)
		}
	}

	// 扫描 target 目录（限制深度，避免扫描整个文件系统）
	home, _ := os.UserHomeDir()
	targetPattern := filepath.Join(home, "**", "target")
	matches, _ := filepath.Glob(targetPattern)
	if matches != nil {
		for _, targetPath := range matches {
			// 跳过某些常见项目的 target 目录
			if strings.Contains(targetPath, "/.rustup/") || strings.Contains(targetPath, "/.cargo/") {
				continue
			}
			if result, ok := p.scanSinglePath(targetPath, "Cargo 构建输出目录"); ok {
				results = append(results, result)
			}
		}
	}

	return results, nil
}

func (p *cargoProvider) scanSinglePath(path, description string) (ScanResult, bool) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return ScanResult{}, false
	} else if err != nil {
		return ScanResult{}, false
	}

	var totalSize int64
	var fileCount int
	var lastMod int64

	filepath.Walk(path, func(walkPath string, fileInfo os.FileInfo, err error) error {
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
		return ScanResult{
			Path:        path,
			Size:        totalSize,
			FileNum:     fileCount,
			LastMod:     lastMod,
			Description: description,
		}, true
	}

	return ScanResult{}, false
}

func (p *cargoProvider) Clean(paths []string) (*CleanResult, error) {
	result := &CleanResult{
		Failed: []string{},
	}

	for _, path := range paths {
		// 检查是否是 target 目录，需谨慎
		if strings.Contains(path, "target") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 target 目录，请使用 `cargo clean`")
			continue
		}
		// 检查是否是 bin 目录或配置文件
		if strings.Contains(path, ".crates") || strings.Contains(path, "/bin/") {
			result.Failed = append(result.Failed, path+": 不建议自动清理 Cargo 安装文件")
			continue
		}
		// 其他缓存可以安全清理
		cleaned, failed := cleanPathDirect(path)
		result.Cleaned += cleaned
		result.Failed = append(result.Failed, failed...)
	}

	// 运行 cargo cache 清理命令（如果安装了 cargo-cache 工具）
	p.cleanCache()

	return result, nil
}

func (p *cargoProvider) cleanCache() {
	// 尝试使用 cargo-cache 工具
	cmd := exec.Command("cargo", "cache", "--autoclean")
	cmd.Run() // 忽略错误

	// 如果没有 cargo-cache，尝试使用 cargo clean 清理全局缓存（无标准命令）
	// 可以使用 `cargo clean --release` 但只对当前项目有效
	// 这里暂时不执行
}

// GetCargoInfo 获取 Cargo 环境信息
func GetCargoInfo() (map[string]string, error) {
	info := make(map[string]string)

	// Cargo 版本
	cmd := exec.Command("cargo", "--version")
	if output, err := cmd.Output(); err == nil {
		info["cargo_version"] = strings.TrimSpace(string(output))
	}

	// Rustc 版本
	cmd = exec.Command("rustc", "--version")
	if output, err := cmd.Output(); err == nil {
		info["rustc_version"] = strings.TrimSpace(string(output))
	}

	// 缓存目录大小估算
	registryPath := expandPath("~/.cargo/registry")
	if stat, err := os.Stat(registryPath); err == nil && stat.IsDir() {
		var totalSize int64
		filepath.Walk(registryPath, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				totalSize += info.Size()
			}
			return nil
		})
		info["registry_cache_size"] = scanner.FormatSize(totalSize)
	}

	return info, nil
}

