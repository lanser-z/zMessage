package share

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"time"
	"zmessage/server/dal"
	"zmessage/server/models"
)

// Service 分享服务接口
type Service interface {
	// CreateShare 创建分享
	CreateShare(req *models.ShareRequest, creatorID int64) (*models.SharedConversationDetail, error)

	// GetShare 获取分享详情
	GetShare(token string) (*models.SharedConversationDetail, error)

	// GetShareMessages 获取分享的消息列表
	GetShareMessages(token string, beforeID int64, limit int) ([]*models.SharedMessage, bool, error)

	// IncrementViewCount 增加访问次数
	IncrementViewCount(token string) error

	// ListShares 获取用户的分享列表
	ListShares(userID int64, page, limit int) ([]*models.ShareListItem, int, error)

	// DeleteShare 删除分享
	DeleteShare(shareID, userID int64) error

	// CleanupExpired 清理过期分享
	CleanupExpired() error
}


type service struct {
	dalMgr dal.Manager
}

// NewService 创建分享服务
func NewService(dalMgr dal.Manager) Service {
	return &service{dalMgr: dalMgr}
}

func (s *service) CreateShare(req *models.ShareRequest, creatorID int64) (*models.SharedConversationDetail, error) {
	// 验证会话是否存在以及用户是否有权限
	conv, err := s.dalMgr.Conversation().GetByID(req.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("conversation not found: %w", err)
	}

	// 检查用户是否是会话参与者
	if conv.UserAID != creatorID && conv.UserBID != creatorID {
		return nil, fmt.Errorf("access denied: not a participant")
	}

	// 确定消息范围
	var firstMessageID, lastMessageID int64
	var messageCount int

	if req.MessageRange == "recent" && req.RecentCount > 0 {
		// 获取最近 N 条消息
		messages, err := s.dalMgr.Message().GetByConversation(req.ConversationID, 0, req.RecentCount)
		if err != nil {
			return nil, fmt.Errorf("get messages: %w", err)
		}
		if len(messages) > 0 {
			// GetByConversation 返回降序，第一个是最新的，最后面是最老的
			lastMessageID = messages[0].ID           // 最新消息ID
			firstMessageID = messages[len(messages)-1].ID  // 最老消息ID
			messageCount = len(messages)
		}
	} else {
		// 获取全部消息数量
		allMessages, err := s.dalMgr.Message().GetByConversation(req.ConversationID, 0, 100000)
		if err != nil {
			return nil, fmt.Errorf("get all messages: %w", err)
		}
		if len(allMessages) > 0 {
			// GetByConversation 返回降序，第一个是最新的，最后面是最老的
			lastMessageID = allMessages[0].ID           // 最新消息ID
			firstMessageID = allMessages[len(allMessages)-1].ID  // 最老消息ID
			messageCount = len(allMessages)
		}
	}

	// 计算过期时间
	var expireAt int64
	if req.ExpireDays > 0 {
		expireAt = time.Now().Add(time.Duration(req.ExpireDays) * 24 * time.Hour).Unix()
	}

	// 生成分享 token
	token, err := generateShareToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	// 创建分享记录
	now := time.Now().Unix()
	sc := &models.SharedConversation{
		ConversationID: req.ConversationID,
		ShareToken:     token,
		CreatedBy:      creatorID,
		ExpireAt:       expireAt,
		CreatedAt:      now,
		FirstMessageID: firstMessageID,
		LastMessageID:  lastMessageID,
		MessageCount:   messageCount,
		ViewCount:      0,
	}

	if err := s.dalMgr.SharedConversation().Create(sc); err != nil {
		return nil, fmt.Errorf("create share: %w", err)
	}

	// 构建详情响应
	return s.buildShareDetail(sc)
}

func (s *service) GetShare(token string) (*models.SharedConversationDetail, error) {
	sc, err := s.dalMgr.SharedConversation().GetByToken(token)
	if err != nil {
		return nil, fmt.Errorf("share not found: %w", err)
	}

	return s.buildShareDetail(sc)
}

func (s *service) GetShareMessages(token string, beforeID int64, limit int) ([]*models.SharedMessage, bool, error) {
	sc, err := s.dalMgr.SharedConversation().GetByToken(token)
	if err != nil {
		return nil, false, fmt.Errorf("share not found: %w", err)
	}

	// 确定查询范围
	// 注意：GetByConversation 使用 id < ? 条件，所以传入 0 表示不限制
	actualBeforeID := beforeID

	// 获取消息
	messages, err := s.dalMgr.Message().GetByConversation(sc.ConversationID, actualBeforeID, limit)
	if err != nil {
		return nil, false, fmt.Errorf("get messages: %w", err)
	}

	// 过滤出分享范围内的消息
	var result []*models.SharedMessage
	var hasMore bool

	for _, msg := range messages {
		// 检查是否超出分享范围（只包含分享时存在的消息）
		if sc.LastMessageID > 0 && msg.ID > sc.LastMessageID {
			continue // 跳过分享后新发送的消息
		}
		if sc.FirstMessageID > 0 && msg.ID < sc.FirstMessageID {
			hasMore = false // 到达分享范围的起点
			break
		}

		// 获取发送者昵称
		sender, err := s.dalMgr.User().GetByID(msg.SenderID)
		if err != nil {
			continue
		}

		result = append(result, &models.SharedMessage{
			ID:             msg.ID,
			SenderID:       msg.SenderID,
			SenderNickname: sender.Nickname,
			Type:           msg.Type,
			Content:        msg.Content,
			CreatedAt:      msg.CreatedAt,
		})
	}

	// 检查是否还有更多
	if sc.FirstMessageID > 0 && len(result) > 0 {
		oldestID := result[len(result)-1].ID
		hasMore = oldestID > sc.FirstMessageID
	} else {
		hasMore = len(result) == limit
	}

	return result, hasMore, nil
}

func (s *service) IncrementViewCount(token string) error {
	sc, err := s.dalMgr.SharedConversation().GetByToken(token)
	if err != nil {
		return fmt.Errorf("share not found: %w", err)
	}

	return s.dalMgr.SharedConversation().UpdateViewCount(sc.ID, sc.ViewCount+1)
}

func (s *service) ListShares(userID int64, page, limit int) ([]*models.ShareListItem, int, error) {
	shares, total, err := s.dalMgr.SharedConversation().GetByCreator(userID, page, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("get shares: %w", err)
	}

	var result []*models.ShareListItem
	for _, sc := range shares {
		// 获取会话信息
		conv, err := s.dalMgr.Conversation().GetByID(sc.ConversationID)
		if err != nil {
			continue
		}

		// 确定对方用户
		var otherUserID int64
		if conv.UserAID == userID {
			otherUserID = conv.UserBID
		} else {
			otherUserID = conv.UserAID
		}

		// 获取对方用户信息
		otherUser, err := s.dalMgr.User().GetByID(otherUserID)
		if err != nil {
			continue
		}

		// 检查是否过期
		isExpired := sc.ExpireAt > 0 && time.Now().Unix() > sc.ExpireAt

		result = append(result, &models.ShareListItem{
			ID:             sc.ID,
			ShareToken:     sc.ShareToken,
			ShareURL:       fmt.Sprintf("/shared/%s", sc.ShareToken),
			ConversationID: sc.ConversationID,
			Participant: models.ParticipantInfo{
				ID:       otherUser.ID,
				Nickname: otherUser.Nickname,
			},
			MessageCount: sc.MessageCount,
			ExpireAt:     sc.ExpireAt,
			CreatedAt:    sc.CreatedAt,
			ViewCount:    sc.ViewCount,
			IsExpired:    isExpired,
		})
	}

	return result, total, nil
}

func (s *service) DeleteShare(shareID, userID int64) error {
	sc, err := s.dalMgr.SharedConversation().GetByID(shareID)
	if err != nil {
		return fmt.Errorf("share not found: %w", err)
	}

	// 检查权限
	if sc.CreatedBy != userID {
		return fmt.Errorf("access denied: not the creator")
	}

	return s.dalMgr.SharedConversation().Delete(shareID)
}

func (s *service) CleanupExpired() error {
	return s.dalMgr.SharedConversation().DeleteExpired()
}

// buildShareDetail 构建分享详情
func (s *service) buildShareDetail(sc *models.SharedConversation) (*models.SharedConversationDetail, error) {
	// 获取会话信息
	conv, err := s.dalMgr.Conversation().GetByID(sc.ConversationID)
	if err != nil {
		return nil, fmt.Errorf("get conversation: %w", err)
	}

	// 获取参与者信息
	userA, err := s.dalMgr.User().GetByID(conv.UserAID)
	if err != nil {
		return nil, fmt.Errorf("get user A: %w", err)
	}

	userB, err := s.dalMgr.User().GetByID(conv.UserBID)
	if err != nil {
		return nil, fmt.Errorf("get user B: %w", err)
	}

	participants := []models.ParticipantInfo{
		{ID: userA.ID, Nickname: userA.Nickname},
		{ID: userB.ID, Nickname: userB.Nickname},
	}

	// 获取分享者昵称
	creator, err := s.dalMgr.User().GetByID(sc.CreatedBy)
	if err != nil {
		return nil, fmt.Errorf("get creator: %w", err)
	}

	// 检查是否过期
	isExpired := sc.ExpireAt > 0 && time.Now().Unix() > sc.ExpireAt

	return &models.SharedConversationDetail{
		ID:               sc.ID,
		ShareToken:       sc.ShareToken,
		ConversationID:   sc.ConversationID,
		Participants:     participants,
		MessageCount:     sc.MessageCount,
		FirstMessageID:   sc.FirstMessageID,
		LastMessageID:    sc.LastMessageID,
		ExpireAt:         sc.ExpireAt,
		CreatedAt:        sc.CreatedAt,
		ViewCount:        sc.ViewCount,
		IsExpired:        isExpired,
		CreatedByNickname: creator.Nickname,
	}, nil
}

// generateShareToken 生成随机分享 token
func generateShareToken() (string, error) {
	b := make([]byte, 16) // 16 bytes = 128 bits
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// 使用 base32 编码，避免特殊字符
	return base32.StdEncoding.EncodeToString(b), nil
}
