#!/bin/bash
# 构建脚本 - 编译生产版本
set -e

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# 获取版本信息
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION=$(go version | awk '{print $3}')

echo -e "${GREEN}=== zMessage 构建脚本 ===${NC}"
echo "版本: $VERSION"
echo "构建时间: $BUILD_TIME"
echo "Go 版本: $GO_VERSION"
echo ""

# 1. 检查 Go 环境
echo -e "${YELLOW}[1/4] 检查 Go 环境...${NC}"
if ! command -v go &>/dev/null; then
    echo -e "${RED}错误: 未找到 Go，请先安装 Go 1.24+${NC}"
    exit 1
fi
echo -e "${GREEN}Go 版本: $(go version)${NC}"

# 2. 下载依赖
echo -e "${YELLOW}[2/4] 下载依赖...${NC}"
go mod download
go mod verify
echo -e "${GREEN}依赖验证完成${NC}"

# 3. 运行测试
echo -e "${YELLOW}[3/4] 运行测试...${NC}"
go test -v -race -cover ./... || echo -e "${YELLOW}警告: 部分测试失败${NC}"

# 4. 编译生产版本
echo -e "${YELLOW}[4/4] 编译生产版本...${NC}"

# 输出目录
OUTPUT_DIR="build"
mkdir -p "$OUTPUT_DIR"

# 构建参数
LDFLAGS="-s -w -X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME"

# 构建多平台版本
PLATFORMS=(
    "linux/amd64"
    "linux/arm64"
    "linux/386"
)

for PLATFORM in "${PLATFORMS[@]}"; do
    GOOS="${PLATFORM%/*}"
    GOARCH="${PLATFORM#*/}"

    OUTPUT_NAME="$OUTPUT_DIR/zmessage-server-$VERSION-$GOOS-$GOARCH"

    echo "构建 $PLATFORM..."

    CGO_ENABLED=0 GOOS="$GOOS" GOARCH="$GOARCH" go build \
        -ldflags="$LDFLAGS" \
        -o "$OUTPUT_NAME" \
        server/main.go

    # 压缩
    gzip -f "$OUTPUT_NAME"
    echo -e "${GREEN}✓ $OUTPUT_NAME.gz${NC}"
done

# 创建符号链接
ln -sf "zmessage-server-$VERSION-linux-amd64" "$OUTPUT_DIR/zmessage-server"

echo ""
echo -e "${GREEN}=== 构建完成 ===${NC}"
echo "输出目录: $OUTPUT_DIR"
ls -lh "$OUTPUT_DIR/"
