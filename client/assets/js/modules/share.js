// 分享模块
export class ShareModule {
    constructor(apiClient, store) {
        this.apiClient = apiClient;
        this.store = store;
    }

    // 创建分享
    async createShare(conversationID, options = {}) {
        const {
            expireDays = 7,      // 默认7天
            messageRange = 'recent', // all/recent
            recentCount = 50      // 默认50条
        } = options;

        const response = await this.apiClient.post(`/api/conversations/${conversationID}/share`, {
            expire_days: expireDays,
            message_range: messageRange,
            recent_count: recentCount
        });

        // 返回完整的分享 URL（使用URL参数方式）
        const baseURL = this.getBaseURL();
        return {
            ...response,
            full_url: `${baseURL}/?shared=${response.share_token}`
        };
    }

    // 获取分享内容（公开接口，无需认证）
    async getSharedContent(token, beforeID = 0, limit = 50) {
        let url = `api/shared/${token}?limit=${limit}`;
        if (beforeID > 0) {
            url += `&before_id=${beforeID}`;
        }

        const response = await fetch(url);
        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || `HTTP ${response.status}`);
        }

        return await response.json();
    }

    // 获取我的分享列表
    async getMyShares(page = 1, limit = 20) {
        const response = await this.apiClient.get(`/api/shares?page=${page}&limit=${limit}`);
        const baseURL = this.getBaseURL();
        // 为每个分享项添加完整 URL
        if (response.shares) {
            response.shares.forEach(share => {
                share.full_url = `${baseURL}/?shared=${share.share_token}`;
            });
        }
        return response;
    }

    // 删除分享
    async deleteShare(shareID) {
        return await this.apiClient.delete(`/api/shares/${shareID}`);
    }

    // 复制分享链接到剪贴板
    async copyShareURL(shareURL) {
        try {
            await navigator.clipboard.writeText(shareURL);
            return true;
        } catch (err) {
            // 降级方案：使用传统方法
            const textArea = document.createElement('textarea');
            textArea.value = shareURL;
            textArea.style.position = 'fixed';
            textArea.style.opacity = '0';
            document.body.appendChild(textArea);
            textArea.select();
            try {
                document.execCommand('copy');
                document.body.removeChild(textArea);
                return true;
            } catch (e) {
                document.body.removeChild(textArea);
                return false;
            }
        }
    }

    // 从 URL 获取分享 token
    getShareTokenFromURL() {
        const params = new URLSearchParams(window.location.search);
        return params.get('shared');
    }

    // 获取基础 URL
    getBaseURL() {
        const base = document.querySelector('base')?.href || window.location.origin;
        return base.replace(/\/$/, '');
    }

    // 格式化过期时间
    formatExpireTime(expireAt) {
        if (expireAt === 0) {
            return '永久有效';
        }

        const now = Math.floor(Date.now() / 1000);
        const diff = expireAt - now;

        if (diff <= 0) {
            return '已过期';
        }

        const days = Math.floor(diff / 86400);
        const hours = Math.floor((diff % 86400) / 3600);

        if (days > 0) {
            return `${days}天后过期`;
        } else if (hours > 0) {
            return `${hours}小时后过期`;
        } else {
            return '即将过期';
        }
    }

    // 格式化访问次数
    formatViewCount(count) {
        if (count === 0) {
            return '暂无访问';
        } else if (count === 1) {
            return '1次访问';
        } else {
            return `${count}次访问`;
        }
    }
}
