package ws

import "zmessage/server/pkg/protocol"

// Connection WebSocket连接接口
type Connection interface {
	// ID 获取连接ID
	ID() string
	// UserID 获取用户ID
	UserID() int64
	// Send 发送消息
	Send(msg *protocol.WSMessage) error
	// authenticate 认证连接
	authenticate(userID int64) error
}

// Manager 连接管理器接口
type Manager interface {
	// HandleConnection 处理新的WebSocket连接
	HandleConnection(conn interface{})
	// GetConnection 获取用户的连接（返回第一个）
	GetConnection(userID int64) Connection
	// GetConnections 获取用户的所有连接
	GetConnections(userID int64) []Connection
	// BroadcastToUser 向用户的所有连接发送消息
	BroadcastToUser(userID int64, msg *protocol.WSMessage) error
	// IsOnline 检查用户是否在线
	IsOnline(userID int64) bool
	// GetOnlineUsers 获取在线用户列表
	GetOnlineUsers() []int64
	// Disconnect 断开指定连接
	Disconnect(connID string)
	// DisconnectUser 断开用户的所有连接
	DisconnectUser(userID int64)
}

// MessageHandler 消息处理器接口
type MessageHandler interface {
	// HandleMessage 处理消息
	HandleMessage(conn Connection, msg *protocol.WSMessage) error
}
