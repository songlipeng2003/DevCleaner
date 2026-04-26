#!/bin/bash

# DevCleaner Build Script
set -e

echo "=== DevCleaner Build Script ==="

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 检查依赖
check_deps() {
    echo -e "${YELLOW}检查依赖...${NC}"
    
    # Node.js
    if ! command -v node &> /dev/null; then
        echo -e "${RED}错误: Node.js 未安装${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Node.js: $(node --version)${NC}"
    
    # npm
    if ! command -v npm &> /dev/null; then
        echo -e "${RED}错误: npm 未安装${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ npm: $(npm --version)${NC}"
    
    # Rust
    if ! command -v rustc &> /dev/null; then
        echo -e "${RED}错误: Rust 未安装${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ Rust: $(rustc --version)${NC}"
}

# 安装依赖
install_deps() {
    echo -e "${YELLOW}安装依赖...${NC}"
    
    # 前端依赖
    echo "安装前端依赖..."
    npm install
    
    # Rust 目标
    echo "添加 Rust 编译目标..."
    rustup target add aarch64-apple-darwin x86_64-apple-darwin
    
    echo -e "${GREEN}✓ 依赖安装完成${NC}"
}

# 构建前端
build_frontend() {
    echo -e "${YELLOW}构建前端...${NC}"
    npm run build
    echo -e "${GREEN}✓ 前端构建完成${NC}"
}

# 构建 Tauri
build_tauri() {
    echo -e "${YELLOW}构建 Tauri 应用...${NC}"
    
    # 根据平台选择目标
    case "$(uname -s)" in
        Darwin*)
            rustup target add aarch64-apple-darwin x86_64-apple-darwin
            ;;
        Linux*)
            rustup target add x86_64-unknown-linux-gnu
            ;;
        MINGW*|CYGWIN*|MSYS*)
            rustup target add x86_64-pc-windows-gnu
            ;;
    esac
    
    npm run tauri:build
    echo -e "${GREEN}✓ Tauri 构建完成${NC}"
}

# 显示帮助
show_help() {
    echo "DevCleaner Build Script"
    echo ""
    echo "用法: $0 [命令]"
    echo ""
    echo "命令:"
    echo "  all         执行所有构建步骤 (默认)"
    echo "  check       检查依赖"
    echo "  install     安装依赖"
    echo "  frontend    仅构建前端"
    echo "  tauri       仅构建 Tauri 应用"
    echo "  clean       清理构建产物"
    echo "  help        显示帮助"
}

# 清理
clean() {
    echo -e "${YELLOW}清理构建产物...${NC}"
    rm -rf dist/
    cd src-tauri && cargo clean && cd ..
    echo -e "${GREEN}✓ 清理完成${NC}"
}

# 主逻辑
case "${1:-all}" in
    all)
        check_deps
        install_deps
        build_frontend
        build_tauri
        echo -e "${GREEN}=== 构建完成 ===${NC}"
        ;;
    check)
        check_deps
        ;;
    install)
        check_deps
        install_deps
        ;;
    frontend)
        build_frontend
        ;;
    tauri)
        build_tauri
        ;;
    clean)
        clean
        ;;
    help)
        show_help
        ;;
    *)
        echo -e "${RED}未知命令: $1${NC}"
        show_help
        exit 1
        ;;
esac
