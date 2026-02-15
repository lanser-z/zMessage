// 主入口
import { ApiClient } from './utils/request.js';
import { Store } from './store/index.js';
import { AuthModule } from './modules/auth.js';
import { ConnectionModule } from './modules/connection_sse.js';  // SSE 版本
import { MessageModule } from './modules/message.js';
import { MediaModule } from './modules/media.js';
import { UIModule } from './modules/ui.js';

// 配置
const config = {
    apiBaseUrl: 'http://203.83.228.222:9405'
};

// 应用状态
let auth, connection, message, media, ui, store, apiClient;

// 初始化应用
async function init() {
    try {
        // 初始化存储
        store = new Store();
        await store.init();

        // 初始化API客户端
        apiClient = new ApiClient(config.apiBaseUrl);

        // 初始化业务模块
        auth = new AuthModule(apiClient, store);
        message = new MessageModule(apiClient, null, store, auth);
        media = new MediaModule(apiClient, store);
        connection = new ConnectionModule(auth, message, config);

        // 设置连接模块到消息模块
        message.connection = connection;

        // 初始化UI
        ui = new UIModule(auth, connection, message, media, apiClient);
        ui.init();

        console.log('Application initialized');
    } catch (error) {
        console.error('Failed to initialize application:', error);
        document.getElementById('app').innerHTML = `
            <div class="error">
                <h2>初始化失败</h2>
                <p>${error.message}</p>
                <button onclick="location.reload()">重试</button>
            </div>
        `;
    }
}

// 启动应用
init();
