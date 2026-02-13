package api

import (
	"zmessage/server/pkg/protocol"
	"zmessage/server/ws"
)

// WSAdapter WebSocket管理器适配器
type WSAdapter struct {
	mgr ws.Manager
}

// NewWSAdapter 创建WebSocket适配器
func NewWSAdapter(mgr ws.Manager) WSManager {
	return &WSAdapter{mgr: mgr}
}

// HandleConnection 处理WebSocket连接
func (a *WSAdapter) HandleConnection(conn interface{}) {
	a.mgr.HandleConnection(conn)
}

// BroadcastToUser 向用户广播消息
func (a *WSAdapter) BroadcastToUser(userID int64, message interface{}) error {
	wsMsg, ok := message.(*protocol.WSMessage)
	if !ok {
		return nil // 忽略非WSMessage类型
	}
	return a.mgr.BroadcastToUser(userID, wsMsg)
}

// IsOnline 检查用户是否在线
func (a *WSAdapter) IsOnline(userID int64) bool {
	return a.mgr.IsOnline(userID)
}
