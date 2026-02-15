// 连接模块
import { MessageType } from '../protocol/messages.js';
import { encodeMessage, decodeMessage } from '../protocol/codec.js';

export class ConnectionModule {
    constructor(auth, messageModule, config) {
        this.auth = auth;
        this.messageModule = messageModule;
        this.config = config;
        this.ws = null;
        this.connected = false;
        this.authenticated = false;
        this.seq = 0;
        this.pendingAcks = new Map();
        this.reconnectTimer = null;
        this.reconnectAttempts = 0;
        this.pingInterval = null;
    }

    // 连接WebSocket
    connect() {
        const wsUrl = this.config.wsUrl;
        console.log('Connecting to WebSocket:', wsUrl);

        // 如果已有连接，先关闭
        if (this.ws) {
            this.ws.close();
        }

        this.ws = new WebSocket(wsUrl);
        this._setupEventHandlers();
    }

    // 设置事件处理
    _setupEventHandlers() {
        this.ws.onopen = () => this._handleOpen();
        this.ws.onclose = () => this._handleClose();
        this.ws.onerror = (error) => this._handleError(error);
        this.ws.onmessage = (event) => this._handleMessage(event);
    }

    // 连接成功
    async _handleOpen() {
        console.log('[DEBUG] _handleOpen called, this.connected =', this.connected);

        // 防止重复触发
        if (this.connected) {
            console.warn('Already connected, skipping authentication');
            return;
        }

        console.log('[DEBUG] Setting this.connected = true');
        this.connected = true;
        this.reconnectAttempts = 0;
        this._notifyConnectionChange(true);

        // 立即发送认证消息（等待完成）
        console.log('[DEBUG] About to call _authenticate');
        await this._authenticate();
        console.log('[DEBUG] _authenticate completed');
    }

    // 认证
    async _authenticate() {
        console.log('Authenticating with token');
        const payload = {
            token: this.auth.getToken()
        };

        await this.send('MsgAuth', payload);
    }

    // 发送消息
    async send(type, payload) {
        if (!this.connected || !this.authenticated) {
            console.warn('Not connected, cannot send message');
            return;
        }

        const seq = ++this.seq;
        const messageType = MessageType[type];

        if (!messageType) {
            console.error('Unknown message type:', type);
            return;
        }

        const encoded = encodeMessage(messageType, seq, payload);

        try {
            this.ws.send(encoded);
            console.log('Sent message:', type, 'seq:', seq);
        } catch (error) {
            console.error('Failed to send message:', error);
        }

        return seq;
    }

    // 处理接收到的消息
    async _handleMessage(event) {
        try {
            const message = decodeMessage(event.data);
            console.log('Received message type:', message.type, 'seq:', message.seq);

            switch (message.type) {
                case MessageType.MsgAuthRsp:
                    await this._handleAuthResponse(message);
                    break;
                case MessageType.MsgChatPush:
                    await this._handleChatPush(message);
                    break;
                case MessageType.MsgSyncRsp:
                    await this._handleSyncResponse(message);
                    break;
                case MessageType.MsgPresencePush:
                    await this._handlePresencePush(message);
                    break;
                case MessageType.MsgPong:
                    // 心跳响应
                    console.log('Received pong');
                    break;
                case MessageType.MsgError:
                    this._handleErrorMessage(message);
                    break;
                default:
                    console.warn('Unknown message type:', message.type);
            }
        } catch (error) {
            console.error('Error handling message:', error);
        }
    }

    // 处理认证响应
    async _handleAuthResponse(message) {
        const payload = message.payload;
        if (payload.success) {
            console.log('Authentication successful');
            this.authenticated = true;
            this._startPing();
            // 请求同步离线消息
            await this.syncOfflineMessages();
        } else {
            console.error('Authentication failed:', payload.error);
            this.ws.close();
        }
    }

    // 处理聊天消息推送
    async _handleChatPush(message) {
        const payload = message.payload;
        await this.messageModule.receiveMessage(payload);

        // 发送确认
        await this.send('MsgAck', {
            message_id: payload.id,
            status: 'delivered'
        });
    }

    // 处理同步响应
    async _handleSyncResponse(message) {
        console.log('Sync response received:', message.payload);
    }

    // 处理在线状态推送
    async _handlePresencePush(message) {
        console.log('Presence push:', message.payload);
    }

    // 处理错误消息
    _handleErrorMessage(message) {
        console.error('Server error:', message.payload);
    }

    // 同步离线消息
    async syncOfflineMessages() {
        const lastMessage = await this.store.messages.getLast();
        const payload = {
            last_message_id: lastMessage ? lastMessage.id : 0,
            last_sync_time: 0
        };
        await this.send('MsgSyncReq', payload);
    }

    // 断开连接
    disconnect() {
        if (this.ws) {
            this.ws.close();
        }
        if (this.pingInterval) {
            clearInterval(this.pingInterval);
            this.pingInterval = null;
        }
    }

    // 处理关闭
    _handleClose() {
        console.log('WebSocket closed');
        this.connected = false;
        this.authenticated = false;
        this._notifyConnectionChange(false);

        // 尝试重连
        this._scheduleReconnect();
    }

    // 处理错误
    _handleError(error) {
        console.error('WebSocket error:', error);
    }

    // 安排重连
    _scheduleReconnect() {
        const delay = Math.min(
            1000 * Math.pow(1.5, this.reconnectAttempts),
            30000
        );

        console.log(`Scheduling reconnect in ${delay}ms (attempt ${this.reconnectAttempts + 1})`);

        this.reconnectTimer = setTimeout(() => {
            this.reconnectAttempts++;
            this.connect();
        }, delay);
    }

    // 发送心跳
    _startPing() {
        this.pingInterval = setInterval(() => {
            if (this.connected && this.authenticated) {
                this.send('MsgPing', {});
            }
        }, 30000); // 30秒
    }

    // 通知连接状态变化
    _notifyConnectionChange(connected) {
        window.dispatchEvent(new CustomEvent('connection:change', {
            detail: { connected }
        }));
    }

    // 获取连接状态
    isConnected() {
        return this.connected && this.authenticated;
    }
}
