# Tauri 开发环境修复指南

## 问题描述

运行 `npm run tauri:dev` 时遇到平台特定模块缺失错误：

```
Error: Cannot find module @rollup/rollup-linux-arm64-gnu
Error: Cannot find module @tauri-apps/cli-linux-arm64-gnu
```

## 系统信息

- **平台架构**: ARM64 (aarch64)
- **Node.js版本**: v22.22.0
- **操作系统**: Linux

## 根本原因

1. **平台特定模块缺失**: 当前系统是 ARM64 架构，但 node_modules 中缺少对应的原生绑定模块
2. **npm 权限问题**: 某些 node_modules 文件权限导致无法正确重新安装
3. **可选依赖安装失败**: npm 在安装可选依赖时遇到权限问题

## 修复步骤

### 方案 1: 手动安装平台特定模块

```bash
# 安装 rollup ARM64 原生模块
npm install @rollup/rollup-linux-arm64-gnu --save-optional
npm install @rollup/rollup-linux-arm64-musl --save-optional

# 安装 Tauri CLI ARM64 原生模块
npm install @tauri-apps/cli-linux-arm64-gnu --save-optional
npm install @tauri-apps/cli-linux-arm64-musl --save-optional
```

### 方案 2: 清理并重新安装依赖

```bash
# 清理 npm 缓存
npm cache clean --force

# 删除有问题的模块
rm -rf node_modules/@rollup
rm -rf node_modules/@tauri-apps
rm -rf node_modules/.cache

# 重新安装依赖
npm install
```

### 方案 3: 使用 yarn 代替 npm

```bash
# 安装 yarn
npm install -g yarn

# 使用 yarn 重新安装
yarn install

# 运行开发服务器
yarn tauri:dev
```

### 方案 4: 修改 package.json 添加平台特定依赖

在 package.json 的 devDependencies 中添加：

```json
{
  "devDependencies": {
    // ... 其他依赖
    "@rollup/rollup-linux-arm64-gnu": "^4.0.0",
    "@rollup/rollup-linux-arm64-musl": "^4.0.0",
    "@tauri-apps/cli-linux-arm64-gnu": "^2.0.0",
    "@tauri-apps/cli-linux-arm64-musl": "^2.0.0"
  },
  "optionalDependencies": {
    "@rollup/rollup-linux-arm64-gnu": "^4.0.0",
    "@rollup/rollup-linux-arm64-musl": "^4.0.0",
    "@tauri-apps/cli-linux-arm64-gnu": "^2.0.0",
    "@tauri-apps/cli-linux-arm64-musl": "^2.0.0"
  }
}
```

### 方案 5: 使用 Docker 容器开发

如果本地环境无法修复，可以使用 Docker 容器：

```dockerfile
FROM node:22-bullseye

# 安装 Rust
RUN curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh -s -- -y

# 设置工作目录
WORKDIR /app

# 复制项目文件
COPY package*.json ./
RUN npm install

COPY . .

# 运行开发服务器
CMD ["npm", "run", "tauri:dev"]
```

## 验证修复

运行以下命令验证修复是否成功：

```bash
# 验证 rollup 模块
ls node_modules/@rollup/ | grep arm64

# 验证 Tauri CLI 模块
ls node_modules/@tauri-apps/ | grep arm64

# 运行开发服务器
npm run tauri:dev
```

## 临时解决方案（用于开发）

如果无法立即修复平台模块问题，可以使用以下临时解决方案：

### 方案 A: 只运行前端开发服务器

```bash
npm run dev
```

这会启动 Vite 开发服务器，但不包含 Tauri 桌面应用部分。

### 方案 B: 手动构建前端

```bash
# 构建前端
npm run build

# 单独测试前端构建产物
npm run preview
```

## 长期解决方案

1. **配置 CI/CD**: 在支持的平台上进行构建
2. **使用多架构 Docker 镜像**: 支持不同的 CPU 架构
3. **提供预编译版本**: 为 ARM64 平台提供预编译的二进制文件
4. **改进文档**: 说明在 ARM64 平台上的特殊要求

## 联系支持

如果以上方案都无法解决问题，可以：

1. 检查 [Tauri 官方文档](https://tauri.app/v1/guides/)
2. 在 [Tauri GitHub Issues](https://github.com/tauri-apps/tauri/issues) 搜索类似问题
3. 在 [Tauri Discord](https://discord.gg/tauri) 寻求帮助

## 相关资源

- [Rollup 多平台支持](https://rollupjs.org/guide/en/#platform-specific-packages)
- [Node.js 原生模块](https://nodejs.org/api/addons.html)
- [ARM64 开发指南](https://developer.arm.com/)