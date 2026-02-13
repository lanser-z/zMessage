#!/bin/bash
# 备份脚本 - 备份数据库和媒体文件
set -e

BACKUP_DIR="/opt/zmessage/data/backups"
DATA_DIR="/opt/zmessage/data"
DATE=$(date +%Y%m%d_%H%M%S)
RETENTION_DAYS=30

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}=== zMessage 备份脚本 ===${NC}"
echo "备份时间: $(date)"

# 创建备份目录
mkdir -p "$BACKUP_DIR"

# 1. 备份数据库
echo -e "${YELLOW}备份数据库...${NC}"
if [ -f "$DATA_DIR/messages.db" ]; then
    sqlite3 "$DATA_DIR/messages.db" ".backup \"$BACKUP_DIR/messages_$DATE.db\""
    echo -e "${GREEN}数据库备份完成: messages_$DATE.db${NC}"
else
    echo -e "${YELLOW}警告: 数据库文件不存在${NC}"
fi

# 2. 备份配置和上传的媒体
echo -e "${YELLOW}备份媒体文件...${NC}"
if [ -d "$DATA_DIR/uploads" ] || [ -d "$DATA_DIR/media" ]; then
    tar -czf "$BACKUP_DIR/media_$DATE.tar.gz" -C "$DATA_DIR" uploads media 2>/dev/null || true
    echo -e "${GREEN}媒体备份完成: media_$DATE.tar.gz${NC}"
fi

# 3. 清理旧备份
echo -e "${YELLOW}清理 $RETENTION_DAYS 天前的旧备份...${NC}"
find "$BACKUP_DIR" -name "*.db" -mtime +$RETENTION_DAYS -delete
find "$BACKUP_DIR" -name "*.tar.gz" -mtime +$RETENTION_DAYS -delete

# 4. 显示备份信息
echo ""
echo -e "${GREEN}=== 备份完成 ===${NC}"
echo "备份目录: $BACKUP_DIR"
du -sh "$BACKUP_DIR"
ls -lh "$BACKUP_DIR" | tail -5
