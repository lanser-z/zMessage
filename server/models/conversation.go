package models

// Conversation 会话模型
type Conversation struct {
	ID        int64 `json:"id"`
	UserAID   int64 `json:"user_a_id"`
	UserBID   int64 `json:"user_b_id"`
	CreatedAt int64 `json:"created_at"`
	UpdatedAt int64 `json:"updated_at"`
}

// ConversationWithInfo 带额外信息的会话
type ConversationWithInfo struct {
	ID           int64       `json:"id"`
	UserAID      int64       `json:"user_a_id"`
	UserBID      int64       `json:"user_b_id"`
	Participant  *UserInfo   `json:"participant"`
	LastMessage  *Message    `json:"last_message,omitempty"`
	UnreadCount  int         `json:"unread_count"`
	CreatedAt    int64       `json:"created_at"`
	UpdatedAt    int64       `json:"updated_at"`
}

// UserInfo 用户简要信息
type UserInfo struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar,omitempty"`
	Online   bool   `json:"online"`
	LastSeen int64  `json:"last_seen"`
}
