// 消息模块
export class MessageModule {
    constructor(apiClient, connection, store) {
        this.apiClient = apiClient;
        this.connection = connection;
        this.store = store;
        this.conversations = [];
        this.currentConversation = null;
    }

    // 发送文本消息
    async sendText(conversationId, text) {
        const message = {
            conversation_id: conversationId,
            type: 'text',
            content: text,
            sender_id: this.apiClient.token ? null : 0, // 临时，需要从token解析
            created_at: Math.floor(Date.now() / 1000),
            status: 'sending'
        };

        try {
            // 保存到本地 (状态为sending)
            await this.store.messages.add(message);

            // 通过WebSocket发送
            const conv = this.conversations.find(c => c.id === conversationId);
            if (conv) {
                const payload = {
                    to: conv.participant.id,
                    type: 'text',
                    content: text
                };

                await this.connection.send('MsgChat', payload);
            }

            return message;
        } catch (error) {
            console.error('Failed to send text message:', error);
            throw error;
        }
    }

    // 发送媒体消息
    async sendMedia(conversationId, mediaId, type) {
        const message = {
            conversation_id: conversationId,
            type: type,
            content: mediaId,
            created_at: Math.floor(Date.now() / 1000),
            status: 'sending'
        };

        try {
            await this.store.messages.add(message);

            const conv = this.conversations.find(c => c.id === conversationId);
            if (conv) {
                const payload = {
                    to: conv.participant.id,
                    type: type,
                    content: mediaId
                };

                await this.connection.send('MsgChat', payload);
            }

            return message;
        } catch (error) {
            console.error('Failed to send media message:', error);
            throw error;
        }
    }

    // 接收消息
    async receiveMessage(payload) {
        const message = {
            id: payload.id,
            conversation_id: payload.conversation_id,
            sender_id: payload.sender_id,
            receiver_id: payload.receiver_id,
            type: payload.type,
            content: payload.content,
            status: payload.status || 'delivered',
            created_at: payload.created_at
        };

        // 保存到本地
        await this.store.messages.put(message);

        // 通知UI更新
        this._notifyMessageReceived(message);
    }

    // 获取会话列表
    async loadConversations() {
        const response = await this.apiClient.get('/api/conversations');
        this.conversations = response.conversations || response || [];

        // 缓存到本地
        for (const conv of this.conversations) {
            await this.store.conversations.put(conv);
        }

        return this.conversations;
    }

    // 获取会话消息
    async loadMessages(conversationId, beforeId = null, limit = 50) {
        let url = `/api/conversations/${conversationId}/messages?limit=${limit}`;
        if (beforeId) {
            url += `&before_id=${beforeId}`;
        }

        const response = await this.apiClient.get(url);
        const messages = response.messages || response || [];

        // 保存到本地
        for (const msg of messages) {
            await this.store.messages.put(msg);
        }

        return messages;
    }

    // 标记已读
    async markAsRead(conversationId) {
        await this.apiClient.post(`/api/conversations/${conversationId}/read`);

        // 更新本地会话
        const conv = this.conversations.find(c => c.id === conversationId);
        if (conv) {
            conv.unread_count = 0;
            await this.store.conversations.put(conv);
        }
    }

    // 获取或创建与用户的会话
    async getConversationWithUser(userId) {
        const response = await this.apiClient.get(`/api/conversations/with/${userId}`);

        // 添加到会话列表
        const existing = this.conversations.find(c => c.id === response.id);
        if (!existing) {
            this.conversations.unshift(response);
            await this.store.conversations.put(response);
        }

        return response;
    }

    // 获取本地缓存的消息
    async getLocalMessages(conversationId) {
        return await this.store.messages.getByConversation(conversationId);
    }

    // 获取本地缓存的会话
    async getLocalConversations() {
        return await this.store.conversations.getAll();
    }

    // 事件通知
    _notifyMessageReceived(message) {
        // 触发自定义事件
        window.dispatchEvent(new CustomEvent('message:received', {
            detail: message
        }));
    }

    // 获取当前会话
    getCurrentConversation() {
        return this.currentConversation;
    }

    // 设置当前会话
    setCurrentConversation(conversationId) {
        this.currentConversation = conversationId;
    }
}
