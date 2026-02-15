package message

import (
	"fmt"
	"time"
	"zmessage/server/dal"
	"zmessage/server/models"
	"zmessage/server/sse"
)

// 广播消息给指定用户的所有 SSE 连接
func broadcastMessage(userID int64, msgType string, data interface{}) {
	sse.Broadcast(userID, msgType, data)
}

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
		return nil, fmt.Errorf("get sender: %w", err)
	}

	receiver, err := s.dal.User().GetByID(req.To)
	if err != nil || receiver == nil {
		return nil, fmt.Errorf("get receiver: %w", err)
	}

	// 获取或创建会话
	conv, err := s.getOrCreateConversation(req.From, req.To)
	if err != nil {
		return nil, err
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
		SyncedAt:      &now,
	}

	if err := s.dal.Message().Create(msg); err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	// 更新会话时间
	if err := s.dal.Conversation().UpdateTime(conv.ID, now); err != nil {
		return nil, fmt.Errorf("update conversation time: %w", err)
	}

	// 推送给接收者（通过 SSE）
	broadcastMessage(req.To, "chat", map[string]interface{}{
		"message_id":     msg.ID,
		"conversation_id": conv.ID,
		"sender_id":      req.From,
		"receiver_id":    req.To,
		"type":           req.Type,
		"content":        req.Content,
		"created_at":     msg.CreatedAt,
	})

	return msg, nil
}

// GetConversation 获取会话详情
func (s *service) GetConversation(id int64, userID int64) (*models.ConversationWithInfo, error) {
	conv, err := s.dal.Conversation().GetByID(id)
	if err != nil {
		return nil, ErrConversationNotFound
	}

	// 验证用户是否是会话参与者
	if conv.UserAID != userID && conv.UserBID != userID {
		return nil, ErrAccessDenied
	}

	// 获取对方用户信息
	var otherUserID int64
	if conv.UserAID == userID {
		otherUserID = conv.UserBID
	} else {
		otherUserID = conv.UserAID
	}

	otherUser, err := s.dal.User().GetByID(otherUserID)
	if err != nil {
		return nil, fmt.Errorf("get other user: %w", err)
	}

	// 获取未读消息数
	unreadCount, err := s.dal.Message().CountUnread(id, userID)
	if err != nil {
		return nil, fmt.Errorf("count unread: %w", err)
	}

	// 构建返回信息
	convInfo := &models.ConversationWithInfo{
		ID:           conv.ID,
		UserAID:      conv.UserAID,
		UserBID:      conv.UserBID,
		Participant:   toUserInfo(otherUser),
		UnreadCount:   unreadCount,
		CreatedAt:     conv.CreatedAt,
		UpdatedAt:     conv.UpdatedAt,
	}

	return convInfo, nil
}

// GetConversationWithUser 获取或创建与指定用户的会话
func (s *service) GetConversationWithUser(userID, otherUserID int64) (*models.ConversationWithInfo, error) {
	// 获取或创建会话
	conv, err := s.getOrCreateConversation(userID, otherUserID)
	if err != nil {
		return nil, err
	}

	// 获取对方用户信息
	var participantID int64
	if conv.UserAID == userID {
		participantID = conv.UserBID
	} else {
		participantID = conv.UserAID
	}

	participant, err := s.dal.User().GetByID(participantID)
	if err != nil {
		return nil, fmt.Errorf("get participant: %w", err)
	}

	// 获取未读消息数
	unreadCount, err := s.dal.Message().CountUnread(conv.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("count unread: %w", err)
	}

	// 获取最后一条消息
	messages, err := s.dal.Message().GetByConversation(conv.ID, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("get last message: %w", err)
	}

	var lastMessage *models.Message
	if len(messages) > 0 {
		lastMessage = messages[0]
	}

	// 构建返回信息
	convInfo := &models.ConversationWithInfo{
		ID:           conv.ID,
		UserAID:      conv.UserAID,
		UserBID:      conv.UserBID,
		Participant:   toUserInfo(participant),
		LastMessage:   lastMessage,
		UnreadCount:   unreadCount,
		CreatedAt:     conv.CreatedAt,
		UpdatedAt:     conv.UpdatedAt,
	}

	return convInfo, nil
}

// GetConversations 获取用户的所有会话
func (s *service) GetConversations(userID int64, page, limit int) ([]*models.ConversationWithInfo, int, error) {
	// 默认值
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	convs, total, err := s.dal.Conversation().GetByUser(userID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*models.ConversationWithInfo, 0, len(convs))
	for _, conv := range convs {
		convInfo, err := s.buildConversationInfo(conv, userID)
		if err != nil {
			return nil, 0, err
		}
		result = append(result, convInfo)
	}

	return result, total, nil
}

// GetMessages 获取消息历史
func (s *service) GetMessages(conversationID int64, userID int64, beforeID int64, limit int) ([]*models.Message, bool, error) {
	// 验证用户是否是会话参与者
	conv, err := s.dal.Conversation().GetByID(conversationID)
	if err != nil {
		return nil, false, ErrConversationNotFound
	}
	if conv.UserAID != userID && conv.UserBID != userID {
		return nil, false, ErrAccessDenied
	}

	// 默认值
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	messages, err := s.dal.Message().GetByConversation(conversationID, beforeID, limit)
	if err != nil {
		return nil, false, err
	}

	// 检查是否还有更多消息
	hasMore := len(messages) == limit
	return messages, hasMore, nil
}

// MarkAsRead 标记会话消息为已读
func (s *service) MarkAsRead(conversationID int64, userID int64) error {
	// 验证用户是否是会话参与者
	conv, err := s.dal.Conversation().GetByID(conversationID)
	if err != nil {
		return ErrConversationNotFound
	}
	if conv.UserAID != userID && conv.UserBID != userID {
		return ErrAccessDenied
	}

	// 更新消息状态
	return s.dal.Message().UpdateMessagesStatus(conversationID, userID, "read")
}

// UpdateMessageStatus 更新消息状态
func (s *service) UpdateMessageStatus(messageID int64, status string) error {
	return s.dal.Message().UpdateStatus(messageID, status)
}

// GetOfflineMessages 获取离线消息
func (s *service) GetOfflineMessages(userID int64, lastMessageID int64, limit int) ([]*models.Message, error) {
	// 默认值
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	return s.dal.Message().GetOfflineMessages(userID, lastMessageID, limit)
}

// GetUnreadCount 获取未读消息数
func (s *service) GetUnreadCount(conversationID int64, userID int64) (int, error) {
	// 验证用户是否是会话参与者
	conv, err := s.dal.Conversation().GetByID(conversationID)
	if err != nil {
		return 0, ErrConversationNotFound
	}
	if conv.UserAID != userID && conv.UserBID != userID {
		return 0, ErrAccessDenied
	}

	return s.dal.Message().CountUnread(conversationID, userID)
}

// getOrCreateConversation 获取或创建会话（内部方法）
func (s *service) getOrCreateConversation(from, to int64) (*models.Conversation, error) {
	// 确保 userA < userB（用于排序）
	userA, userB := from, to
	if userA > userB {
		userA, userB = userB, userA
	}

	// 检查是否已存在
	conv, err := s.dal.Conversation().GetByUsers(userA, userB)
	if err == nil {
		return conv, nil
	}

	// 创建新会话
	now := time.Now().Unix()
	conv = &models.Conversation{
		UserAID:   userA,
		UserBID:   userB,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.dal.Conversation().Create(conv); err != nil {
		return nil, fmt.Errorf("create conversation: %w", err)
	}

	return conv, nil
}

// buildConversationInfo 构建会话信息（内部方法）
func (s *service) buildConversationInfo(conv *models.Conversation, userID int64) (*models.ConversationWithInfo, error) {
	// 获取对方用户ID
	var otherUserID int64
	if conv.UserAID == userID {
		otherUserID = conv.UserBID
	} else {
		otherUserID = conv.UserAID
	}

	// 获取对方用户信息
	otherUser, err := s.dal.User().GetByID(otherUserID)
	if err != nil {
		return nil, fmt.Errorf("get participant: %w", err)
	}

	// 获取未读消息数
	unreadCount, err := s.dal.Message().CountUnread(conv.ID, userID)
	if err != nil {
		return nil, fmt.Errorf("count unread: %w", err)
	}

	// 获取最后一条消息
	messages, err := s.dal.Message().GetByConversation(conv.ID, 0, 1)
	if err != nil {
		return nil, fmt.Errorf("get last message: %w", err)
	}

	var lastMessage *models.Message
	if len(messages) > 0 {
		lastMessage = messages[0]
	}

	return &models.ConversationWithInfo{
		ID:           conv.ID,
		UserAID:      conv.UserAID,
		UserBID:      conv.UserBID,
		Participant:   toUserInfo(otherUser),
		LastMessage:   lastMessage,
		UnreadCount:   unreadCount,
		CreatedAt:     conv.CreatedAt,
		UpdatedAt:     conv.UpdatedAt,
	}, nil
}

// toUserInfo 转换用户信息为简要信息（内部方法）
func toUserInfo(user *models.User) *models.UserInfo {
	return &models.UserInfo{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Avatar:   avatarToString(user.AvatarID),
		LastSeen: user.LastSeen,
	}
}

// avatarToString 转换头像ID为字符串（内部方法）
func avatarToString(avatarID *int64) string {
	if avatarID == nil {
		return ""
	}
	return fmt.Sprintf("%d", *avatarID)
}

// isValidMessageType 验证消息类型（内部方法）
func isValidMessageType(msgType string) bool {
	switch msgType {
	case "text", "voice", "image":
		return true
	default:
		return false
	}
}
