#!/bin/bash
# 生产环境安装脚本
# 用法: sudo ./install.sh

set -e

# 颜色输出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# 配置变量
DOMAIN="${1:-lanser.fun}"
EMAIL="${2:-}"
INSTALL_DIR="/opt/zmessage"
SERVICE_USER="zmessage"

echo -e "${GREEN}=== zMessage 生产环境安装 ===${NC}"
echo "域名: $DOMAIN"
echo ""

# 检查 root 权限
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}请使用 sudo 运行此脚本${NC}"
    exit 1
fi

# 1. 创建用户
echo -e "${YELLOW}[1/8] 创建服务用户...${NC}"
if ! id "$SERVICE_USER" &>/dev/null; then
    useradd -r -s /bin/false -m -d "$INSTALL_DIR" "$SERVICE_USER"
    echo -e "${GREEN}用户 $SERVICE_USER 已创建${NC}"
else
    echo -e "${GREEN}用户 $SERVICE_USER 已存在${NC}"
fi

# 2. 创建目录结构
echo -e "${YELLOW}[2/8] 创建目录结构...${NC}"
mkdir -p "$INSTALL_DIR"/{bin,data,logs,scripts}
mkdir -p "$INSTALL_DIR/data"/{uploads,media,backups}
chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"
chmod -R 750 "$INSTALL_DIR"
echo -e "${GREEN}目录结构已创建${NC}"

# 3. 安装系统依赖
echo -e "${YELLOW}[3/8] 安装系统依赖...${NC}"
apt update
apt install -y nginx certbot python3-certbot-nginx fail2ban
echo -e "${GREEN}系统依赖已安装${NC}"

# 4. 安装 systemd 服务
echo -e "${YELLOW}[4/8] 安装 systemd 服务...${NC}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cp "$SCRIPT_DIR/../deploy/systemd/zmessage.service" /etc/systemd/system/
systemctl daemon-reload
echo -e "${GREEN}systemd 服务已安装${NC}"

# 5. 安装 Nginx 配置
echo -e "${YELLOW}[5/8] 安装 Nginx 配置...${NC}"
sed "s/lanser.fun/$DOMAIN/g" "$SCRIPT_DIR/../deploy/nginx/zmessage.conf" > /etc/nginx/sites-available/zmessage
ln -sf /etc/nginx/sites-available/zmessage /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
echo -e "${GREEN}Nginx 配置已安装${NC}"

# 6. 配置 logrotate
echo -e "${YELLOW}[6/8] 配置日志轮转...${NC}"
cp "$SCRIPT_DIR/../deploy/logrotate/zmessage" /etc/logrotate.d/
cp "$SCRIPT_DIR/../deploy/logrotate/nginx-zmessage" /etc/logrotate.d/
echo -e "${GREEN}日志轮转已配置${NC}"

# 7. 配置防火墙
echo -e "${YELLOW}[7/8] 配置防火墙...${NC}"
if command -v ufw &>/dev/null; then
    ufw allow 22/tcp
    ufw allow 80/tcp
    ufw allow 443/tcp
    echo -e "${GREEN}防火墙规则已添加${NC}"
else
    echo -e "${YELLOW}ufw 未安装，跳过防火墙配置${NC}"
fi

# 8. 获取 SSL 证书
echo -e "${YELLOW}[8/8] 获取 SSL 证书...${NC}"
if [ -n "$EMAIL" ]; then
    certbot --nginx -d "$DOMAIN" --email "$EMAIL" --agree-tos --no-eff-email --non-interactive
else
    echo -e "${YELLOW}未提供邮箱，跳过自动证书获取${NC}"
    echo "请手动运行: sudo certbot --nginx -d $DOMAIN"
fi

# 安装监控脚本
cp "$SCRIPT_DIR/backup.sh" "$INSTALL_DIR/scripts/"
cp "$SCRIPT_DIR/healthcheck.sh" "$INSTALL_DIR/scripts/"
chmod +x "$INSTALL_DIR/scripts/"*.sh
chown "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR/scripts/"*.sh

# 配置 fail2ban
cat > /etc/fail2ban/jail.d/zmessage.local << EOF
[zmessage]
enabled = true
port = http,https
filter = zmessage
logpath = /var/log/nginx/zmessage-access.log
maxretry = 10
findtime = 600
bantime = 3600
EOF

cat > /etc/fail2ban/filter.d/zmessage.conf << 'EOF'
[Definition]
failregex = <HOST> .* "(GET|POST) .*" 401
ignoreregex =
EOF

systemctl enable fail2ban
systemctl start fail2ban

echo ""
echo -e "${GREEN}=== 安装完成 ===${NC}"
echo ""
echo "接下来的步骤:"
echo "1. 将编译好的二进制文件复制到 $INSTALL_DIR/bin/zmessage-server"
echo "2. 将前端文件复制到 $INSTALL_DIR/client/"
echo "3. 启动服务: sudo systemctl start zmessage"
echo "4. 检查状态: sudo systemctl status zmessage"
echo ""
echo "健康检查: $INSTALL_DIR/scripts/healthcheck.sh"
echo "查看日志: sudo journalctl -u zmessage -f"
