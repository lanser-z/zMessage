// UI模块
export class UIModule {
    constructor(auth, connection, message, media) {
        this.auth = auth;
        this.connection = connection;
        this.message = message;
        this.media = media;
        this.currentView = null;
        this.users = [];
        this.recordingState = null;
    }

    // 初始化
    init() {
        this._setupEventListeners();
        this._checkAuth();
    }

    // 设置事件监听
    _setupEventListeners() {
        // 消息接收事件
        window.addEventListener('message:received', (e) => {
            this._handleMessageReceived(e.detail);
        });

        // 连接状态变化
        window.addEventListener('connection:change', (e) => {
            this._updateConnectionStatus(e.detail.connected);
        });
    }

    // 检查认证状态
    async _checkAuth() {
        const hasSession = await this.auth.restoreSession();

        if (!hasSession) {
            this._showLoginView();
        } else {
            this._showMainView();
        }
    }

    // 显示登录视图
    _showLoginView() {
        this.currentView = 'login';
        this._renderLogin();
    }

    // 显示主视图
    async _showMainView() {
        this.currentView = 'main';
        await this._connect();
        await this._renderMain();
    }

    // 连接WebSocket
    async _connect() {
        this.connection.connect();
    }

    // 渲染登录界面
    _renderLogin() {
        const app = document.getElementById('app');
        app.innerHTML = `
            <div class="login-container">
                <h1>zMessage</h1>
                <div class="login-form">
                    <input type="text" id="username" placeholder="用户名" required>
                    <input type="password" id="password" placeholder="密码" required>
                    <input type="text" id="nickname" placeholder="昵称 (可选)">
                    <div class="form-actions">
                        <button type="button" id="login-btn">登录</button>
                        <button type="button" id="register-btn" class="secondary">注册</button>
                    </div>
                </div>
                <p class="form-toggle">
                    <span id="login-toggle">已有账号? 去登录</span>
                    <span id="register-toggle">没有账号? 去注册</span>
                </p>
            </div>
        `;

        document.getElementById('login-toggle').style.display = 'none';

        document.getElementById('login-btn').addEventListener('click', () => this._handleLogin());
        document.getElementById('register-btn').addEventListener('click', () => this._handleRegister());

        document.getElementById('login-toggle').addEventListener('click', () => this._toggleForm('login'));
        document.getElementById('register-toggle').addEventListener('click', () => this._toggleForm('register'));
    }

    // 切换表单
    _toggleForm(type) {
        const loginToggle = document.getElementById('login-toggle');
        const registerToggle = document.getElementById('register-toggle');
        const nicknameInput = document.getElementById('nickname');

        if (type === 'register') {
            loginToggle.style.display = 'inline';
            registerToggle.style.display = 'none';
            nicknameInput.style.display = 'block';
        } else {
            loginToggle.style.display = 'none';
            registerToggle.style.display = 'inline';
            nicknameInput.style.display = 'none';
        }
    }

    // 处理登录
    async _handleLogin() {
        const username = document.getElementById('username').value.trim();
        const password = document.getElementById('password').value;

        if (!username || !password) {
            alert('请输入用户名和密码');
            return;
        }

        try {
            await this.auth.login(username, password);
            this._showMainView();
        } catch (error) {
            alert('登录失败: ' + error.message);
        }
    }

    // 处理注册
    async _handleRegister() {
        const username = document.getElementById('username').value.trim();
        const password = document.getElementById('password').value;
        const nickname = document.getElementById('nickname').value.trim();

        if (!username || !password) {
            alert('请输入用户名和密码');
            return;
        }

        try {
            await this.auth.register(username, password, nickname || username);
            this._showMainView();
        } catch (error) {
            alert('注册失败: ' + error.message);
        }
    }

    // 渲染主界面
    async _renderMain() {
        const app = document.getElementById('app');

        // 先加载用户列表
        await this._loadUsers();

        app.innerHTML = `
            <div class="main-container">
                <aside class="sidebar">
                    <div class="user-info">
                        <div class="user-avatar">
                            ${this.auth.getCurrentUser().nickname[0]}
                        </div>
                        <span class="user-name">${this.auth.getCurrentUser().nickname}</span>
                        <button id="logout-btn" class="icon-btn" title="退出">退出</button>
                    </div>
                    <div class="connection-status" id="connection-status">
                        <span class="status-dot connecting"></span>
                        <span class="status-text">连接中...</span>
                    </div>
                    <div class="user-list-header">
                        <h2>用户</h2>
                    </div>
                    <div class="user-list" id="user-list">
                        <p>加载中...</p>
                    </div>
                </aside>
                <main class="chat-area">
                    <div id="chat-container">
                        <div class="empty-state">
                            <p>选择一个用户开始聊天</p>
                        </div>
                    </div>
                </main>
            </div>
        `;

        document.getElementById('logout-btn').addEventListener('click', () => this._handleLogout());

        // 渲染用户列表
        this._renderUserList();
    }

    // 加载用户列表
    async _loadUsers() {
        try {
            const response = await fetch('/api/users', {
                headers: {
                    'Authorization': 'Bearer ' + this.auth.getToken()
                }
            });
            const result = await response.json();
            this.users = result.data.users || result.data || result.users || result || [];

            // 过滤掉当前用户
            const currentUserId = this.auth.getCurrentUser().id;
            this.users = this.users.filter(u => u.id !== currentUserId);
        } catch (error) {
            console.error('Failed to load users:', error);
            this.users = [];
        }
    }

    // 渲染用户列表
    _renderUserList() {
        const list = document.getElementById('user-list');

        if (this.users.length === 0) {
            list.innerHTML = '<p class="empty-hint">暂无其他用户</p>';
            return;
        }

        list.innerHTML = this.users.map(user => `
            <div class="user-item" data-id="${user.id}">
                <div class="avatar">${user.nickname[0]}</div>
                <div class="info">
                    <div class="name">${this._escapeHtml(user.nickname)}</div>
                </div>
            </div>
        `).join('');

        // 绑定点击事件
        list.querySelectorAll('.user-item').forEach(item => {
            item.addEventListener('click', () => {
                this._openChat(parseInt(item.dataset.id));
            });
        });
    }

    // 打开聊天
    async _openChat(userId) {
        try {
            // 获取或创建会话
            const conversation = await this.message.getConversationWithUser(userId);
            this.message.setCurrentConversation(conversation.id);

            // 加载消息
            await this._loadMessages(conversation.id);

            // 渲染聊天界面
            this._renderChatView(conversation);
        } catch (error) {
            console.error('Failed to open chat:', error);
            alert('打开聊天失败: ' + error.message);
        }
    }

    // 加载消息
    async _loadMessages(conversationId) {
        try {
            await this.message.loadMessages(conversationId);
        } catch (error) {
            console.error('Failed to load messages:', error);
        }
    }

    // 渲染聊天视图
    _renderChatView(conversation) {
        const container = document.getElementById('chat-container');

        // 获取用户信息
        const user = this.users.find(u => {
            const participant = conversation.participant;
            return participant.id === u.id || participant.id === conversation.participant_id;
        }) || conversation.participant || { nickname: '用户' };

        container.innerHTML = `
            <div class="chat-header">
                <div class="chat-user-info">
                    <div class="avatar">${user.nickname[0]}</div>
                    <div class="name">${this._escapeHtml(user.nickname)}</div>
                </div>
            </div>
            <div class="messages-container" id="messages-container">
                <p class="loading">加载消息中...</p>
            </div>
            <div class="message-input-area">
                <button id="image-btn" class="icon-btn" title="发送图片">📷</button>
                <button id="voice-btn" class="icon-btn" title="按住录音">🎤</button>
                <input type="text" id="message-input" placeholder="输入消息...">
                <button id="send-btn">发送</button>
            </div>
        `;

        // 渲染消息
        this._renderMessages();

        // 绑定事件
        document.getElementById('send-btn').addEventListener('click', () => this._sendTextMessage());
        document.getElementById('message-input').addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                this._sendTextMessage();
            }
        });
        document.getElementById('image-btn').addEventListener('click', () => this._sendImageMessage());
        document.getElementById('voice-btn').addEventListener('mousedown', () => this._startVoiceRecord());
        document.getElementById('voice-btn').addEventListener('mouseup', () => this._stopVoiceRecord());
        document.getElementById('voice-btn').addEventListener('mouseleave', () => this._cancelVoiceRecord());
    }

    // 渲染消息
    async _renderMessages() {
        const container = document.getElementById('messages-container');
        const messages = await this.message.getLocalMessages(this.message.getCurrentConversation());
        const currentUser = this.auth.getCurrentUser();

        if (messages.length === 0) {
            container.innerHTML = '<p class="empty-hint">暂无消息</p>';
            return;
        }

        container.innerHTML = messages.map(msg => {
            const isOwn = msg.sender_id === currentUser.id;
            return `
                <div class="message ${isOwn ? 'own' : 'other'}">
                    ${this._renderMessageContent(msg)}
                </div>
            `;
        }).join('');

        container.scrollTop = container.scrollHeight;
    }

    // 渲染消息内容
    _renderMessageContent(msg) {
        switch (msg.type) {
            case 'text':
                return `<div class="message-content">${this._escapeHtml(msg.content)}</div>`;
            case 'image':
                return `<img src="${this.media.getMediaUrl(msg.content)}" alt="图片" class="message-image">`;
            case 'voice':
                return `<audio src="${this.media.getMediaUrl(msg.content)}" controls class="message-voice"></audio>`;
            default:
                return `<div class="message-content">[不支持的消息类型]</div>`;
        }
    }

    // 发送文本消息
    async _sendTextMessage() {
        const input = document.getElementById('message-input');
        const text = input.value.trim();

        if (!text) return;

        try {
            await this.message.sendText(this.message.getCurrentConversation(), text);
            input.value = '';
            await this._renderMessages();
        } catch (error) {
            console.error('Failed to send message:', error);
            alert('发送失败: ' + error.message);
        }
    }

    // 发送图片消息
    async _sendImageMessage() {
        try {
            const file = await this.media.selectImage();
            const response = await this.media.uploadImage(file);
            await this.message.sendMedia(this.message.getCurrentConversation(), response.id, 'image');
            await this._renderMessages();
        } catch (error) {
            console.error('Failed to send image:', error);
            if (error.message !== '取消选择') {
                alert('发送图片失败: ' + error.message);
            }
        }
    }

    // 开始录音
    async _startVoiceRecord() {
        const btn = document.getElementById('voice-btn');
        btn.classList.add('recording');

        try {
            const recorder = await this.media.recordVoice();
            this.recordingState = recorder;
        } catch (error) {
            console.error('Failed to start recording:', error);
            btn.classList.remove('recording');
            if (error.message !== '取消录音') {
                alert('录音失败: ' + error.message);
            }
        }
    }

    // 停止录音
    async _stopVoiceRecord() {
        const btn = document.getElementById('voice-btn');
        btn.classList.remove('recording');

        if (this.recordingState && this.recordingState.stop) {
            try {
                const audioBlob = await this.recordingState.stop();
                const file = new File([audioBlob], 'voice.webm', { type: 'audio/webm' });
                const response = await this.media.uploadVoice(file);
                await this.message.sendMedia(this.message.getCurrentConversation(), response.id, 'voice');
                await this._renderMessages();
            } catch (error) {
                console.error('Failed to send voice:', error);
                alert('发送语音失败: ' + error.message);
            }
            this.recordingState = null;
        }
    }

    // 取消录音
    _cancelVoiceRecord() {
        const btn = document.getElementById('voice-btn');
        btn.classList.remove('recording');

        if (this.recordingState && this.recordingState.cancel) {
            this.recordingState.cancel();
            this.recordingState = null;
        }
    }

    // 处理消息接收
    async _handleMessageReceived(message) {
        // 如果是当前会话的消息, 刷新界面
        if (message.conversation_id === this.message.getCurrentConversation()) {
            await this._renderMessages();
        }
    }

    // 更新连接状态
    _updateConnectionStatus(connected) {
        const status = document.getElementById('connection-status');
        if (!status) return;

        const dot = status.querySelector('.status-dot');
        const text = status.querySelector('.status-text');

        if (connected) {
            dot.className = 'status-dot connected';
            text.textContent = '已连接';
        } else {
            dot.className = 'status-dot disconnected';
            text.textContent = '未连接';
        }
    }

    // 处理登出
    async _handleLogout() {
        await this.auth.logout();
        this.connection.disconnect();
        this._showLoginView();
    }

    // 工具方法
    _escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
}
