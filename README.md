# DevCleaner

专为开发者设计的跨平台磁盘清理工具

## 产品定位

- **一句话描述**：专为开发者设计的跨平台磁盘清理工具，一站式扫描并清理各类开发工具缓存
- **核心价值**：
  - 开发者专属：深度覆盖 15+ 开发工具
  - 跨平台：macOS、Windows 11、Linux
  - 安全优先：白名单策略 + 清理前确认
  - 自动化：定期扫描 + 阈值通知

## 技术栈

| 维度 | 决策 |
|------|------|
| 桌面框架 | Tauri (Rust) |
| 前端框架 | Vue 3 + TypeScript |
| UI 组件库 | Ant Design Vue |
| 构建工具 | Vite |
| 后端服务 | Go |
| 平台 | macOS（V1）→ Windows 11 → Linux |

## 支持的开发工具

- npm / yarn / pnpm
- Docker
- Xcode (DerivedData, Archives, SPM)
- Homebrew
- Python (pip, conda, venv)
- Go modules
- Ruby gems
- Maven / Gradle
- CocoaPods
- Carthage
- Unity

## 快速开始

```bash
# 安装依赖
npm install

# 开发模式
npm run tauri dev

# 构建
npm run tauri build
```

## 许可证

MIT License
