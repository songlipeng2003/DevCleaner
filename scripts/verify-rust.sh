#!/bin/bash
# Rust 代码验证脚本

set -e

echo "=== Rust 代码验证 ==="

cd src-tauri

# 1. 代码格式化检查
echo "检查代码格式..."
if ! cargo fmt --check; then
    echo "代码格式不符合规范，请运行 'cargo fmt' 格式化代码"
    exit 1
fi
echo "✓ 代码格式正确"

# 2. Clippy 检查（允许未使用代码和 dead code）
echo "运行 Clippy 检查..."
cargo clippy --all-targets -- -D warnings -A dead_code -A unused_variables -A unused_imports
echo "✓ Clippy 检查通过"

# 3. 编译检查
echo "编译检查..."
cargo check --all-targets
echo "✓ 编译检查通过"

# 4. 运行测试（如果有）
if [ -f Cargo.toml ] && grep -q "\[dev-dependencies\]" Cargo.toml; then
    echo "运行测试..."
    cargo test
    echo "✓ 测试通过"
fi

echo ""
echo "=== 所有验证通过 ==="