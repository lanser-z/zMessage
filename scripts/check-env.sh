#!/bin/bash

echo "=== 开发环境检查 ==="
echo ""

# 检查 Go
echo -n "Go: "
if command -v go &> /dev/null; then
    go version
else
    echo "未安装"
    exit 1
fi

# 检查 Git
echo -n "Git: "
if command -v git &> /dev/null; then
    git --version
else
    echo "未安装"
fi

# 检查 SQLite3
echo -n "SQLite3: "
if command -v sqlite3 &> /dev/null; then
    sqlite3 --version
else
    echo "未安装 (可选)"
fi

# 检查 curl
echo -n "curl: "
if command -v curl &> /dev/null; then
    curl --version | head -1
else
    echo "未安装 (可选)"
fi

# 检查项目目录
echo ""
echo "=== 项目目录检查 ==="
if [ -d "/home/lanser/aiprj/zMessage" ]; then
    echo "✓ 项目目录存在"
else
    echo "✗ 项目目录不存在"
    exit 1
fi

# 检查 Go 模块
echo ""
echo "=== Go 模块检查 ==="
if [ -f "/home/lanser/aiprj/zMessage/go.mod" ]; then
    echo "✓ go.mod 已存在"
else
    echo "✗ go.mod 不存在"
    exit 1
fi

# 检查证书
echo ""
echo "=== 证书检查 ==="
if [ -f "/home/lanser/aiprj/zMessage/data/certs/cert.pem" ]; then
    echo "✓ SSL 证书已生成"
    openssl x509 -in /home/lanser/aiprj/zMessage/data/certs/cert.pem -noout -subject -dates 2>/dev/null | sed 's/^/  /'
else
    echo "✗ SSL 证书不存在"
fi

# 检查目录结构
echo ""
echo "=== 目录结构检查 ==="
dirs=(
    "server"
    "server/api"
    "server/ws"
    "server/modules"
    "server/dal"
    "server/models"
    "server/pkg"
    "client"
    "client/assets"
    "data"
    "data/uploads"
    "data/certs"
    "docs"
    "scripts"
)

for dir in "${dirs[@]}"; do
    if [ -d "/home/lanser/aiprj/zMessage/$dir" ]; then
        echo "✓ $dir"
    else
        echo "✗ $dir 缺失"
    fi
done

echo ""
echo "=== 环境检查完成 ==="
