package dal

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// schema 数据库Schema
var schema = `
-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    nickname TEXT NOT NULL,
    avatar_id INTEGER,
    created_at INTEGER NOT NULL,
    last_seen INTEGER NOT NULL,
    FOREIGN KEY (avatar_id) REFERENCES media(id)
);

-- 会话表
CREATE TABLE IF NOT EXISTS conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_a_id INTEGER NOT NULL,
    user_b_id INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    FOREIGN KEY (user_a_id) REFERENCES users(id),
    FOREIGN KEY (user_b_id) REFERENCES users(id),
    UNIQUE(user_a_id, user_b_id)
);

-- 消息表
CREATE TABLE IF NOT EXISTS messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL,
    sender_id INTEGER NOT NULL,
    receiver_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'sent',
    created_at INTEGER NOT NULL,
    synced_at INTEGER,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id),
    FOREIGN KEY (sender_id) REFERENCES users(id),
    FOREIGN KEY (receiver_id) REFERENCES users(id)
);

-- 媒体文件表
CREATE TABLE IF NOT EXISTS media (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id INTEGER NOT NULL,
    type TEXT NOT NULL,
    original_path TEXT NOT NULL,
    thumbnail_path TEXT,
    size INTEGER NOT NULL,
    mime_type TEXT NOT NULL,
    width INTEGER,
    height INTEGER,
    duration INTEGER,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (owner_id) REFERENCES users(id)
);

-- 分享表
CREATE TABLE IF NOT EXISTS shared_conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL,
    share_token TEXT UNIQUE NOT NULL,
    created_by INTEGER NOT NULL,
    expire_at INTEGER NOT NULL,
    created_at INTEGER NOT NULL,
    first_message_id INTEGER DEFAULT 0,
    last_message_id INTEGER DEFAULT 0,
    message_count INTEGER DEFAULT 0,
    view_count INTEGER DEFAULT 0,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

-- 索引
CREATE INDEX IF NOT EXISTS idx_messages_conversation ON messages(conversation_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_messages_receiver_status ON messages(receiver_id, status) WHERE status != 'read';
CREATE INDEX IF NOT EXISTS idx_conversations_users ON conversations(user_a_id, user_b_id);
CREATE INDEX IF NOT EXISTS idx_conversations_updated ON conversations(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_media_owner ON media(owner_id);
CREATE INDEX IF NOT EXISTS idx_media_type ON media(type);
`

// manager 数据库管理器实现
type manager struct {
	db    *sql.DB
	user  UserDAL
	conv  ConversationDAL
	msg   MessageDAL
	media MediaDAL
	shared SharedConversationDAL
}

// NewManager 创建数据库管理器
func NewManager(dataDir string) (Manager, error) {
	// 确保数据目录存在
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("create data directory: %w", err)
	}

	// 打开数据库连接
	dbPath := filepath.Join(dataDir, "messages.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	// 设置连接池参数
	db.SetMaxOpenConns(1) // SQLite写操作需要单连接
	db.SetMaxIdleConns(1)

	// 执行Schema
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("create schema: %w", err)
	}

	// 启用外键约束
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	m := &manager{
		db:    db,
		user:  NewUserDAL(db),
		conv:  NewConversationDAL(db),
		msg:   NewMessageDAL(db),
		media: NewMediaDAL(db),
		shared: NewSharedConversationDAL(db),
	}

	return m, nil
}

// DB 获取数据库连接
func (m *manager) DB() DB {
	return m.db
}

// User 用户数据访问
func (m *manager) User() UserDAL {
	return m.user
}

// Conversation 会话数据访问
func (m *manager) Conversation() ConversationDAL {
	return m.conv
}

// Message 消息数据访问
func (m *manager) Message() MessageDAL {
	return m.msg
}

// Media 媒体数据访问
func (m *manager) Media() MediaDAL {
	return m.media
}

// SharedConversation 分享数据访问
func (m *manager) SharedConversation() SharedConversationDAL {
	return m.shared
}

// Close 关闭数据库连接
func (m *manager) Close() error {
	return m.db.Close()
}
