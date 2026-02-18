// UI模块
export class UIModule {
    constructor(auth, connection, message, media, share, apiClient) {
        this.auth = auth;
        this.connection = connection;
        this.message = message;
        this.media = media;
        this.share = share;
        this.apiClient = apiClient;
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

        // 消息删除事件
        window.addEventListener('message:deleted', (e) => {
            this._handleMessageDeleted(e.detail);
        });

        // 连接状态变化
        window.addEventListener('connection:change', (e) => {
            this._updateConnectionStatus(e.detail.connected);
        });
    }

    // 检查认证状态
    async _checkAuth() {
        // 检查是否有分享 token 参数
        const shareToken = this.share.getShareTokenFromURL();
        if (shareToken) {
            // 显示分享页面（无需登录）
            await this.renderSharedView(shareToken);
            return;
        }

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
            <div class="sidebar-overlay" id="sidebar-overlay"></div>
            <div class="main-container">
                <aside class="sidebar" id="sidebar">
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
                    <div class="sidebar-menu">
                        <div class="menu-item" id="my-shares-btn">
                            <span class="menu-icon">📋</span>
                            <span class="menu-text">我的分享</span>
                        </div>
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
                        <div class="chat-header">
                            <button class="menu-btn" id="menu-btn">☰</button>
                            <div class="chat-user-info">
                                <span>选择用户开始聊天</span>
                            </div>
                        </div>
                        <div class="messages-container" id="messages-container">
                            <div class="empty-state">
                                <p>从菜单选择一个用户开始聊天</p>
                            </div>
                        </div>
                    </div>
                </main>
            </div>
        `;

        // 绑定侧边栏切换
        document.getElementById('sidebar-overlay').addEventListener('click', () => this._toggleSidebar(false));
        document.getElementById('logout-btn').addEventListener('click', () => this._handleLogout());

        // 绑定"我的分享"按钮
        const mySharesBtn = document.getElementById('my-shares-btn');
        if (mySharesBtn) {
            mySharesBtn.addEventListener('click', () => {
                this._toggleSidebar(false);
                this.showMyShares();
            });
        }

        // 绑定菜单按钮
        const menuBtn = document.getElementById('menu-btn');
        if (menuBtn) {
            menuBtn.addEventListener('click', () => this._toggleSidebar(true));
        }

        // 渲染用户列表
        this._renderUserList();
    }

    // 加载用户列表
    async _loadUsers() {
        try {
            console.log('[UI] Loading users...');
            const result = await this.apiClient.get('/api/users');
            console.log('[UI] API result:', result);

            this.users = result.users || result.data || result || [];
            console.log('[UI] Parsed users:', this.users);

            // 过滤掉当前用户
            const currentUserId = this.auth.getCurrentUser().id;
            console.log('[UI] Current user ID:', currentUserId);
            this.users = this.users.filter(u => u.id !== currentUserId);
            console.log('[UI] Filtered users:', this.users);
        } catch (error) {
            console.error('[UI] Failed to load users:', error);
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
                    <div class="username">@${this._escapeHtml(user.username)}</div>
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
            // 移动端选择用户后关闭侧边栏
            this._toggleSidebar(false);

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
                <button class="menu-btn" id="menu-btn">☰</button>
                <div class="chat-user-info">
                    <div class="avatar">${user.nickname[0]}</div>
                    <div class="name">${this._escapeHtml(user.nickname)}</div>
                </div>
                <button class="share-btn" id="share-btn" title="分享对话">📋</button>
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
        const menuBtn = document.getElementById('menu-btn');
        if (menuBtn) {
            menuBtn.addEventListener('click', () => this._toggleSidebar(true));
        }
        // 绑定分享按钮
        const shareBtn = document.getElementById('share-btn');
        if (shareBtn) {
            shareBtn.addEventListener('click', () => {
                const conv = this.message.getCurrentConversation();
                if (conv) {
                    this.showShareDialog(conv);
                }
            });
        }
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

    // 渲染消息（只显示最近10条）
    async _renderMessages() {
        const container = document.getElementById('messages-container');
        const messages = await this.message.getLocalMessages(this.message.getCurrentConversation());
        const currentUser = this.auth.getCurrentUser();

        if (messages.length === 0) {
            container.innerHTML = '<p class="empty-hint">暂无消息</p>';
            return;
        }

        // 只显示最近10条消息
        const recentMessages = messages.slice(-10);

        container.innerHTML = recentMessages.map(msg => {
            const isOwn = msg.sender_id === currentUser.id;
            return `
                <div class="message ${isOwn ? 'own' : 'other'}" data-message-id="${msg.id}">
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
            // sendText 内部已经有乐观更新和事件通知，不需要手动刷新
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
            // sendMedia 内部已经有乐观更新和事件通知，不需要手动刷新
            await this.message.sendMedia(this.message.getCurrentConversation(), response.id, 'image');
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
                // sendMedia 内部已经有乐观更新和事件通知，不需要手动刷新
                await this.message.sendMedia(this.message.getCurrentConversation(), response.id, 'voice');
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
        console.log('[UI] Message received:', message);
        console.log('[UI] Current conversation:', this.message.getCurrentConversation());

        // 如果是当前会话的消息, 刷新界面
        if (message.conversation_id === this.message.getCurrentConversation()) {
            console.log('[UI] Rendering messages for current conversation');
            await this._renderMessages();
        } else {
            console.log('[UI] Message for different conversation, ignoring');
        }
    }

    // 处理消息删除
    async _handleMessageDeleted(messageId) {
        console.log('[UI] Message deleted:', messageId);
        // 临时消息被删除后，真实消息已经添加，需要刷新UI
        await this._renderMessages();
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

    // 切换侧边栏显示/隐藏
    _toggleSidebar(show) {
        const sidebar = document.getElementById('sidebar');
        const overlay = document.getElementById('sidebar-overlay');
        if (!sidebar || !overlay) return;

        if (show) {
            sidebar.classList.add('open');
            overlay.classList.add('active');
        } else {
            sidebar.classList.remove('open');
            overlay.classList.remove('active');
        }
    }

    // 工具方法
    _escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    // 显示分享配置对话框
    showShareDialog(conversationID) {
        const dialog = document.createElement('div');
        dialog.className = 'share-dialog-overlay';
        dialog.innerHTML = `
            <div class="share-dialog">
                <div class="share-dialog-header">
                    <h3>分享对话</h3>
                    <button class="close-btn" data-action="close">&times;</button>
                </div>
                <div class="share-dialog-content">
                    <div class="form-group">
                        <label>分享范围</label>
                        <div class="radio-group">
                            <label><input type="radio" name="message_range" value="recent" checked> 最近 <input type="number" id="recent-count" value="50" min="1" max="500"> 条</label>
                            <label><input type="radio" name="message_range" value="all"> 全部消息</label>
                        </div>
                    </div>
                    <div class="form-group">
                        <label>过期时间</label>
                        <div class="radio-group">
                            <label><input type="radio" name="expire_days" value="1"> 1天</label>
                            <label><input type="radio" name="expire_days" value="7" checked> 7天</label>
                            <label><input type="radio" name="expire_days" value="30"> 30天</label>
                            <label><input type="radio" name="expire_days" value="0"> 永久有效</label>
                        </div>
                    </div>
                </div>
                <div class="share-dialog-actions">
                    <button class="btn-secondary" data-action="cancel">取消</button>
                    <button class="btn-primary" data-action="share">分享</button>
                </div>
            </div>
        `;

        document.body.appendChild(dialog);

        // 绑定事件
        dialog.querySelectorAll('[data-action]').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                const action = e.currentTarget.dataset.action;

                if (action === 'close' || action === 'cancel') {
                    document.body.removeChild(dialog);
                } else if (action === 'share') {
                    await this._handleShare(conversationID, dialog);
                }
            });
        });
    }

    // 处理分享创建
    async _handleShare(conversationID, dialog) {
        const messageRange = dialog.querySelector('input[name="message_range"]:checked').value;
        const recentCount = parseInt(dialog.querySelector('#recent-count').value) || 50;
        const expireDays = parseInt(dialog.querySelector('input[name="expire_days"]:checked').value);

        try {
            const shareData = await this.share.createShare(conversationID, {
                messageRange,
                recentCount,
                expireDays
            });

            // 显示成功对话框
            document.body.removeChild(dialog);
            this._showShareSuccess(shareData);
        } catch (error) {
            console.error('创建分享失败:', error);
            alert('创建分享失败: ' + error.message);
        }
    }

    // 显示分享成功对话框
    _showShareSuccess(shareData) {
        const dialog = document.createElement('div');
        dialog.className = 'share-dialog-overlay';
        dialog.innerHTML = `
            <div class="share-dialog share-success-dialog">
                <div class="share-dialog-header">
                    <h3>分享成功</h3>
                </div>
                <div class="share-dialog-content">
                    <p>分享链接已生成，${this.share.formatExpireTime(shareData.expire_at)}</p>
                    <div class="share-url-box">
                        <input type="text" id="share-url" value="${shareData.full_url}" readonly>
                        <button class="btn-copy" data-action="copy">复制</button>
                    </div>
                    <p class="hint">分享 ${shareData.message_count} 条消息</p>
                </div>
                <div class="share-dialog-actions">
                    <button class="btn-primary" data-action="done">完成</button>
                </div>
            </div>
        `;

        document.body.appendChild(dialog);

        // 绑定事件
        dialog.querySelectorAll('[data-action]').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                const action = e.target.dataset.action;
                const button = e.target;

                if (action === 'copy') {
                    const success = await this.share.copyShareURL(shareData.full_url);
                    if (success) {
                        button.textContent = '✓ 已复制';
                        button.classList.add('copied');
                        setTimeout(() => {
                            button.textContent = '复制';
                            button.classList.remove('copied');
                        }, 2000);
                    } else {
                        alert('复制失败，请手动复制');
                    }
                } else if (action === 'done') {
                    document.body.removeChild(dialog);
                }
            });
        });
    }

    // 显示我的分享列表
    async showMyShares() {
        try {
            const data = await this.share.getMyShares(1, 50);

            const app = document.getElementById('app');
            app.innerHTML = `
                <div class="shares-view">
                    <div class="shares-header">
                        <button class="btn-back" data-action="back">← 返回</button>
                        <h2>我的分享</h2>
                    </div>
                    <div class="shares-list">
                        ${this._renderSharesList(data.shares)}
                    </div>
                </div>
            `;

            // 绑定返回按钮
            app.querySelector('[data-action="back"]')?.addEventListener('click', () => {
                this._showMainView();
            });

            // 绑定分享项操作
            app.querySelectorAll('[data-action]').forEach(el => {
                el.addEventListener('click', async (e) => {
                    e.preventDefault();
                    const action = el.dataset.action;
                    const shareId = el.dataset.shareId;
                    const shareUrl = el.dataset.shareUrl;

                    if (action === 'copy-share') {
                        const success = await this.share.copyShareURL(shareUrl);
                        if (success) {
                            el.textContent = '✓ 已复制';
                            el.classList.add('copied');
                            setTimeout(() => {
                                el.textContent = '复制链接';
                                el.classList.remove('copied');
                            }, 2000);
                        } else {
                            alert('复制失败');
                        }
                    } else if (action === 'delete-share') {
                        if (confirm('确定要取消这个分享吗？')) {
                            try {
                                await this.share.deleteShare(parseInt(shareId));
                                this.showMyShares(); // 刷新列表
                            } catch (error) {
                                alert('删除失败: ' + error.message);
                            }
                        }
                    }
                });
            });
        } catch (error) {
            console.error('获取分享列表失败:', error);
            alert('获取分享列表失败: ' + error.message);
            this._showMainView();
        }
    }

    // 渲染分享列表
    _renderSharesList(shares) {
        if (!shares || shares.length === 0) {
            return '<div class="empty-state">暂无分享记录</div>';
        }

        return shares.map(share => `
            <div class="share-item ${share.is_expired ? 'expired' : ''}">
                <div class="share-item-info">
                    <div class="share-item-title">与 ${share.participant.nickname} 的对话</div>
                    <div class="share-item-meta">
                        ${share.message_count} 条消息
                        · ${this.share.formatExpireTime(share.expire_at)}
                        · ${this.share.formatViewCount(share.view_count)}
                    </div>
                </div>
                <div class="share-item-actions">
                    ${!share.is_expired ? `
                        <button class="btn-link" data-action="copy-share" data-share-url="${share.full_url || share.share_url}">复制链接</button>
                        <button class="btn-link btn-danger" data-action="delete-share" data-share-id="${share.id}">取消分享</button>
                    ` : `
                        <span class="expired-badge">已过期</span>
                        <button class="btn-link" data-action="delete-share" data-share-id="${share.id}">删除记录</button>
                    `}
                </div>
            </div>
        `).join('');
    }

    // 渲染分享页面（公开访问）
    async renderSharedView(token) {
        try {
            const data = await this.share.getSharedContent(token);

            const app = document.getElementById('app');
            app.innerHTML = `
                <div class="shared-view">
                    <div class="shared-header">
                        <div class="shared-icon">📱</div>
                        <h2>对话分享</h2>
                        <div class="shared-participants">
                            ${data.share.participants.map(p => `
                                <span class="participant-badge">${this._escapeHtml(p.nickname)}</span>
                            `).join('')}
                        </div>
                        <div class="shared-meta">
                            ${data.share.message_count} 条消息
                            · 分享于 ${new Date(data.share.created_at * 1000).toLocaleDateString('zh-CN')}
                            ${data.share.view_count > 0 ? `· ${this.share.formatViewCount(data.share.view_count)}` : ''}
                        </div>
                    </div>

                    <div class="shared-messages" id="shared-messages">
                        ${this._renderSharedMessages(data.messages)}
                    </div>

                    ${data.has_more ? `
                        <div class="shared-footer">
                            <button class="btn-secondary" id="load-more">加载更多</button>
                        </div>
                    ` : ''}

                    <div class="shared-expire-info">
                        ${data.share.is_expired ? '此分享已过期' : this.share.formatExpireTime(data.share.expire_at)}
                    </div>

                    <div class="shared-branding">
                        由 <strong>zMessage</strong> 提供支持
                    </div>
                </div>
            `;

            // 绑定加载更多按钮
            const loadMoreBtn = document.getElementById('load-more');
            if (loadMoreBtn && data.has_more) {
                loadMoreBtn.addEventListener('click', async () => {
                    const oldestId = data.messages[data.messages.length - 1]?.id || 0;
                    if (oldestId > 0) {
                        await this._loadMoreSharedMessages(token, oldestId);
                    }
                });
            }
        } catch (error) {
            console.error('加载分享内容失败:', error);
            document.getElementById('app').innerHTML = `
                <div class="error-page">
                    <div class="error-icon">🔒</div>
                    <h2>分享不存在或已过期</h2>
                    <p>该分享链接可能已被删除或已过期</p>
                </div>
            `;
        }
    }

    // 渲染分享的消息
    _renderSharedMessages(messages) {
        if (!messages || messages.length === 0) {
            return '<div class="empty-messages">暂无消息</div>';
        }
        return messages.map(msg => `
            <div class="shared-message ${msg.sender_id === messages[0]?.sender_id ? 'same-sender' : ''}">
                <div class="shared-message-sender">${this._escapeHtml(msg.sender_nickname)}</div>
                <div class="shared-message-content">
                    ${this._renderMessageContent(msg)}
                </div>
                <div class="shared-message-time">${new Date(msg.created_at * 1000).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })}</div>
            </div>
        `).join('');
    }

    // 加载更多分享消息
    async _loadMoreSharedMessages(token, beforeId) {
        try {
            const data = await this.share.getSharedContent(token, beforeId, 50);
            const messagesContainer = document.getElementById('shared-messages');
            const loadMoreBtn = document.getElementById('load-more');

            // 将新消息添加到开头
            const newMessagesHtml = this._renderSharedMessages(data.messages);
            messagesContainer.insertAdjacentHTML('afterbegin', newMessagesHtml);

            // 更新或删除加载更多按钮
            if (!data.has_more) {
                loadMoreBtn?.remove();
            }
        } catch (error) {
            console.error('加载更多消息失败:', error);
            alert('加载失败: ' + error.message);
        }
    }
}
