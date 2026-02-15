// IndexedDB 封装
export class Store {
    constructor() {
        this.dbName = 'zMessage';
        this.dbVersion = 1;
        this.db = null;
    }

    async init() {
        return new Promise((resolve, reject) => {
            const request = indexedDB.open(this.dbName, this.dbVersion);

            request.onerror = () => reject(request.error);
            request.onsuccess = () => {
                this.db = request.result;
                resolve();
            };

            request.onupgradeneeded = (event) => {
                const db = event.target.result;

                // 用户数据存储
                if (!db.objectStoreNames.contains('auth')) {
                    db.createObjectStore('auth', { keyPath: 'id' });
                }

                // 消息存储
                if (!db.objectStoreNames.contains('messages')) {
                    const msgStore = db.createObjectStore('messages', { keyPath: 'id', autoIncrement: true });
                    msgStore.createIndex('conversation_id', 'conversation_id', { unique: false });
                }

                // 会话存储
                if (!db.objectStoreNames.contains('conversations')) {
                    db.createObjectStore('conversations', { keyPath: 'id', autoIncrement: true });
                }

                // 媒体缓存
                if (!db.objectStoreNames.contains('media')) {
                    db.createObjectStore('media', { keyPath: 'id', autoIncrement: true });
                }
            };
        });
    }

    // 认证数据
    get auth() {
        return new AuthStore(this.db);
    }

    // 消息数据
    get messages() {
        return new MessageStore(this.db);
    }

    // 会话数据
    get conversations() {
        return new ConversationStore(this.db);
    }

    // 媒体数据
    get media() {
        return new MediaStore(this.db);
    }
}

// 认证存储
class AuthStore {
    constructor(db) {
        this.db = db;
    }

    async set(user, token) {
        const tx = this.db.transaction(['auth'], 'readwrite');
        const store = tx.objectStore('auth');
        await new Promise((resolve, reject) => {
            const request = store.put({ id: 'current', user, token });
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }

    async get() {
        const tx = this.db.transaction(['auth'], 'readonly');
        const store = tx.objectStore('auth');
        return new Promise((resolve, reject) => {
            const request = store.get('current');
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async clear() {
        const tx = this.db.transaction(['auth'], 'readwrite');
        const store = tx.objectStore('auth');
        await new Promise((resolve, reject) => {
            const request = store.delete('current');
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }
}

// 消息存储
class MessageStore {
    constructor(db) {
        this.db = db;
    }

    async add(message) {
        const tx = this.db.transaction(['messages'], 'readwrite');
        const store = tx.objectStore('messages');
        return new Promise((resolve, reject) => {
            const request = store.add(message);
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async delete(id) {
        const tx = this.db.transaction(['messages'], 'readwrite');
        const store = tx.objectStore('messages');
        return new Promise((resolve, reject) => {
            const request = store.delete(id);
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }

    async put(message) {
        const tx = this.db.transaction(['messages'], 'readwrite');
        const store = tx.objectStore('messages');
        return new Promise((resolve, reject) => {
            const request = store.put(message);
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }

    async getByConversation(conversationId) {
        return new Promise((resolve, reject) => {
            const tx = this.db.transaction(['messages'], 'readonly');
            const store = tx.objectStore('messages');
            const index = store.index('conversation_id');
            const request = index.getAll(conversationId);

            request.onsuccess = () => resolve(request.result || []);
            request.onerror = () => reject(request.error);
        });
    }

    async getLast() {
        return new Promise((resolve, reject) => {
            const tx = this.db.transaction(['messages'], 'readonly');
            const store = tx.objectStore('messages');
            const request = store.openCursor(null, 'prev');

            request.onsuccess = () => {
                const cursor = request.result;
                resolve(cursor ? cursor.value : null);
            };
            request.onerror = () => reject(request.error);
        });
    }

    async getAll() {
        return new Promise((resolve, reject) => {
            const tx = this.db.transaction(['messages'], 'readonly');
            const store = tx.objectStore('messages');
            const request = store.getAll();

            request.onsuccess = () => resolve(request.result || []);
            request.onerror = () => reject(request.error);
        });
    }
}

// 会话存储
class ConversationStore {
    constructor(db) {
        this.db = db;
    }

    async add(conversation) {
        const tx = this.db.transaction(['conversations'], 'readwrite');
        const store = tx.objectStore('conversations');
        return new Promise((resolve, reject) => {
            const request = store.add(conversation);
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async put(conversation) {
        const tx = this.db.transaction(['conversations'], 'readwrite');
        const store = tx.objectStore('conversations');
        return new Promise((resolve, reject) => {
            const request = store.put(conversation);
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }

    async getAll() {
        return new Promise((resolve, reject) => {
            const tx = this.db.transaction(['conversations'], 'readonly');
            const store = tx.objectStore('conversations');
            const request = store.getAll();

            request.onsuccess = () => resolve(request.result || []);
            request.onerror = () => reject(request.error);
        });
    }

    async clear() {
        const tx = this.db.transaction(['conversations'], 'readwrite');
        const store = tx.objectStore('conversations');
        return new Promise((resolve, reject) => {
            const request = store.clear();
            request.onsuccess = () => resolve();
            request.onerror = () => reject(request.error);
        });
    }
}

// 媒体存储
class MediaStore {
    constructor(db) {
        this.db = db;
    }

    async add(media) {
        const tx = this.db.transaction(['media'], 'readwrite');
        const store = tx.objectStore('media');
        return new Promise((resolve, reject) => {
            const request = store.add(media);
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async get(id) {
        const tx = this.db.transaction(['media'], 'readonly');
        const store = tx.objectStore('media');
        return new Promise((resolve, reject) => {
            const request = store.get(id);
            request.onsuccess = () => resolve(request.result);
            request.onerror = () => reject(request.error);
        });
    }

    async getAll() {
        return new Promise((resolve, reject) => {
            const tx = this.db.transaction(['media'], 'readonly');
            const store = tx.objectStore('media');
            const request = store.getAll();

            request.onsuccess = () => resolve(request.result || []);
            request.onerror = () => reject(request.error);
        });
    }
}
