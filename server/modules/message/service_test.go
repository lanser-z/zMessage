package message

import (
	"testing"
	"time"

	"zmessage/server/dal"
	"zmessage/server/models"
)

func setupTestService(t *testing.T) Service {
	t.Helper()

	mgr, err := dal.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("create dal manager: %v", err)
	}

	return NewService(mgr)
}

func setupTestUsers(t *testing.T, mgr dal.Manager) (*models.User, *models.User) {
	t.Helper()

	// 创建用户1
	user1 := &models.User{
		Username:     "alice",
		PasswordHash: "hash1",
		Nickname:     "Alice",
		CreatedAt:    time.Now().Unix(),
		LastSeen:     time.Now().Unix(),
	}
	if err := mgr.User().Create(user1); err != nil {
		t.Fatalf("create user1: %v", err)
	}

	// 创建用户2
	user2 := &models.User{
		Username:     "bob",
		PasswordHash: "hash2",
		Nickname:     "Bob",
		CreatedAt:    time.Now().Unix(),
		LastSeen:     time.Now().Unix(),
	}
	if err := mgr.User().Create(user2); err != nil {
		t.Fatalf("create user2: %v", err)
	}

	return user1, user2
}

func TestService_SendMessage(t *testing.T) {
	mgr, err := dal.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("create dal manager: %v", err)
	}

	svc := NewService(mgr)
	user1, user2 := setupTestUsers(t, mgr)

	// 测试发送文本消息
	req := &SendMessageRequest{
		From:    user1.ID,
		To:      user2.ID,
		Type:    "text",
		Content: "Hello Bob!",
	}

	msg, err := svc.SendMessage(req)
	if err != nil {
		t.Fatalf("send message failed: %v", err)
	}

	if msg.ID == 0 {
		t.Error("message ID not set")
	}

	if msg.SenderID != user1.ID {
		t.Errorf("expected sender_id %d, got %d", user1.ID, msg.SenderID)
	}

	if msg.ReceiverID != user2.ID {
		t.Errorf("expected receiver_id %d, got %d", user2.ID, msg.ReceiverID)
	}

	if msg.Type != "text" {
		t.Errorf("expected type 'text', got '%s'", msg.Type)
	}

	if msg.Content != "Hello Bob!" {
		t.Errorf("expected content 'Hello Bob!', got '%s'", msg.Content)
	}

	if msg.Status != "sent" {
		t.Errorf("expected status 'sent', got '%s'", msg.Status)
	}

	if msg.ConversationID == 0 {
		t.Error("conversation ID not set")
	}

	// 测试不能给自己发消息
	req.To = user1.ID
	_, err = svc.SendMessage(req)
	if err != ErrSendToSelf {
		t.Errorf("expected ErrSendToSelf, got: %v", err)
	}

	// 测试发送给不存在的用户
	req.To = 99999
	_, err = svc.SendMessage(req)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}

	// 测试无效的消息类型
	req.To = user2.ID
	req.Type = "invalid"
	_, err = svc.SendMessage(req)
	if err == nil {
		t.Error("expected error for invalid message type")
	}
}

func TestService_GetConversationWithUser(t *testing.T) {
	mgr, err := dal.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("create dal manager: %v", err)
	}

	svc := NewService(mgr)
	user1, user2 := setupTestUsers(t, mgr)

	// 发送消息创建会话
	req := &SendMessageRequest{
		From:    user1.ID,
		To:      user2.ID,
		Type:    "text",
		Content: "Hello!",
	}
	svc.SendMessage(req)

	// 获取会话
	conv, err := svc.GetConversationWithUser(user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("get conversation failed: %v", err)
	}

	if conv.ID == 0 {
		t.Error("conversation ID not set")
	}

	// 验证参与者
	if conv.Participant == nil {
		t.Fatal("participant is nil")
	}

	if conv.Participant.ID != user2.ID {
		t.Errorf("expected participant ID %d, got %d", user2.ID, conv.Participant.ID)
	}

	// 验证最后一条消息
	if conv.LastMessage == nil {
		t.Fatal("last message is nil")
	}

	if conv.LastMessage.Content != "Hello!" {
		t.Errorf("expected last message 'Hello!', got '%s'", conv.LastMessage.Content)
	}

	// 测试反向获取
	conv2, err := svc.GetConversationWithUser(user2.ID, user1.ID)
	if err != nil {
		t.Fatalf("get conversation reversed failed: %v", err)
	}

	if conv2.ID != conv.ID {
		t.Errorf("conversation ID mismatch: %d vs %d", conv2.ID, conv.ID)
	}
}

func TestService_GetConversations(t *testing.T) {
	mgr, err := dal.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("create dal manager: %v", err)
	}

	svc := NewService(mgr)
	user1, user2 := setupTestUsers(t, mgr)

	// 发送多条消息创建多个会话
	for i := 0; i < 3; i++ {
		req := &SendMessageRequest{
			From:    user1.ID,
			To:      user2.ID,
			Type:    "text",
			Content: "Message " + string(rune('1'+i)),
		}
		svc.SendMessage(req)
	}

	// 获取会话列表
	convs, total, err := svc.GetConversations(user1.ID, 1, 10)
	if err != nil {
		t.Fatalf("get conversations failed: %v", err)
	}

	if len(convs) == 0 {
		t.Error("expected at least 1 conversation")
	}

	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}

	// 验证会话信息
	conv := convs[0]
	if conv.Participant.ID != user2.ID {
		t.Errorf("expected participant ID %d, got %d", user2.ID, conv.Participant.ID)
	}
}

func TestService_GetMessages(t *testing.T) {
	mgr, err := dal.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("create dal manager: %v", err)
	}

	svc := NewService(mgr)
	user1, user2 := setupTestUsers(t, mgr)

	// 发送多条消息
	req := &SendMessageRequest{
		From:    user1.ID,
		To:      user2.ID,
		Type:    "text",
		Content: "First message",
	}
	svc.SendMessage(req)

	req.Content = "Second message"
	svc.SendMessage(req)

	// 获取会话ID
	conv, _ := svc.GetConversationWithUser(user1.ID, user2.ID)

	// 获取消息历史
	msgs, hasMore, err := svc.GetMessages(conv.ID, user1.ID, 0, 10)
	if err != nil {
		t.Fatalf("get messages failed: %v", err)
	}

	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}

	if hasMore {
		t.Error("expected no more messages")
	}

	// 验证消息顺序（最新的在前）
	if msgs[0].Content != "Second message" {
		t.Errorf("expected first message 'Second message', got '%s'", msgs[0].Content)
	}

	// 测试分页
	msgs, hasMore, err = svc.GetMessages(conv.ID, user1.ID, msgs[0].ID, 10)
	if err != nil {
		t.Fatalf("get messages with before_id failed: %v", err)
	}

	if len(msgs) != 1 {
		t.Errorf("expected 1 message with before_id, got %d", len(msgs))
	}

	if msgs[0].Content != "First message" {
		t.Errorf("expected 'First message', got '%s'", msgs[0].Content)
	}

	// 测试访问不存在的会话
	_, _, err = svc.GetMessages(99999, user1.ID, 0, 10)
	if err != ErrConversationNotFound {
		t.Errorf("expected ErrConversationNotFound, got: %v", err)
	}
}

func TestService_MarkAsRead(t *testing.T) {
	mgr, err := dal.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("create dal manager: %v", err)
	}

	svc := NewService(mgr)
	user1, user2 := setupTestUsers(t, mgr)

	// 发送消息
	req := &SendMessageRequest{
		From:    user1.ID,
		To:      user2.ID,
		Type:    "text",
		Content: "Hello!",
	}
	svc.SendMessage(req)

	// 获取会话
	conv, _ := svc.GetConversationWithUser(user1.ID, user2.ID)

	// 检查未读数
	unread, _ := svc.GetUnreadCount(conv.ID, user2.ID)
	if unread != 1 {
		t.Errorf("expected unread count 1, got %d", unread)
	}

	// 标记为已读
	if err := svc.MarkAsRead(conv.ID, user2.ID); err != nil {
		t.Fatalf("mark as read failed: %v", err)
	}

	// 检查未读数
	unread, _ = svc.GetUnreadCount(conv.ID, user2.ID)
	if unread != 0 {
		t.Errorf("expected unread count 0 after read, got %d", unread)
	}
}

func TestService_GetOfflineMessages(t *testing.T) {
	mgr, err := dal.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("create dal manager: %v", err)
	}

	svc := NewService(mgr)
	user1, user2 := setupTestUsers(t, mgr)

	// 发送多条消息
	for i := 0; i < 3; i++ {
		req := &SendMessageRequest{
			From:    user1.ID,
			To:      user2.ID,
			Type:    "text",
			Content: "Message " + string(rune('1'+i)),
		}
		svc.SendMessage(req)
	}

	// 获取离线消息（从0开始）
	msgs, err := svc.GetOfflineMessages(user2.ID, 0, 10)
	if err != nil {
		t.Fatalf("get offline messages failed: %v", err)
	}

	if len(msgs) != 3 {
		t.Errorf("expected 3 offline messages, got %d", len(msgs))
	}

	// 测试从某个ID之后获取
	msgs2, err := svc.GetOfflineMessages(user2.ID, msgs[0].ID, 10)
	if err != nil {
		t.Fatalf("get offline messages after id failed: %v", err)
	}

	if len(msgs2) != 2 {
		t.Errorf("expected 2 messages after first, got %d", len(msgs2))
	}
}

func TestService_UpdateMessageStatus(t *testing.T) {
	mgr, err := dal.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("create dal manager: %v", err)
	}

	svc := NewService(mgr)
	user1, user2 := setupTestUsers(t, mgr)

	// 发送消息
	req := &SendMessageRequest{
		From:    user1.ID,
		To:      user2.ID,
		Type:    "text",
		Content: "Hello!",
	}
	msg, _ := svc.SendMessage(req)

	// 更新为已投递
	if err := svc.UpdateMessageStatus(msg.ID, "delivered"); err != nil {
		t.Fatalf("update status to delivered failed: %v", err)
	}

	// 验证状态
	updatedMsg, err := mgr.Message().GetByID(msg.ID)
	if err != nil {
		t.Fatalf("get updated message failed: %v", err)
	}

	if updatedMsg.Status != "delivered" {
		t.Errorf("expected status 'delivered', got '%s'", updatedMsg.Status)
	}

	// 更新为已读
	if err := svc.UpdateMessageStatus(msg.ID, "read"); err != nil {
		t.Fatalf("update status to read failed: %v", err)
	}
}

func TestService_UnreadCount(t *testing.T) {
	mgr, err := dal.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("create dal manager: %v", err)
	}

	svc := NewService(mgr)
	user1, user2 := setupTestUsers(t, mgr)

	// 获取会话（发送消息后自动创建）
	req := &SendMessageRequest{
		From:    user1.ID,
		To:      user2.ID,
		Type:    "text",
		Content: "Hello!",
	}
	svc.SendMessage(req)

	conv, _ := svc.GetConversationWithUser(user1.ID, user2.ID)

	// 检查接收者的未读数
	count, err := svc.GetUnreadCount(conv.ID, user2.ID)
	if err != nil {
		t.Fatalf("get unread count failed: %v", err)
	}

	if count != 1 {
		t.Errorf("expected unread count 1, got %d", count)
	}

	// 发送者的未读数应该是0
	count, err = svc.GetUnreadCount(conv.ID, user1.ID)
	if err != nil {
		t.Fatalf("get sender unread count failed: %v", err)
	}

	if count != 0 {
		t.Errorf("expected sender unread count 0, got %d", count)
	}
}
