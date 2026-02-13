package user

import (
	"context"
	"testing"

	"zmessage/server/dal"
)

func setupTestService(t *testing.T) Service {
	t.Helper()

	mgr, err := dal.NewManager(t.TempDir())
	if err != nil {
		t.Fatalf("create dal manager: %v", err)
	}

	return NewService(mgr, "test-secret")
}

func TestService_Register(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	// 测试成功注册
	req := &RegisterRequest{
		Username: "testuser",
		Password: "password123",
		Nickname: "Test User",
	}

	resp, err := svc.Register(ctx, req)
	if err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if resp.User.ID == 0 {
		t.Error("user ID not set")
	}

	if resp.Token == "" {
		t.Error("token not generated")
	}

	if resp.User.Username != req.Username {
		t.Errorf("username mismatch")
	}

	if resp.User.Nickname != req.Nickname {
		t.Errorf("nickname mismatch")
	}

	// 密码不应该明文返回
	if resp.User.PasswordHash == req.Password {
		t.Error("password hash is plain text")
	}

	// 测试重复用户名
	_, err = svc.Register(ctx, req)
	if err != ErrUserExists {
		t.Errorf("expected ErrUserExists, got: %v", err)
	}

	// 测试用户名格式验证
	invalidUsernames := []string{
		"ab",          // 太短
		"user@name",    // 无效字符
		"",            // 空
	}

	for _, username := range invalidUsernames {
		req.Username = username
		_, err := svc.Register(ctx, req)
		if err == nil {
			t.Errorf("expected error for username '%s'", username)
		}
	}
}

func TestService_Login(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	// 先注册用户
	registerReq := &RegisterRequest{
		Username: "testuser",
		Password: "password123",
	}
	svc.Register(ctx, registerReq)

	// 测试成功登录
	loginReq := &LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	resp, err := svc.Login(ctx, loginReq)
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if resp.User.ID == 0 {
		t.Error("user ID not set")
	}

	if resp.Token == "" {
		t.Error("token not generated")
	}

	// 测试错误密码
	loginReq.Password = "wrongpassword"
	_, err = svc.Login(ctx, loginReq)
	if err != ErrInvalidPassword {
		t.Errorf("expected ErrInvalidPassword, got: %v", err)
	}

	// 测试不存在的用户
	loginReq.Username = "nonexistent"
	loginReq.Password = "password123"
	_, err = svc.Login(ctx, loginReq)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestService_GetUserByID(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	// 先注册用户
	registerReq := &RegisterRequest{
		Username: "testuser",
		Password: "password123",
	}
	resp, _ := svc.Register(ctx, registerReq)

	// 测试获取存在的用户
	user, err := svc.GetUserByID(ctx, resp.User.ID)
	if err != nil {
		t.Fatalf("get user by id failed: %v", err)
	}

	if user.ID != resp.User.ID {
		t.Errorf("user ID mismatch")
	}

	// 测试获取不存在的用户
	_, err = svc.GetUserByID(ctx, 99999)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestService_GetUsers(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	// 注册多个用户
	usernames := []string{"alice", "bob", "charlie", "david", "eve"}
	for _, username := range usernames {
		req := &RegisterRequest{
			Username: username,
			Password: "password123",
		}
		svc.Register(ctx, req)
	}

	// 测试获取用户列表
	users, err := svc.GetUsers(ctx, "", 10)
	if err != nil {
		t.Fatalf("get users failed: %v", err)
	}

	if len(users) < 5 {
		t.Errorf("expected at least 5 users, got %d", len(users))
	}

	// 测试搜索
	users, err = svc.GetUsers(ctx, "ali", 10)
	if err != nil {
		t.Fatalf("search users failed: %v", err)
	}

	if len(users) != 1 {
		t.Errorf("expected 1 user in search, got %d", len(users))
	}
}

func TestService_UpdateUser(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	// 注册用户
	registerReq := &RegisterRequest{
		Username: "testuser",
		Password: "password123",
	}
	resp, _ := svc.Register(ctx, registerReq)

	// 测试更新用户
	updateReq := &UpdateRequest{
		Nickname: ptr("Updated Name"),
	}

	user, err := svc.UpdateUser(ctx, resp.User.ID, updateReq)
	if err != nil {
		t.Fatalf("update user failed: %v", err)
	}

	if user.Nickname != "Updated Name" {
		t.Errorf("nickname not updated")
	}

	// 测试更新不存在的用户
	_, err = svc.UpdateUser(ctx, 99999, updateReq)
	if err != ErrUserNotFound {
		t.Errorf("expected ErrUserNotFound, got: %v", err)
	}
}

func TestService_ValidateToken(t *testing.T) {
	svc := setupTestService(t)
	ctx := context.Background()

	// 注册用户
	registerReq := &RegisterRequest{
		Username: "testuser",
		Password: "password123",
	}
	resp, _ := svc.Register(ctx, registerReq)

	// 测试验证有效Token
	userID, err := svc.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("validate valid token failed: %v", err)
	}

	if userID != resp.User.ID {
		t.Errorf("user ID mismatch: got %d, want %d", userID, resp.User.ID)
	}

	// 测试验证无效Token
	_, err = svc.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestOnlineStatusManager(t *testing.T) {
	mgr := NewOnlineStatusManager()

	// 测试设置在线状态
	mgr.SetOnline(1, true)
	if !mgr.IsOnline(1) {
		t.Error("user should be online")
	}

	mgr.SetOnline(1, false)
	if mgr.IsOnline(1) {
		t.Error("user should be offline")
	}

	// 测试获取在线用户列表
	mgr.SetOnline(1, true)
	mgr.SetOnline(2, true)
	mgr.SetOnline(3, false)

	online := mgr.GetOnlineUsers()
	if len(online) != 2 {
		t.Errorf("expected 2 online users, got %d", len(online))
	}
}

func ptr(s string) *string {
	return &s
}
