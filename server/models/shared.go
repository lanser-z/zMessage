package models

// SharedConversation 分享的会话模型
type SharedConversation struct {
	ID             int64  `json:"id"`
	ConversationID int64  `json:"conversation_id"`
	ShareToken     string `json:"share_token"`
	CreatedBy      int64  `json:"created_by"`
	ExpireAt       int64  `json:"expire_at"`
	CreatedAt      int64  `json:"created_at"`

	// 扩展字段
	FirstMessageID int64 `json:"first_message_id"` // 分享起始消息ID（0表示从最早开始）
	LastMessageID  int64 `json:"last_message_id"`  // 分享结束消息ID（0表示到最新）
	MessageCount   int   `json:"message_count"`    // 分享的消息数量
	ViewCount      int   `json:"view_count"`       // 访问次数
}

// ShareRequest 创建分享请求
type ShareRequest struct {
	ConversationID int64  `json:"conversation_id" binding:"required"`
	ExpireDays     int    `json:"expire_days"`      // 过期天数，0=永久，默认7
	MessageRange   string `json:"message_range"`    // "all", "recent"
	RecentCount    int    `json:"recent_count"`     // 最近N条，默认50
}

// ParticipantInfo 参与者信息（脱敏）
type ParticipantInfo struct {
	ID       int64  `json:"id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar,omitempty"`
}

// SharedConversationDetail 分享详情（用于展示）
type SharedConversationDetail struct {
	ID             int64              `json:"id"`
	ShareToken     string             `json:"share_token"`
	ConversationID int64              `json:"conversation_id"`
	Participants   []ParticipantInfo  `json:"participants"` // 脱敏的参与者信息
	MessageCount   int                `json:"message_count"`
	FirstMessageID int64              `json:"first_message_id"`
	LastMessageID  int64              `json:"last_message_id"`
	ExpireAt       int64              `json:"expire_at"`
	CreatedAt      int64              `json:"created_at"`
	ViewCount      int                `json:"view_count"`
	IsExpired      bool               `json:"is_expired"`

	// 分享者信息
	CreatedByNickname string `json:"created_by_nickname,omitempty"`
}

// ShareListItem 分享列表项
type ShareListItem struct {
	ID             int64             `json:"id"`
	ShareToken     string            `json:"share_token"`
	ShareURL       string            `json:"share_url"`
	ConversationID int64             `json:"conversation_id"`
	Participant    ParticipantInfo   `json:"participant"`
	MessageCount   int               `json:"message_count"`
	ExpireAt       int64             `json:"expire_at"`
	CreatedAt      int64             `json:"created_at"`
	ViewCount      int               `json:"view_count"`
	IsExpired      bool              `json:"is_expired"`
}

// SharedMessage 分享的消息（脱敏）
type SharedMessage struct {
	ID             int64  `json:"id"`
	SenderID       int64  `json:"sender_id"`
	SenderNickname string `json:"sender_nickname"`
	Type           string `json:"type"`
	Content        string `json:"content"`
	CreatedAt      int64  `json:"created_at"`
}
