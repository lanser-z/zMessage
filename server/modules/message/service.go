package message

import (
	"fmt"
	"time"

	"zmessage/server/dal"
	"zmessage/server/models"
)

// NewService 创建消息服务
func NewService(dalMgr dal.Manager) Service {
	return &service{
		dal: dalMgr,
	}
}

// service 消息服务实现
type service struct {
	dal dal.Manager
}

// SendMessage 发送消息
func (s *service) SendMessage(req *SendMessageRequest) (*models.Message, error) {
	// 验证发送者和接收者
	if req.From == req.To {
		return nil, ErrSendToSelf
	}

	// 验证消息类型
	if !isValidMessageType(req.Type) {
		return nil, ErrInvalidMessageType
	}

	// 验证消息内容
	if req.Content == "" {
		return nil, ErrInvalidMessageContent
	}

	// 验证用户存在
	sender, err := s.dal.User().GetByID(req.From)
	if err != nil || sender == nil {
		return nil, ErrUserNotFound
	}

	receiver, err := s.dal.User().GetByID(req.To)
	if err != nil || receiver == nil {
		return nil, ErrUserNotFound
	}

	// 获取或创建会话
	conv, err := s.getOrCreateConversation(req.From, req.To)
	if err != nil {
		return nil, fmt.Errorf("get or create conversation: %w", err)
	}

	// 创建消息
	now := time.Now().Unix()
	msg := &models.Message{
		ConversationID: conv.ID,
		SenderID:      req.From,
		ReceiverID:    req.To,
		Type:          req.Type,
		Content:       req.Content,
		Status:        "sent",
		CreatedAt:     now,
	}

	if err := s.dal.Message().Create(msg); err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	// 更新会话时间
	if err := s.dal.Conversation().UpdateTime(conv.ID, now); err != nil {
		return nil, fmt.Errorf("update conversation time: %w", err)
	}

	return msg, nil
}

// GetConversation 获取会话详情
func (s *service) GetConversation(id int64, userID int64) (*models.ConversationWithInfo, error) {
	conv, err := s.dal.Conversation().GetByID(id)
	if err != nil {
		return nil, ErrConversationNotFound
	}

	// 验证用户是会话参与者
	if conv.UserAID != userID && conv.UserBID != userID {
		return nil, ErrAccessDenied
	}

	// 获取对方用户ID
	var participantID int64
	if conv.UserAID == userID {
		participantID = conv.UserBID
	} else {
		participantID = conv.UserAID
	}

	// 获取参与者信息
	user, err := s.dal.User().GetByID(participantID)
	if err != nil {
		return nil, fmt.Errorf("get participant: %w", err)
	}

	// 获取最后一条消息
	lastMsg, _ := s.dal.Message().GetByConversation(conv.ID, 0, 1)
	var lastMessage *models.Message
	if len(lastMsg) > 0 {
		lastMessage = lastMsg[0]
	}

	// 获取未读数
	unreadCount, _ := s.dal.Message().CountUnread(conv.ID, userID)

	return &models.ConversationWithInfo{
		ID:          conv.ID,
		UserAID:     conv.UserAID,
		UserBID:     conv.UserBID,
		Participant:  toUserInfo(user),
		LastMessage:  lastMessage,
		UnreadCount:  unreadCount,
		CreatedAt:    conv.CreatedAt,
		UpdatedAt:    conv.UpdatedAt,
	}, nil
}

// GetConversationWithUser 获取与指定用户的会话
func (s *service) GetConversationWithUser(userID, otherUserID int64) (*models.ConversationWithInfo, error) {
	// 获取或创建会话
	conv, err := s.getOrCreateConversation(userID, otherUserID)
	if err != nil {
		return nil, fmt.Errorf("get or create conversation: %w", err)
	}

	return s.GetConversation(conv.ID, userID)
}

// GetConversations 获取会话列表
func (s *service) GetConversations(userID int64, page, limit int) ([]*models.ConversationWithInfo, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	convs, total, err := s.dal.Conversation().GetByUser(userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("get conversations: %w", err)
	}

	result := make([]*models.ConversationWithInfo, 0, len(convs))
	for _, conv := range convs {
		info, err := s.GetConversation(conv.ID, userID)
		if err != nil {
			continue // 跳过有错误的会话
		}
		result = append(result, info)
	}

	return result, total, nil
}

// GetMessages 获取消息历史
func (s *service) GetMessages(conversationID int64, userID int64, beforeID int64, limit int) ([]*models.Message, bool, error) {
	// 验证会话存在且用户有权限
	conv, err := s.dal.Conversation().GetByID(conversationID)
	if err != nil {
		return nil, false, ErrConversationNotFound
	}

	if conv.UserAID != userID && conv.UserBID != userID {
		return nil, false, ErrAccessDenied
	}

	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	msgs, err := s.dal.Message().GetByConversation(conversationID, beforeID, limit)
	if err != nil {
		return nil, false, fmt.Errorf("get messages: %w", err)
	}

	hasMore := len(msgs) == limit
	return msgs, hasMore, nil
}

// MarkAsRead 标记会话消息为已读
func (s *service) MarkAsRead(conversationID int64, userID int64) error {
	// 验证会话存在且用户有权限
	conv, err := s.dal.Conversation().GetByID(conversationID)
	if err != nil {
		return ErrConversationNotFound
	}

	if conv.UserAID != userID && conv.UserBID != userID {
		return ErrAccessDenied
	}

	// 更新该会话中该用户接收的所有未读消息为已读
	return s.dal.Message().UpdateMessagesStatus(conversationID, userID, "read")
}

// UpdateMessageStatus 更新消息状态
func (s *service) UpdateMessageStatus(messageID int64, status string) error {
	if !isValidMessageStatus(status) {
		return fmt.Errorf("invalid message status: %s", status)
	}

	return s.dal.Message().UpdateStatus(messageID, status)
}

// GetOfflineMessages 获取离线消息
func (s *service) GetOfflineMessages(userID int64, lastMessageID int64, limit int) ([]*models.Message, error) {
	if limit < 1 {
		limit = 100
	}

	return s.dal.Message().GetOfflineMessages(userID, lastMessageID, limit)
}

// GetUnreadCount 获取未读消息数
func (s *service) GetUnreadCount(conversationID int64, userID int64) (int, error) {
	// 验证会话存在
	_, err := s.dal.Conversation().GetByID(conversationID)
	if err != nil {
		return 0, ErrConversationNotFound
	}

	return s.dal.Message().CountUnread(conversationID, userID)
}

// getOrCreateConversation 获取或创建会话
func (s *service) getOrCreateConversation(userA, userB int64) (*models.Conversation, error) {
	// 确保 userA < userB 保证唯一性
	small, large := userA, userB
	if small > large {
		small, large = large, small
	}

	// 尝试获取已有会话
	conv, err := s.dal.Conversation().GetByUsers(small, large)
	if err == nil && conv != nil {
		return conv, nil
	}

	// 创建新会话
	now := time.Now().Unix()
	conv = &models.Conversation{
		UserAID:   small,
		UserBID:   large,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if err := s.dal.Conversation().Create(conv); err != nil {
		return nil, fmt.Errorf("create conversation: %w", err)
	}

	return conv, nil
}

// toUserInfo 转换为UserInfo
func toUserInfo(user *models.User) *models.UserInfo {
	if user == nil {
		return nil
	}

	avatar := ""
	if user.AvatarID != nil {
		// TODO: 查询媒体表获取实际路径
		// 简化处理：暂时使用ID作为avatar值
		avatar = fmt.Sprintf("/api/media/%d", *user.AvatarID)
	}

	return &models.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   avatar,
		LastSeen: user.LastSeen,
	}
}

// isValidMessageType 验证消息类型
func isValidMessageType(t string) bool {
	switch t {
	case "text", "voice", "image":
		return true
	default:
		return false
	}
}

// isValidMessageStatus 验证消息状态
func isValidMessageStatus(s string) bool {
	switch s {
	case "sent", "delivered", "read":
		return true
	default:
		return false
	}
}
