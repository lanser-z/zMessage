package models

// Message 消息模型
type Message struct {
	ID             int64  `json:"id"`
	ConversationID int64  `json:"conversation_id"`
	SenderID       int64  `json:"sender_id"`
	ReceiverID     int64  `json:"receiver_id"`
	Type           string `json:"type"` // text, voice, image
	Content        string `json:"content"`
	Status         string `json:"status"` // sent, delivered, read
	CreatedAt      int64  `json:"created_at"`
	SyncedAt       *int64 `json:"synced_at,omitempty"`
}
