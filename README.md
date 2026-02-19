# zMessage

私有协议小团队即时通讯系统

## 简介

zMessage 是一个轻量级的 1对1 即时通讯系统，采用 BS 架构设计：
- **服务端**：Go 语言开发，资源占用极低
- **客户端**：纯 Web 界面，无需安装 App
- **消息同步**：支持实时消息推送（SSE）
- **媒体支持**：文字、语音、图片消息
- **对话分享**：可分享对话给他人查看

### 核心特性

- ✅ **轻量高效**：单机可支持 50 用户、20 并发
- ✅ **实时通讯**：Server-Sent Events 实时推送
- ✅ **跨平台**：浏览器访问，无需安装客户端
- ✅ **多媒体消息**：文字、语音录音、图片上传
- ✅ **对话分享**：生成分享链接，支持消息范围选择
- ✅ **端到端传输**：媒体文件直接上传，节省服务端存储

## 技术栈

### 后端
- **语言**：Go 1.24+
- **框架**：Gin Web Framework
- **数据库**：SQLite
- **认证**：JWT
- **实时通讯**：Server-Sent Events (SSE)

### 前端
- **纯 JavaScript**：ES6+ 模块化开发
- **无框架依赖**：原生 DOM 操作
- **移动优先**：响应式设计，适配手机浏览器

## 快速开始

### 前置要求

- Go 1.24+
- 现代浏览器（Chrome、Firefox、Safari）

### 本地运行

```bash
# 克隆项目
git clone https://github.com/yourusername/zmessage.git
cd zmessage

# 运行服务端
cd server
go run main.go

# 访问客户端
# 浏览器打开 http://localhost:9405
```

### 编译生产版本

```bash
# 编译（静态链接）
CGO_ENABLED=0 go build -o zmessage-server server/main.go

# 运行
./zmessage-server
```

## 项目结构

```
zmessage/
├── client/                 # 前端静态文件
│   ├── assets/
│   │   ├── css/           # 样式文件
│   │   └── js/            # JavaScript 模块
│   └── index.html         # 入口页面
├── server/                # 后端服务
│   ├── api/              # HTTP 处理器
│   ├── dal/              # 数据访问层
│   ├── models/           # 数据模型
│   ├── modules/          # 业务模块
│   ├── pkg/              # 工具包
│   └── main.go           # 服务入口
├── docs/                 # 项目文档
│   ├── design/          # 设计文档
│   ├── 开发环境安装指南.md
│   └── 生产环境部署指南.md
└── deploy/               # 部署脚本
```

## 功能模块

| 模块 | 功能 |
|------|------|
| 用户模块 | 注册、登录、用户管理 |
| 连接模块 | SSE 实时连接、离线消息同步 |
| 消息模块 | 消息发送、接收、历史记录 |
| 媒体模块 | 图片上传、语音录制、缩略图生成 |
| 分享模块 | 对话分享、访问统计、过期管理 |

## API 文档

详细的 API 文档请参考：[docs/design/07-API接口文档.md](docs/design/07-API接口文档.md)

### 主要接口

- `POST /api/auth/register` - 用户注册
- `POST /api/auth/login` - 用户登录
- `GET /api/conversations` - 获取会话列表
- `POST /api/conversations/:id/messages` - 发送消息
- `GET /api/sse/subscribe` - SSE 订阅（实时消息）
- `POST /api/conversations/:id/share` - 创建分享
- `GET /api/shared/:token` - 查看分享（公开）

## 部署

### 生产环境部署

完整的生产环境部署指南请参考：[docs/生产环境部署指南.md](docs/生产环境部署指南.md)

### 快速部署

```bash
# 1. 编译
CGO_ENABLED=0 go build -o zmessage-server server/main.go

# 2. 创建目录
sudo mkdir -p /opt/zmessage/{bin,data,logs}
sudo cp zmessage-server /opt/zmessage/bin/
sudo cp -r client /opt/zmessage/

# 3. 创建 systemd 服务
sudo tee /etc/systemd/system/zmessage.service > /dev/null <<EOF
[Unit]
Description=zMessage IM Server
After=network.target

[Service]
Type=simple
User=zmessage
WorkingDirectory=/opt/zmessage
ExecStart=/opt/zmessage/bin/zmessage-server /opt/zmessage/data localhost:9405
Restart=always

[Install]
WantedBy=multi-user.target
EOF

# 4. 启动服务
sudo systemctl daemon-reload
sudo systemctl enable zmessage
sudo systemctl start zmessage
```

### Nginx 配置

支持子路径部署（如 `/zmessage`）：

```nginx
location /zmessage/ {
    rewrite ^/zmessage/(.*)$ /$1 break;
    proxy_pass http://localhost:9405;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
}

location /zmessage/api/sse/subscribe {
    rewrite ^/zmessage/(.*)$ /$1 break;
    proxy_pass http://localhost:9405;
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    proxy_buffering off;
}
```

## 设计文档

- [整体架构设计](docs/design/01-整体架构设计.md)
- [用户模块设计](docs/design/02-用户模块设计.md)
- [连接模块设计](docs/design/03-连接模块设计.md)
- [消息模块设计](docs/design/04-消息模块设计.md)
- [媒体模块设计](docs/design/05-媒体模块设计.md)
- [客户端模块设计](docs/design/06-客户端模块设计.md)
- [分享功能设计](docs/design/10-分享功能设计.md)

## 开发指南

### 环境准备

```bash
# 安装 Go 1.24+
# 参考：docs/开发环境安装指南.md
```

### 运行测试

```bash
cd server
go test ./...
```

### 代码规范

- 后端遵循 Go 官方代码风格
- 前端使用 ES6+ 模块化开发
- 提交前请运行 `go fmt` 和 `go vet`

## 常见问题

**Q: 语音录制需要什么条件？**

A: 语音录制使用 MediaRecorder API，需要 HTTPS 环境（localhost 除外）。

**Q: 支持群聊吗？**

A: 目前仅支持 1对1 聊天，群聊功能暂未实现。

**Q: 可以部署到哪些平台？**

A: 任何支持 Go 和 Linux 环境的平台（VPS、云服务器、本地服务器）。

**Q: 数据库支持 MySQL/PostgreSQL 吗？**

A: 当前使用 SQLite，未来可能支持其他数据库。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！

## 联系方式

- 项目地址：https://github.com/yourusername/zmessage
- 在线演示：https://lanser.fun/zmessage/
