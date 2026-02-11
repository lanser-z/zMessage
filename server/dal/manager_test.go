package dal

import (
	"testing"
	"time"

	"zmessage/server/models"
)

// 测试用数据库管理器
func setupTestDB(t *testing.T) Manager {
	t.Helper()

	// 创建临时目录
	tmpDir := t.TempDir()

	// 创建管理器
	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("create manager: %v", err)
	}

	return mgr
}

func TestManager_NewManager(t *testing.T) {
	mgr := setupTestDB(t)
	defer mgr.Close()

	// 验证数据库连接
	if mgr.DB() == nil {
		t.Error("DB() returned nil")
	}

	// 验证所有DAL已初始化
	if mgr.User() == nil {
		t.Error("User() returned nil")
	}
	if mgr.Conversation() == nil {
		t.Error("Conversation() returned nil")
	}
	if mgr.Message() == nil {
		t.Error("Message() returned nil")
	}
	if mgr.Media() == nil {
		t.Error("Media() returned nil")
	}
	if mgr.SharedConversation() == nil {
		t.Error("SharedConversation() returned nil")
	}
}


func TestUserDAL(t *testing.T) {
	mgr := setupTestDB(t)
	defer mgr.Close()
	dal := mgr.User()

	// 测试创建用户
	user := &models.User{
		Username:     "testuser",
		PasswordHash: "hash123",
		Nickname:     "Test User",
		CreatedAt:    time.Now().Unix(),
		LastSeen:     time.Now().Unix(),
	}

	err := dal.Create(user)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("user ID not set")
	}

	// 测试根据ID获取用户
	fetched, err := dal.GetByID(user.ID)
	if err != nil {
		t.Fatalf("get user by id: %v", err)
	}

	if fetched.Username != user.Username {
		t.Errorf("username mismatch: got %s, want %s", fetched.Username, user.Username)
	}

	if fetched.PasswordHash != user.PasswordHash {
		t.Errorf("password hash mismatch")
	}

	// 测试根据用户名获取用户
	fetched, err = dal.GetByUsername(user.Username)
	if err != nil {
		t.Fatalf("get user by username: %v", err)
	}

	if fetched.ID != user.ID {
		t.Errorf("user ID mismatch")
	}

	// 测试获取不存在的用户
	_, err = dal.GetByID(99999)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got: %v", err)
	}

	// 测试列表
	users, err := dal.List("", 10)
	if err != nil {
		t.Fatalf("list users: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("expected 1 user, got %d", len(users))
	}

	// 测试搜索
	users, err = dal.List("test", 10)
	if err != nil {
		t.Fatalf("search users: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("expected 1 user in search, got %d", len(users))
	}

	// 测试更新
	user.Nickname = "Updated Name"
	user.LastSeen = time.Now().Unix()
	err = dal.Update(user)
	if err != nil {
		t.Fatalf("update user: %v", err)
	}

	fetched, _ = dal.GetByID(user.ID)
	if fetched.Nickname != "Updated Name" {
		t.Errorf("nickname not updated")
	}

	// 测试统计
	count, err := dal.Count()
	if err != nil {
		t.Fatalf("count users: %v", err)
	}

	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	// 测试删除
	err = dal.Delete(user.ID)
	if err != nil {
		t.Fatalf("delete user: %v", err)
	}

	_, err = dal.GetByID(user.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got: %v", err)
	}
}

func TestConversationDAL(t *testing.T) {
	mgr := setupTestDB(t)
	defer mgr.Close()
	dal := mgr.Conversation()
	userDAL := mgr.User()

	// 创建测试用户
	user1 := &models.User{
		Username:     "user1",
		PasswordHash: "hash1",
		Nickname:     "User One",
		CreatedAt:    time.Now().Unix(),
		LastSeen:     time.Now().Unix(),
	}
	userDAL.Create(user1)

	user2 := &models.User{
		Username:     "user2",
		PasswordHash: "hash2",
		Nickname:     "User Two",
		CreatedAt:    time.Now().Unix(),
		LastSeen:     time.Now().Unix(),
	}
	userDAL.Create(user2)

	// 测试创建会话
	conv := &models.Conversation{
		UserAID:   user1.ID,
		UserBID:   user2.ID,
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	err := dal.Create(conv)
	if err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	if conv.ID == 0 {
		t.Error("conversation ID not set")
	}

	// 验证user_a_id < user_b_id
	if conv.UserAID >= conv.UserBID {
		t.Error("user_a_id should be less than user_b_id")
	}

	// 测试根据ID获取会话
	fetched, err := dal.GetByID(conv.ID)
	if err != nil {
		t.Fatalf("get conversation by id: %v", err)
	}

	if fetched.UserAID != conv.UserAID {
		t.Errorf("user_a_id mismatch")
	}

	// 测试根据用户ID获取会话
	fetched, err = dal.GetByUsers(user1.ID, user2.ID)
	if err != nil {
		t.Fatalf("get conversation by users: %v", err)
	}

	if fetched.ID != conv.ID {
		t.Errorf("conversation ID mismatch")
	}

	// 测试获取用户的会话列表
	convs, total, err := dal.GetByUser(user1.ID, 1, 10)
	if err != nil {
		t.Fatalf("get conversations by user: %v", err)
	}

	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}

	if len(convs) != 1 {
		t.Errorf("expected 1 conversation, got %d", len(convs))
	}

	// 测试更新会话时间
	newTime := time.Now().Unix()
	err = dal.UpdateTime(conv.ID, newTime)
	if err != nil {
		t.Fatalf("update conversation time: %v", err)
	}

	fetched, _ = dal.GetByID(conv.ID)
	if fetched.UpdatedAt != newTime {
		t.Errorf("updated_at not changed")
	}

	// 测试删除
	err = dal.Delete(conv.ID)
	if err != nil {
		t.Fatalf("delete conversation: %v", err)
	}

	_, err = dal.GetByID(conv.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got: %v", err)
	}
}

func TestMessageDAL(t *testing.T) {
	mgr := setupTestDB(t)
	defer mgr.Close()
	dal := mgr.Message()
	convDAL := mgr.Conversation()
	userDAL := mgr.User()

	// 创建测试数据
	user1 := &models.User{Username: "user1", PasswordHash: "hash1", Nickname: "User1", CreatedAt: time.Now().Unix(), LastSeen: time.Now().Unix()}
	userDAL.Create(user1)

	user2 := &models.User{Username: "user2", PasswordHash: "hash2", Nickname: "User2", CreatedAt: time.Now().Unix(), LastSeen: time.Now().Unix()}
	userDAL.Create(user2)

	conv := &models.Conversation{UserAID: user1.ID, UserBID: user2.ID, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()}
	convDAL.Create(conv)

	// 测试创建消息
	msg := &models.Message{
		ConversationID: conv.ID,
		SenderID:       user1.ID,
		ReceiverID:     user2.ID,
		Type:           "text",
		Content:        "Hello, World!",
		Status:         "sent",
		CreatedAt:      time.Now().Unix(),
	}

	err := dal.Create(msg)
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	if msg.ID == 0 {
		t.Error("message ID not set")
	}

	// 测试根据ID获取消息
	fetched, err := dal.GetByID(msg.ID)
	if err != nil {
		t.Fatalf("get message by id: %v", err)
	}

	if fetched.Content != msg.Content {
		t.Errorf("content mismatch")
	}

	// 测试根据会话获取消息
	msgs, err := dal.GetByConversation(conv.ID, 0, 10)
	if err != nil {
		t.Fatalf("get messages by conversation: %v", err)
	}

	if len(msgs) != 1 {
		t.Errorf("expected 1 message, got %d", len(msgs))
	}

	// 测试离线消息
	offlineMsgs, err := dal.GetOfflineMessages(user2.ID, 0, 10)
	if err != nil {
		t.Fatalf("get offline messages: %v", err)
	}

	if len(offlineMsgs) != 1 {
		t.Errorf("expected 1 offline message, got %d", len(offlineMsgs))
	}

	// 测试更新消息状态
	err = dal.UpdateStatus(msg.ID, "delivered")
	if err != nil {
		t.Fatalf("update message status: %v", err)
	}

	fetched, _ = dal.GetByID(msg.ID)
	if fetched.Status != "delivered" {
		t.Errorf("status not updated")
	}

	// 测试统计未读消息
	count, err := dal.CountUnread(conv.ID, user2.ID)
	if err != nil {
		t.Fatalf("count unread: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 unread message, got %d", count)
	}

	// 测试批量更新消息状态
	err = dal.UpdateMessagesStatus(conv.ID, user2.ID, "read")
	if err != nil {
		t.Fatalf("update messages status: %v", err)
	}

	count, _ = dal.CountUnread(conv.ID, user2.ID)
	if count != 0 {
		t.Errorf("expected 0 unread messages after mark read, got %d", count)
	}

	// 测试删除消息
	err = dal.Delete(msg.ID)
	if err != nil {
		t.Fatalf("delete message: %v", err)
	}

	_, err = dal.GetByID(msg.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got: %v", err)
	}
}

func TestMediaDAL(t *testing.T) {
	mgr := setupTestDB(t)
	defer mgr.Close()
	dal := mgr.Media()
	userDAL := mgr.User()

	// 创建测试用户
	user := &models.User{Username: "user1", PasswordHash: "hash1", Nickname: "User1", CreatedAt: time.Now().Unix(), LastSeen: time.Now().Unix()}
	userDAL.Create(user)

	// 测试创建媒体
	media := &models.Media{
		OwnerID:       user.ID,
		Type:          "image",
		OriginalPath:  "/uploads/original/1.jpg",
		ThumbnailPath: "/uploads/thumb/1.jpg",
		Size:          1024000,
		MimeType:      "image/jpeg",
		Width:         ptr(1920),
		Height:        ptr(1080),
		CreatedAt:     time.Now().Unix(),
	}

	err := dal.Create(media)
	if err != nil {
		t.Fatalf("create media: %v", err)
	}

	if media.ID == 0 {
		t.Error("media ID not set")
	}

	// 测试根据ID获取媒体
	fetched, err := dal.GetByID(media.ID)
	if err != nil {
		t.Fatalf("get media by id: %v", err)
	}

	if fetched.Type != media.Type {
		t.Errorf("type mismatch")
	}

	// 测试根据所有者获取媒体
	medias, err := dal.GetByOwner(user.ID)
	if err != nil {
		t.Fatalf("get media by owner: %v", err)
	}

	if len(medias) != 1 {
		t.Errorf("expected 1 media, got %d", len(medias))
	}

	// 测试更新媒体
	media.Width = ptr(800)
	media.Height = ptr(600)
	err = dal.Update(media)
	if err != nil {
		t.Fatalf("update media: %v", err)
	}

	fetched, _ = dal.GetByID(media.ID)
	if *fetched.Width != 800 {
		t.Errorf("width not updated")
	}

	// 测试删除媒体
	err = dal.Delete(media.ID)
	if err != nil {
		t.Fatalf("delete media: %v", err)
	}

	_, err = dal.GetByID(media.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got: %v", err)
	}
}

func TestSharedConversationDAL(t *testing.T) {
	mgr := setupTestDB(t)
	defer mgr.Close()
	dal := mgr.SharedConversation()
	convDAL := mgr.Conversation()
	userDAL := mgr.User()

	// 创建测试数据
	user1 := &models.User{Username: "user1", PasswordHash: "hash1", Nickname: "User1", CreatedAt: time.Now().Unix(), LastSeen: time.Now().Unix()}
	userDAL.Create(user1)

	user2 := &models.User{Username: "user2", PasswordHash: "hash2", Nickname: "User2", CreatedAt: time.Now().Unix(), LastSeen: time.Now().Unix()}
	userDAL.Create(user2)

	conv := &models.Conversation{UserAID: user1.ID, UserBID: user2.ID, CreatedAt: time.Now().Unix(), UpdatedAt: time.Now().Unix()}
	convDAL.Create(conv)

	// 测试创建分享
	sc := &models.SharedConversation{
		ConversationID: conv.ID,
		ShareToken:     "token123",
		CreatedBy:      user1.ID,
		ExpireAt:       time.Now().Add(24 * time.Hour).Unix(),
		CreatedAt:      time.Now().Unix(),
	}

	err := dal.Create(sc)
	if err != nil {
		t.Fatalf("create shared conversation: %v", err)
	}

	if sc.ID == 0 {
		t.Error("shared conversation ID not set")
	}

	// 测试根据Token获取
	fetched, err := dal.GetByToken(sc.ShareToken)
	if err != nil {
		t.Fatalf("get shared conversation by token: %v", err)
	}

	if fetched.ID != sc.ID {
		t.Errorf("ID mismatch")
	}

	// 测试根据会话获取分享列表
	scs, err := dal.GetByConversation(conv.ID)
	if err != nil {
		t.Fatalf("get shared conversations: %v", err)
	}

	if len(scs) != 1 {
		t.Errorf("expected 1 shared conversation, got %d", len(scs))
	}

	// 测试删除
	err = dal.Delete(sc.ID)
	if err != nil {
		t.Fatalf("delete shared conversation: %v", err)
	}

	_, err = dal.GetByID(sc.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got: %v", err)
	}
}

// ptr 返回int指针的辅助函数
func ptr(i int) *int {
	return &i
}
