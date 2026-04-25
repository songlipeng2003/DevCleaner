# DevCleaner

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Tauri](https://img.shields.io/badge/Tauri-2.0-FFC131?logo=tauri&logoColor=white)](https://tauri.app/)
[![Vue.js](https://img.shields.io/badge/Vue.js-3.4-4FC08D?logo=vue.js&logoColor=white)](https://vuejs.org/)
[![Go](https://img.shields.io/badge/Go-1.21-00ADD8?logo=go&logoColor=white)](https://golang.org/)

专为开发者设计的跨平台磁盘清理工具，一站式扫描并清理各类开发工具缓存，释放宝贵的磁盘空间。

## ✨ 特性

- **🛠️ 开发者专属**：深度覆盖 15+ 开发工具，理解开发者工作流
- **🌐 跨平台支持**：原生支持 macOS、Windows 11、Linux
- **🔒 安全优先**：智能白名单策略 + 清理前确认，避免误删重要文件
- **🤖 自动化**：定期扫描 + 阈值通知，保持系统整洁
- **📊 可视化分析**：直观的图表展示磁盘使用情况
- **⚡ 高性能**：Rust + Go 后端，快速扫描和清理
- **🎯 精准清理**：支持按工具、按时间、按大小筛选清理目标

## 📋 支持的工具

| 类别 | 工具 | 清理内容 |
|------|------|----------|
| **包管理器** | npm / yarn / pnpm | 全局缓存、_cacache、下载包 |
| **容器** | Docker | 镜像、容器、卷、构建缓存 |
| **iOS/macOS** | Xcode | DerivedData、Archives、设备支持文件、SPM缓存 |
| **包管理器** | Homebrew | 下载缓存、Cellar、日志 |
| **Python** | pip / conda / venv | pip缓存、__pycache__、虚拟环境 |
| **Go** | Go modules | 模块缓存、构建缓存、测试缓存 |
| **Ruby** | Ruby gems | gem缓存、Bundler缓存、Rails临时文件 |
| **Java** | Maven / Gradle | 本地仓库、Wrapper发行版、构建缓存 |
| **iOS依赖** | CocoaPods / Carthage | 本地仓库、构建缓存 |
| **游戏开发** | Unity | 编辑器缓存、日志、下载缓存 |

## 🚀 快速开始

### 系统要求

- **macOS**: 10.15+ (Catalina 及以上)
- **Windows**: Windows 11 或 Windows 10 (1809+)
- **Linux**: 支持 GTK3 的发行版 (Ubuntu 20.04+, Fedora 33+, 等)
- **内存**: 4GB RAM (推荐 8GB+)
- **磁盘空间**: 200MB 可用空间

### 安装

#### 下载预编译版本

访问 [Releases 页面](https://github.com/yourusername/devcleaner/releases) 下载对应平台的安装包。

#### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/yourusername/devcleaner.git
cd devcleaner

# 安装前端依赖
npm install

# 安装 Rust 工具链 (如果未安装)
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh

# 安装 Go (如果未安装)
# 请参考 https://golang.org/doc/install

# 开发模式运行
npm run tauri dev

# 构建应用
npm run tauri build
```

## 🖥️ 使用方法

### 首次启动

1. 启动 DevCleaner 应用
2. 应用会自动检测系统中已安装的开发工具
3. 点击"开始扫描"按钮，扫描所有支持的缓存目录
4. 查看扫描结果，按工具、大小、最后访问时间排序

### 扫描选项

- **快速扫描**：扫描常用工具的缓存目录
- **深度扫描**：扫描所有支持的工具，包括项目级缓存
- **自定义扫描**：选择特定工具进行扫描

### 清理选项

- **安全清理**：仅清理已知安全的缓存文件
- **深度清理**：清理所有可安全删除的缓存（包括较旧的文件）
- **自定义清理**：手动选择要清理的项目

### 自动化设置

- **定期扫描**：设置每日/每周自动扫描
- **磁盘阈值**：当磁盘空间低于指定阈值时自动提醒
- **排除列表**：添加需要排除的目录或文件

## ⚙️ 配置

### 配置文件位置

- **macOS**: `~/Library/Application Support/devcleaner/config.json`
- **Windows**: `%APPDATA%\devcleaner\config.json`
- **Linux**: `~/.config/devcleaner/config.json`

### 配置示例

```json
{
  "scan": {
    "interval": "weekly",
    "deepScan": false,
    "excludedPaths": [
      "~/projects/important-cache",
      "/usr/local/opt"
    ]
  },
  "clean": {
    "strategy": "safe",
    "keepDays": 30,
    "confirmBeforeClean": true
  },
  "notifications": {
    "enabled": true,
    "thresholdGB": 10
  }
}
```

## 🛠️ 开发指南

### 项目结构

```
devcleaner/
├── src/                    # 前端源码 (Vue 3 + TypeScript)
│   ├── components/        # Vue 组件
│   ├── views/            # 页面视图
│   ├── stores/           # Pinia 状态管理
│   └── assets/           # 静态资源
├── src-tauri/            # Tauri 后端 (Rust)
│   ├── src/              # Rust 源码
│   └── Cargo.toml        # Rust 依赖配置
├── backend/              # Go 后端服务
│   ├── provider/         # 各工具提供商实现
│   ├── scanner/          # 扫描器
│   ├── cleaner/          # 清理器
│   └── go.mod            # Go 模块配置
├── tests/                # 测试文件
└── scripts/              # 构建和部署脚本
```

### 开发环境设置

```bash
# 1. 安装 Node.js (v18+)
# 2. 安装 Go (1.21+)
# 3. 安装 Rust (通过 rustup)
# 4. 安装 Tauri CLI
npm install -g @tauri-apps/cli

# 5. 安装项目依赖
npm install

# 6. 启动开发服务器
npm run tauri dev
```

### 添加新的工具支持

1. 在 `backend/provider/` 目录下创建新的 provider 文件
2. 实现 `Provider` 接口：
   - `ID()` 和 `Name()` 方法
   - `Paths()` 返回要扫描的路径配置
   - `Scan()` 扫描指定路径
   - `Clean()` 清理指定路径
3. 在 `backend/provider/provider.go` 的 `GetProvider()` 和 `GetAllProviders()` 中注册新 provider
4. 在前端添加对应的工具图标和配置

### 运行测试

```bash
# 运行单元测试
npm run test:unit

# 运行端到端测试
npm run test:e2e

# 运行 Go 测试
cd backend && go test ./...

# 运行 Rust 测试
cd src-tauri && cargo test
```

## 🤝 贡献指南

我们欢迎各种形式的贡献！请参考以下步骤：

1. **报告问题**：使用 [GitHub Issues](https://github.com/yourusername/devcleaner/issues) 报告 bug 或提出功能建议
2. **提交 Pull Request**：
   - Fork 本仓库
   - 创建功能分支 (`git checkout -b feature/amazing-feature`)
   - 提交更改 (`git commit -m 'Add some amazing feature'`)
   - 推送到分支 (`git push origin feature/amazing-feature`)
   - 开启 Pull Request

### 开发规范

- 代码风格：遵循各语言的官方代码风格指南
- 提交信息：使用约定式提交 (Conventional Commits)
- 测试：新功能需要包含相应的测试用例
- 文档：更新相关文档和 README

## 📄 许可证

本项目基于 MIT 许可证开源 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🙏 致谢

- [Tauri](https://tauri.app/) - 提供优秀的跨平台桌面应用框架
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Ant Design Vue](https://antdv.com/) - 企业级 UI 组件库
- [Go](https://golang.org/) - 简单高效的编程语言
- [Rust](https://www.rust-lang.org/) - 赋予每个人构建可靠高效软件的能力

## 📞 支持与反馈

- **GitHub Issues**: [报告问题或请求功能](https://github.com/yourusername/devcleaner/issues)
- **讨论区**: [加入讨论](https://github.com/yourusername/devcleaner/discussions)
- **电子邮件**: team@devcleaner.app (示例)

---

<p align="center">
  Made with ❤️ for developers everywhere
</p>
