// HTTP请求封装
export class ApiClient {
    constructor(baseURL) {
        this.baseURL = baseURL;
        this.token = null;
    }

    setToken(token) {
        this.token = token;
    }

    getToken() {
        return this.token;
    }

    async request(method, path, data = null, options = {}) {
        // 拼接 URL
        let url = this.baseURL + path;

        // 如果 baseURL 为空，去掉 path 开头的 / 使其成为相对路径
        // 这样可以受 index.html 中 <base href> 的控制
        // 开发环境: <base href="/"> → /api/...
        // 生产环境: <base href="/zmessage/"> → /zmessage/api/...
        if (!this.baseURL && path.startsWith('/')) {
            url = path.substring(1);
        }
        const headers = {
            'Content-Type': 'application/json',
            ...options.headers
        };

        if (this.token) {
            headers['Authorization'] = `Bearer ${this.token}`;
        }

        const config = {
            method,
            headers
        };

        if (data) {
            if (headers['Content-Type'] === 'multipart/form-data') {
                // FormData 不需要设置 Content-Type
                delete headers['Content-Type'];
                config.body = data;
            } else {
                config.body = JSON.stringify(data);
            }
        }

        const response = await fetch(url, config);

        const result = await response.json();

        if (!response.ok) {
            throw new Error(result.error || `HTTP ${response.status}: ${response.statusText}`);
        }

        return result.data || result;
    }

    get(path, options) {
        return this.request('GET', path, null, options);
    }

    post(path, data, options) {
        return this.request('POST', path, data, options);
    }

    put(path, data, options) {
        return this.request('PUT', path, data, options);
    }

    delete(path, options) {
        return this.request('DELETE', path, null, options);
    }
}
