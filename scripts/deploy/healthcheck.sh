#!/bin/bash
# 健康检查脚本 - 检查服务状态
set -e

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

ERRORS=0

echo -e "${GREEN}=== zMessage 健康检查 ===${NC}"
echo "检查时间: $(date)"
echo ""

# 1. 检查服务状态
echo -n "检查服务状态... "
if systemctl is-active --quiet zmessage; then
    echo -e "${GREEN}运行中${NC}"
else
    echo -e "${RED}未运行${NC}"
    ((ERRORS++))
fi

# 2. 检查端口监听
echo -n "检查端口监听 (9405)... "
if ss -ln 2>/dev/null | grep -q ":9405"; then
    echo -e "${GREEN}正常${NC}"
else
    echo -e "${RED}未监听${NC}"
    ((ERRORS++))
fi

# 3. 检查 API 响应
echo -n "检查 API 响应... "
if curl -sf http://127.0.0.1:9405/ > /dev/null 2>&1; then
    echo -e "${GREEN}正常${NC}"
else
    echo -e "${RED}无响应${NC}"
    ((ERRORS++))
fi

# 4. 检查磁盘空间
echo -n "检查磁盘空间... "
DISK_USAGE=$(df /opt/zmessage | awk 'NR==2 {print $5}' | sed 's/%//')
if [ "$DISK_USAGE" -lt 90 ]; then
    echo -e "${GREEN}正常 ($DISK_USAGE%)${NC}"
else
    echo -e "${YELLOW}警告 ($DISK_USAGE%)${NC}"
fi

# 5. 检查数据库文件
echo -n "检查数据库文件... "
if [ -f "/opt/zmessage/data/messages.db" ]; then
    DB_SIZE=$(du -h /opt/zmessage/data/messages.db | cut -f1)
    echo -e "${GREEN}存在 ($DB_SIZE)${NC}"
else
    echo -e "${YELLOW}未初始化${NC}"
fi

# 6. 检查内存使用
echo -n "检查内存使用... "
MEMORY=$(ps aux | grep '[z]message-server' | awk '{print $6}' | head -1)
if [ -n "$MEMORY" ]; then
    MEMORY_MB=$((MEMORY / 1024))
    echo -e "${GREEN}正常 (${MEMORY_MB}MB)${NC}"
else
    echo -e "${YELLOW}未运行${NC}"
fi

# 7. 检查日志文件
echo -n "检查日志文件... "
if [ -d "/opt/zmessage/logs" ]; then
    LOG_SIZE=$(du -sh /opt/zmessage/logs 2>/dev/null | cut -f1)
    echo -e "${GREEN}正常 ($LOG_SIZE)${NC}"
else
    echo -e "${YELLOW}不存在${NC}"
fi

# 汇总
echo ""
if [ $ERRORS -eq 0 ]; then
    echo -e "${GREEN}=== 所有检查通过 ===${NC}"
    exit 0
else
    echo -e "${RED}=== 发现 $ERRORS 个错误 ===${NC}"
    exit 1
fi
