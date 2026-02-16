// SSE 连接模块
export class ConnectionModule {
    constructor(auth, message, config) {
        this.auth = auth;
        this.message = message;
        this.config = config;
        this.eventSource = null;
        this.connected = false;
        this.reconnectAttempts = 0;
        this.reconnectTimer = null;
        this.reconnecting = false;
        this.lastConnectTime = 0; // 上次连接时间，用于拉取离线消息
        this.lastPollTime = 0; // 上次轮询时间
        this.pollingTimer = null; // 轮询定时器
        this.isMobile = this._detectMobile(); // 是否为移动设备
        this.lastMessageId = 0; // 记录收到的最新消息ID
        this.hasReceivedMessage = false; // 是否收到过消息
    }

    // 检测是否为移动设备
    _detectMobile() {
        const userAgent = navigator.userAgent;
        return /Mobi|Android|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(userAgent);
    }

    // 连接 SSE
    connect() {
        const token = this.auth.getToken();
        // 使用相对路径，受 <base href> 影响
        const sseUrl = `api/sse/subscribe?token=${token}`;
        console.log(`[SSE] Connecting (${this.isMobile ? 'mobile' : 'desktop'} detected):`, sseUrl);

        // 关闭旧的连接（防止累积）
        if (this.eventSource) {
            this.eventSource.close();
        }

        this.reconnecting = false;
        this.eventSource = new EventSource(sseUrl);

        // 移动端启动轮询作为后备方案
        if (this.isMobile) {
            console.log('[SSE] Mobile device detected, starting polling fallback');
            this._startPolling();
        }

        this.eventSource.onopen = () => {
            console.log('[SSE] Connection opened');
            this.connected = true;
            this.reconnectAttempts = 0;
            this.reconnecting = false;
            this._notifyConnectionChange(true);
        };

        this.eventSource.addEventListener('connected', (event) => {
            const data = JSON.parse(event.data);
            console.log('[SSE] Connected event:', data);

            // 检查是否是重连（有上次连接时间）
            const now = Date.now() / 1000;
            const isReconnect = this.lastConnectTime > 0;
            const disconnectedTime = isReconnect ? now - this.lastConnectTime : 0;

            if (isReconnect && disconnectedTime > 5) {
                console.log(`[SSE] Reconnected after ${disconnectedTime.toFixed(1)}s, fetching offline messages...`);
                this._fetchOfflineMessages();
            }

            this.lastConnectTime = now;
            this.hasReceivedMessage = false; // 重置标志，等待第一条真实消息
        });

        this.eventSource.addEventListener('chat', (event) => {
            const data = JSON.parse(event.data);
            console.log('[SSE] Chat message push:', data);
            // 记录收到的最新消息ID
            if (data.message_id) {
                this.lastMessageId = data.message_id;
                this.hasReceivedMessage = true;
            }
            this._handleMessagePush(data);
        });

        this.eventSource.addEventListener('heartbeat', (event) => {
            const data = JSON.parse(event.data);
            console.log('[SSE] Heartbeat:', data.timestamp);
        });

        this.eventSource.onerror = (error) => {
            console.error('[SSE] Error:', error);
            this.connected = false;
            this._notifyConnectionChange(false);

            // 避免重复设置重连
            if (this.reconnecting) {
                return;
            }
            this.reconnecting = true;

            // 尝试重连（指数退避）
            const delay = Math.min(1000 * Math.pow(1.5, this.reconnectAttempts), 30000);
            console.log(`[SSE] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts + 1})`);

            this.reconnectTimer = setTimeout(() => {
                this.reconnectAttempts++;
                this.connect();
            }, delay);
        };

        this.eventSource.onclose = () => {
            console.log('[SSE] Connection closed');
            this.connected = false;
            this.reconnecting = false;
            this._notifyConnectionChange(false);
        };
    }

    // 断开连接
    disconnect() {
        if (this.eventSource) {
            this.eventSource.close();
            this.eventSource = null;
        }
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
            this.reconnectTimer = null;
        }
        // 停止轮询
        if (this.pollingTimer) {
            clearInterval(this.pollingTimer);
            this.pollingTimer = null;
        }
        this.connected = false;
        this.hasReceivedMessage = false;
    }

    // 处理消息推送
    async _handleMessagePush(push) {
        try {
            console.log('[SSE] Message push received:', push);

            // chat 事件推送的是消息数据
            const message = {
                id: push.message_id,
                conversation_id: push.conversation_id,
                sender_id: push.sender_id,
                receiver_id: push.receiver_id,
                type: push.type,
                content: push.content,
                status: 'delivered',
                created_at: push.created_at
            };

            console.log('[SSE] Calling message.receiveMessage with:', message);
            await this.message.receiveMessage(message);
            console.log('[SSE] Message saved and notified');
        } catch (error) {
            console.error('[SSE] Failed to handle message push:', error);
        }
    }

    // 拉取离线消息（重连后调用）
    async _fetchOfflineMessages() {
        try {
            const currentConversation = this.message.getCurrentConversation();
            if (!currentConversation) {
                console.log('[SSE] No current conversation, skipping offline message fetch');
                return;
            }

            // 获取本地最新的消息ID
            const localMessages = await this.message.getLocalMessages(currentConversation);
            let lastMessageId = 0;
            if (localMessages.length > 0) {
                // 按ID降序排列，获取最大的ID
                localMessages.sort((a, b) => b.id - a.id);
                lastMessageId = localMessages[0].id;
            }

            console.log(`[SSE] Fetching offline messages for conversation ${currentConversation}, last ID: ${lastMessageId}`);

            // 从服务器拉取新消息
            const newMessages = await this.message.loadMessages(currentConversation, lastMessageId, 50);

            console.log(`[SSE] Fetched ${newMessages.length} offline messages`);

            // 通知UI刷新（已经通过receiveMessage保存到本地并通知）
            for (const msg of newMessages) {
                this._notifyMessageReceived(msg);
            }
        } catch (error) {
            console.error('[SSE] Failed to fetch offline messages:', error);
        }
    }

    // 通知消息接收
    _notifyMessageReceived(message) {
        window.dispatchEvent(new CustomEvent('message:received', {
            detail: message
        }));
    }

    // 通知连接状态变化
    _notifyConnectionChange(connected) {
        window.dispatchEvent(new CustomEvent('connection:change', {
            detail: { connected }
        }));
    }

    // 获取连接状态
    isConnected() {
        return this.connected;
    }

    // 获取认证状态
    isAuthenticated() {
        return this.connected; // SSE 连接即已认证
    }

    // 启动轮询（移动端使用）
    _startPolling() {
        if (this.pollingTimer) {
            return; // 已经启动
        }

        // 移动端每3秒轮询一次
        const pollInterval = 3000;
        console.log(`[SSE] Starting polling fallback (interval: ${pollInterval}ms)`);

        this.pollingTimer = setInterval(async () => {
            await this._doPoll();
        }, pollInterval);

        // 立即执行一次轮询
        setTimeout(async () => {
            await this._doPoll();
        }, 1000);
    }

    // 执行轮询
    async _doPoll() {
        try {
            const currentConversation = this.message.getCurrentConversation();
            if (!currentConversation) {
                console.log('[SSE Poll] No current conversation, skipping');
                return;
            }

            // 检查是否需要轮询
            const now = Date.now();
            const timeSinceLastPoll = now - this.lastPollTime;
            const timeSinceLastMessage = this.lastMessageId > 0 ? now - this.lastMessageId * 1000 : 0;

            // 如果最近3秒内轮询过，跳过
            if (timeSinceLastPoll < 3000) {
                return;
            }

            // 如果SSE正常工作且最近5秒内收到过消息，减少轮询频率
            if (this.connected && this.hasReceivedMessage && timeSinceLastMessage < 5000) {
                console.log('[SSE Poll] SSE working well, skipping poll');
                return;
            }

            this.lastPollTime = now;
            console.log('[SSE Poll] Checking for new messages...');

            // 获取本地最新的消息ID
            const localMessages = await this.message.getLocalMessages(currentConversation);
            let lastId = 0;
            if (localMessages.length > 0) {
                localMessages.sort((a, b) => b.id - a.id);
                lastId = localMessages[0].id;
            }

            // 从服务器拉取新消息
            const newMessages = await this.message.loadMessages(currentConversation, lastId, 50);

            if (newMessages.length > 0) {
                console.log(`[SSE Poll] Found ${newMessages.length} new messages`);

                // 更新最后收到的消息ID
                const latestMsg = newMessages[newMessages.length - 1];
                if (latestMsg && latestMsg.id > this.lastMessageId) {
                    this.lastMessageId = latestMsg.id;
                    this.hasReceivedMessage = true;
                }

                // 通知UI刷新
                for (const msg of newMessages) {
                    await this.message.receiveMessage(msg);
                }
            } else {
                console.log('[SSE Poll] No new messages');
            }
        } catch (error) {
            console.error('[SSE Poll] Error:', error);
        }
    }
}
