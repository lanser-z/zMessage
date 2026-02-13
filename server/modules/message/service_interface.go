package message

import (
	"zmessage/server/models"
)

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	From    int64  `json:"from"`
	To      int64  `json:"to"`
	Type    string `json:"type"`    // text, voice, image
	Content string `json:"content"` // 文本内容或媒体ID
}

// ConversationListRequest 获取会话列表请求
type ConversationListRequest struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

// MessageListRequest 获取消息历史请求
type MessageListRequest struct {
	ConversationID int64 `json:"conversation_id"`
	BeforeID       int64 `json:"before_id"` // 获取此ID之前的消息
	Limit          int   `json:"limit"`
}

// Service 消息服务接口
type Service interface {
	// SendMessage 发送消息
	SendMessage(req *SendMessageRequest) (*models.Message, error)

	// GetConversation 获取会话详情
	GetConversation(id int64, userID int64) (*models.ConversationWithInfo, error)

	// GetConversationWithUser 获取与指定用户的会话（如不存在则创建）
	GetConversationWithUser(userID, otherUserID int64) (*models.ConversationWithInfo, error)

	// GetConversations 获取会话列表
	GetConversations(userID int64, page, limit int) ([]*models.ConversationWithInfo, int, error)

	// GetMessages 获取消息历史
	GetMessages(conversationID int64, userID int64, beforeID int64, limit int) ([]*models.Message, bool, error)

	// MarkAsRead 标记会话消息为已读
	MarkAsRead(conversationID int64, userID int64) error

	// UpdateMessageStatus 更新消息状态
	UpdateMessageStatus(messageID int64, status string) error

	// GetOfflineMessages 获取离线消息
	GetOfflineMessages(userID int64, lastMessageID int64, limit int) ([]*models.Message, error)

	// GetUnreadCount 获取未读消息数
	GetUnreadCount(conversationID int64, userID int64) (int, error)
}
