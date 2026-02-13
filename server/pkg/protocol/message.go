package protocol

// MessageType WebSocket消息类型
type MessageType int16

// 客户端 → 服务端
const (
	MsgAuth     MessageType = 1   // 认证
	MsgChat     MessageType = 2   // 聊天消息
	MsgAck      MessageType = 3   // 确认
	MsgSyncReq  MessageType = 4   // 同步请求
	MsgPresence MessageType = 5   // 在线状态
	MsgPing     MessageType = 6   // 心跳
)

// 服务端 → 客户端
const (
	MsgAuthRsp     MessageType = 101 // 认证响应
	MsgChatPush    MessageType = 102 // 消息推送
	MsgSyncRsp     MessageType = 103 // 同步响应
	MsgPresencePush MessageType = 104 // 在线状态推送
	MsgPong        MessageType = 105 // 心跳响应
	MsgError       MessageType = 106 // 错误通知
)

// WSMessage WebSocket消息
type WSMessage struct {
	Type    MessageType `msgpack:"type"`
	Seq     int64       `msgpack:"seq"`
	Payload []byte      `msgpack:"payload"`
}

// MessageStatus 消息状态
type MessageStatus string

const (
	StatusSent      MessageStatus = "sent"
	StatusDelivered MessageStatus = "delivered"
	StatusRead      MessageStatus = "read"
)

// AuthPayload 认证请求负载
type AuthPayload struct {
	Token string `msgpack:"token"`
}

// AuthResponsePayload 认证响应负载
type AuthResponsePayload struct {
	Success bool   `msgpack:"success"`
	UserID  int64  `msgpack:"user_id,omitempty"`
	Error   string `msgpack:"error,omitempty"`
}

// ChatPayload 聊天消息负载
type ChatPayload struct {
	To      int64  `msgpack:"to"`        // 接收者ID
	Type    string `msgpack:"type"`      // text/voice/image
	Content string `msgpack:"content"`   // 消息内容或媒体ID
}

// ChatPushPayload 消息推送负载
type ChatPushPayload struct {
	MessageID   int64  `msgpack:"message_id"`
	From        int64  `msgpack:"from"`
	To          int64  `msgpack:"to"`
	Type        string `msgpack:"type"`
	Content     string `msgpack:"content"`
	CreatedAt   int64  `msgpack:"created_at"`
}

// AckPayload 确认负载
type AckPayload struct {
	MessageID int64  `msgpack:"message_id"` // 消息ID
	Status    string `msgpack:"status"`     // delivered/read
}

// SyncRequestPayload 同步请求负载
type SyncRequestPayload struct {
	LastMessageID int64 `msgpack:"last_message_id"` // 最后消息ID
	LastSyncTime  int64 `msgpack:"last_sync_time"`  // 最后同步时间
}

// SyncResponsePayload 同步响应负载
type SyncResponsePayload struct {
	Messages []ChatPushPayload `msgpack:"messages"` // 离线消息列表
	HasMore  bool                 `msgpack:"has_more"` // 是否还有更多
}

// PresencePayload 在线状态负载
type PresencePayload struct {
	Status string `msgpack:"status"` // online/offline/away
}

// PresencePushPayload 在线状态推送负载
type PresencePushPayload struct {
	UserID int64  `msgpack:"user_id"`
	Status string `msgpack:"status"`
}

// ErrorPayload 错误通知负载
type ErrorPayload struct {
	Code    string `msgpack:"code"`    // 错误码
	Message string `msgpack:"message"` // 错误信息
}
