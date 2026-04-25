#!/bin/bash

# DevCleaner macOS DMG 打包脚本
set -e

APP_NAME="DevCleaner"
VERSION="0.1.0-alpha"
BUNDLE_ID="com.devcleaner.app"

echo "=== DevCleaner macOS 打包脚本 ==="

# 检查是否是 Tauri 构建
TAURI_DIR="src-tauri/target/release/bundle/macos"
if [ ! -d "$TAURI_DIR" ]; then
    echo "错误: 请先运行 npm run tauri:build 构建应用"
    exit 1
fi

# 查找构建产物
APP_FILE=$(find "$TAURI_DIR" -name "*.app" -type d | head -1)
if [ -z "$APP_FILE" ]; then
    echo "错误: 未找到 .app 文件"
    exit 1
fi

echo "找到应用: $APP_FILE"

# 创建 DMG
DMG_NAME="${APP_NAME}-${VERSION}-macOS.dmg"
DMG_TEMP="/tmp/${APP_NAME}.dmg"
MOUNT_POINT="/tmp/${APP_NAME}-mount"

echo "创建 DMG: $DMG_NAME"

# 清理旧文件
rm -f "$DMG_NAME" "$DMG_TEMP"
rm -rf "$MOUNT_POINT"

# 创建临时 DMG
hdiutil create -size 500M -fs HFS+ -volname "$APP_NAME" -o "$DMG_TEMP"

# 挂载并复制文件
hdiutil attach "$DMG_TEMP" -mountpoint "$MOUNT_POINT"

# 复制应用
cp -R "$APP_FILE" "$MOUNT_POINT/"

# 创建 Applications 链接
ln -sf "/Applications/$APP_NAME.app" "$MOUNT_POINT/Applications"

# 卸载
hdiutil detach "$MOUNT_POINT"

# 转换 DMG 格式
hdiutil convert "$DMG_TEMP" -format UDZO -o "$DMG_NAME"

# 清理
rm -f "$DMG_TEMP"

echo "✓ DMG 创建完成: $DMG_NAME"
ls -lh "$DMG_NAME"
