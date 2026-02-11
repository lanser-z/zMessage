package models

// SharedConversation 分享的会话模型
type SharedConversation struct {
	ID             int64  `json:"id"`
	ConversationID int64  `json:"conversation_id"`
	ShareToken     string `json:"share_token"`
	CreatedBy      int64  `json:"created_by"`
	ExpireAt       int64  `json:"expire_at"`
	CreatedAt      int64  `json:"created_at"`
}
