// 消息模块
export class MessageModule {
    constructor(apiClient, connection, store, auth = null) {
        this.apiClient = apiClient;
        this.connection = connection;
        this.store = store;
        this.auth = auth;
        this.conversations = [];
        this.currentConversation = null;
    }

    // 发送文本消息
    async sendText(conversationId, text) {
        const currentUser = this.auth ? this.auth.getCurrentUser() : null;
        if (!currentUser) {
            throw new Error('用户未登录');
        }

        // 乐观更新：先保存到本地
        const tempId = `temp_${Date.now()}`; // 临时ID，用前缀标识
        const tempMessage = {
            id: tempId,
            conversation_id: conversationId,
            sender_id: currentUser.id,
            receiver_id: null,
            type: 'text',
            content: text,
            status: 'sending',
            created_at: Math.floor(Date.now() / 1000)
        };

        // 保存到本地存储（立即显示）
        await this.store.messages.put(tempMessage);

        // 通知UI更新
        this._notifyMessageReceived(tempMessage);

        try {
            // 通过 HTTP API 发送
            const response = await this.apiClient.post(`/api/conversations/${conversationId}/messages`, {
                type: 'text',
                content: text
            });

            // 删除临时消息，添加真实消息
            await this.store.messages.delete(tempId);
            await this.store.messages.put(response);

            // 通知UI删除临时消息并添加真实消息
            this._notifyMessageDeleted(tempId);
            this._notifyMessageReceived(response);

            console.log('Message sent successfully');
        } catch (error) {
            // 发送失败，更新状态
            tempMessage.status = 'failed';
            await this.store.messages.put(tempMessage);
            this._notifyMessageReceived(tempMessage);
            console.error('Failed to send text message:', error);
            throw error;
        }
    }

    // 发送媒体消息
    async sendMedia(conversationId, mediaId, type) {
        try {
            // 通过 HTTP API 发送
            await this.apiClient.post(`/api/conversations/${conversationId}/messages`, {
                type: type,
                content: mediaId
            });

            console.log('Media message sent successfully');
        } catch (error) {
            console.error('Failed to send media message:', error);
            throw error;
        }
    }

    // 接收消息
    async receiveMessage(payload) {
        console.log('[Message] receiveMessage called with:', payload);

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

        console.log('[Message] Saving to store:', message);

        // 保存到本地
        await this.store.messages.put(message);

        console.log('[Message] Notifying UI');

        // 通知UI更新
        this._notifyMessageReceived(message);

        console.log('[Message] receiveMessage completed');
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

    // 通知消息删除
    _notifyMessageDeleted(messageId) {
        window.dispatchEvent(new CustomEvent('message:deleted', {
            detail: messageId
        }));
    }
}
