// 认证模块
import { ApiClient } from '../utils/request.js';

export class AuthModule {
    constructor(apiClient, store) {
        this.apiClient = apiClient;
        this.store = store;
        this.currentUser = null;
        this.token = null;
    }

    // 注册
    async register(username, password, nickname) {
        const response = await this.apiClient.post('/api/auth/register', {
            username,
            password,
            nickname: nickname || username
        });
        await this._setAuth(response.user, response.token);
        return response.user;
    }

    // 登录
    async login(username, password) {
        console.log('[AUTH] Login attempt for:', username);
        const response = await this.apiClient.post('/api/auth/login', {
            username,
            password
        });
        console.log('[AUTH] Login response:', response);
        await this._setAuth(response.user, response.token);
        console.log('[AUTH] Auth set successfully');
        return response.user;
    }

    // 登出
    async logout() {
        await this._clearAuth();
    }

    // 获取当前用户
    getCurrentUser() {
        return this.currentUser;
    }

    // 获取Token
    getToken() {
        return this.token;
    }

    // 检查是否已登录
    isAuthenticated() {
        return this.token !== null;
    }

    // 内部方法
    async _setAuth(user, token) {
        this.currentUser = user;
        this.token = token;
        this.apiClient.setToken(token);
        await this.store.auth.set(user, token);
    }

    async _clearAuth() {
        this.currentUser = null;
        this.token = null;
        this.apiClient.setToken(null);
        await this.store.auth.clear();
    }

    // 从本地恢复会话
    async restoreSession() {
        const authData = await this.store.auth.get();
        if (authData && authData.user && authData.token) {
            this.currentUser = authData.user;
            this.token = authData.token;
            this.apiClient.setToken(this.token);
            return true;
        }
        return false;
    }

    // 获取当前用户信息
    async getCurrentUserInfo() {
        const response = await this.apiClient.get('/api/users/me');
        this.currentUser = response;
        return response;
    }
}
