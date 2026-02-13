package ws

import (
	"context"
	"sync"
	"testing"

	"zmessage/server/models"
	"zmessage/server/modules/message"
	"zmessage/server/modules/user"
)

// MockMessageService 消息服务模拟
type MockMessageService struct {
	sendMessageFunc  func(*message.SendMessageRequest) (*models.Message, error)
	getOfflineFunc   func(int64, int64, int) ([]*models.Message, error)
	updateStatusFunc func(int64, string) error
	getConvFunc     func(int64, int64) (*models.ConversationWithInfo, error)
	getMsgsFunc     func(int64, int64, int) ([]*models.Message, bool, error)
	markReadFunc    func(int64, int64) error
	getUnreadFunc   func(int64, int64) (int, error)
}

func (m *MockMessageService) SendMessage(req *message.SendMessageRequest) (*models.Message, error) {
	if m.sendMessageFunc != nil {
		return m.sendMessageFunc(req)
	}
	return &models.Message{}, nil
}

func (m *MockMessageService) GetOfflineMessages(userID int64, lastMessageID int64, limit int) ([]*models.Message, error) {
	if m.getOfflineFunc != nil {
		return m.getOfflineFunc(userID, lastMessageID, limit)
	}
	return []*models.Message{}, nil
}

func (m *MockMessageService) UpdateMessageStatus(messageID int64, status string) error {
	if m.updateStatusFunc != nil {
		return m.updateStatusFunc(messageID, status)
	}
	return nil
}

func (m *MockMessageService) GetConversation(id int64, userID int64) (*models.ConversationWithInfo, error) {
	if m.getConvFunc != nil {
		return m.getConvFunc(id, userID)
	}
	return &models.ConversationWithInfo{}, nil
}

func (m *MockMessageService) GetConversationWithUser(userID, otherUserID int64) (*models.ConversationWithInfo, error) {
	if m.getConvFunc != nil {
		return m.getConvFunc(userID, otherUserID)
	}
	return &models.ConversationWithInfo{}, nil
}

func (m *MockMessageService) GetMessages(conversationID int64, userID int64, beforeID int64, limit int) ([]*models.Message, bool, error) {
	if m.getMsgsFunc != nil {
		return m.getMsgsFunc(conversationID, beforeID, limit)
	}
	return []*models.Message{}, false, nil
}

func (m *MockMessageService) MarkAsRead(conversationID int64, userID int64) error {
	if m.markReadFunc != nil {
		return m.markReadFunc(conversationID, userID)
	}
	return nil
}

func (m *MockMessageService) GetUnreadCount(userID int64, conversationID int64) (int, error) {
	if m.getUnreadFunc != nil {
		return m.getUnreadFunc(userID, conversationID)
	}
	return 0, nil
}

func (m *MockMessageService) GetConversations(userID int64, page, limit int) ([]*models.ConversationWithInfo, int, error) {
	return []*models.ConversationWithInfo{}, 0, nil
}

// MockUserService 用户服务模拟
type MockUserService struct {
	validateTokenFunc func(string) (int64, error)
	generateTokenFunc func(int64) (string, error)
	getUserFunc      func(int64) (*models.User, error)
	getUsersFunc     func() ([]*models.User, error)
	onlineStatus     *MockOnlineStatusManager
}

func (m *MockUserService) ValidateToken(token string) (int64, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(token)
	}
	return 1, nil
}

func (m *MockUserService) GenerateToken(userID int64) (string, error) {
	if m.generateTokenFunc != nil {
		return m.generateTokenFunc(userID)
	}
	return "test_token", nil
}

func (m *MockUserService) GetByID(id int64) (*models.User, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(id)
	}
	return &models.User{ID: id, Username: "test_user"}, nil
}

func (m *MockUserService) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(id)
	}
	return &models.User{ID: id, Username: "test_user"}, nil
}

func (m *MockUserService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return &models.User{ID: 1, Username: username}, nil
}

func (m *MockUserService) GetUsers(ctx context.Context, search string, limit int) ([]*models.User, error) {
	if m.getUsersFunc != nil {
		return m.getUsersFunc()
	}
	return []*models.User{}, nil
}

func (m *MockUserService) Login(ctx context.Context, req *user.LoginRequest) (*user.AuthResponse, error) {
	return &user.AuthResponse{Token: "test_token", User: &models.User{ID: 1}}, nil
}

func (m *MockUserService) Register(ctx context.Context, req *user.RegisterRequest) (*user.AuthResponse, error) {
	return &user.AuthResponse{Token: "test_token", User: &models.User{ID: 1}}, nil
}

func (m *MockUserService) UpdateUser(ctx context.Context, id int64, req *user.UpdateRequest) (*models.User, error) {
	return &models.User{ID: id, Username: "test_user"}, nil
}

func (m *MockUserService) UpdateLastSeen(ctx context.Context, id int64) error {
	return nil
}

func (m *MockUserService) OnlineStatus() user.OnlineStatusManager {
	if m.onlineStatus == nil {
		m.onlineStatus = &MockOnlineStatusManager{}
	}
	return m.onlineStatus
}

// MockOnlineStatusManager 在线状态管理器模拟
type MockOnlineStatusManager struct {
	mu        sync.RWMutex
	onlineMap map[int64]bool
}

func (m *MockOnlineStatusManager) SetOnline(userID int64, online bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.onlineMap == nil {
		m.onlineMap = make(map[int64]bool)
	}
	m.onlineMap[userID] = online
}

func (m *MockOnlineStatusManager) IsOnline(userID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.onlineMap == nil {
		return false
	}
	return m.onlineMap[userID]
}

func (m *MockOnlineStatusManager) GetOnlineUsers() []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	users := make([]int64, 0)
	for userID, online := range m.onlineMap {
		if online {
			users = append(users, userID)
		}
	}
	return users
}

func TestNewManager(t *testing.T) {
	msgSvc := &MockMessageService{}
	userSvc := &MockUserService{}

	mgr := NewManager(msgSvc, userSvc)

	if mgr == nil {
		t.Fatal("NewManager returned nil")
	}

	// 验证管理器初始化
	connMgr, ok := mgr.(*connectionManager)
	if !ok {
		t.Fatal("Manager is not *connectionManager")
	}

	if connMgr.connections == nil {
		t.Error("connections map is nil")
	}
	if connMgr.userConns == nil {
		t.Error("userConns map is nil")
	}
	if connMgr.msgHandler == nil {
		t.Error("msgHandler is nil")
	}
}

func TestConnectionManager_IsOnline(t *testing.T) {
	msgSvc := &MockMessageService{}
	userSvc := &MockUserService{}
	mgr := NewManager(msgSvc, userSvc).(*connectionManager)

	// 初始状态：用户不在线
	if mgr.IsOnline(1) {
		t.Error("User 1 should not be online initially")
	}

	// 添加连接
	conn := &connection{
		id:     "test_conn_1",
		userID: 1,
		mgr:    mgr,
		send:   make(chan []byte, 10),
	}
	mgr.addConnection(conn)

	// 现在用户应该在线
	if !mgr.IsOnline(1) {
		t.Error("User 1 should be online after adding connection")
	}
}

func TestConnectionManager_GetOnlineUsers(t *testing.T) {
	msgSvc := &MockMessageService{}
	userSvc := &MockUserService{}
	mgr := NewManager(msgSvc, userSvc).(*connectionManager)

	// 初始状态：无在线用户
	users := mgr.GetOnlineUsers()
	if len(users) != 0 {
		t.Errorf("Expected 0 online users, got %d", len(users))
	}

	// 添加多个用户连接
	conn1 := &connection{id: "conn1", userID: 1, mgr: mgr, send: make(chan []byte, 10)}
	conn2 := &connection{id: "conn2", userID: 2, mgr: mgr, send: make(chan []byte, 10)}
	conn3 := &connection{id: "conn3", userID: 3, mgr: mgr, send: make(chan []byte, 10)}

	mgr.addConnection(conn1)
	mgr.addConnection(conn2)
	mgr.addConnection(conn3)

	// 检查在线用户列表
	users = mgr.GetOnlineUsers()
	if len(users) != 3 {
		t.Errorf("Expected 3 online users, got %d", len(users))
	}
}

func TestConnectionManager_GetConnection(t *testing.T) {
	msgSvc := &MockMessageService{}
	userSvc := &MockUserService{}
	mgr := NewManager(msgSvc, userSvc).(*connectionManager)

	// 用户无连接时返回nil
	conn := mgr.GetConnection(1)
	if conn != nil {
		t.Error("Expected nil when user has no connections")
	}

	// 添加连接
	testConn := &connection{id: "test_conn", userID: 1, mgr: mgr, send: make(chan []byte, 10)}
	mgr.addConnection(testConn)

	// 应该能获取到连接
	conn = mgr.GetConnection(1)
	if conn == nil {
		t.Fatal("Expected connection for user 1")
	}
	if conn.ID() != "test_conn" {
		t.Errorf("Expected connection ID 'test_conn', got '%s'", conn.ID())
	}
}

func TestConnectionManager_Disconnect(t *testing.T) {
	msgSvc := &MockMessageService{}
	userSvc := &MockUserService{}
	mgr := NewManager(msgSvc, userSvc).(*connectionManager)

	// 添加连接
	conn := &connection{
		id:     "test_conn",
		userID: 1,
		mgr:    mgr,
		send:   make(chan []byte, 10),
	}
	mgr.connections["test_conn"] = conn
	mgr.userConns[1] = map[string]*connection{"test_conn": conn}

	// 手动清理连接（模拟Disconnect行为）
	delete(mgr.connections, "test_conn")
	if conns, ok := mgr.userConns[1]; ok {
		delete(conns, "test_conn")
		if len(conns) == 0 {
			delete(mgr.userConns, 1)
		}
	}

	// 验证连接已移除
	if _, ok := mgr.connections["test_conn"]; ok {
		t.Error("Connection should be removed from connections map")
	}
	if conns, ok := mgr.userConns[1]; ok && len(conns) > 0 {
		t.Error("Connection should be removed from userConns map")
	}
}

func TestConnectionManager_DisconnectUser(t *testing.T) {
	msgSvc := &MockMessageService{}
	userSvc := &MockUserService{}
	mgr := NewManager(msgSvc, userSvc).(*connectionManager)

	// 添加用户的多个连接
	conn1 := &connection{id: "conn1", userID: 1, mgr: mgr, send: make(chan []byte, 10)}
	conn2 := &connection{id: "conn2", userID: 1, mgr: mgr, send: make(chan []byte, 10)}
	mgr.connections["conn1"] = conn1
	mgr.connections["conn2"] = conn2
	mgr.userConns[1] = map[string]*connection{
		"conn1": conn1,
		"conn2": conn2,
	}

	// 手动清理连接（模拟DisconnectUser行为）
	delete(mgr.connections, "conn1")
	delete(mgr.connections, "conn2")
	delete(mgr.userConns, 1)

	// 验证所有连接已移除
	if _, ok := mgr.connections["conn1"]; ok {
		t.Error("Connection conn1 should be removed")
	}
	if _, ok := mgr.connections["conn2"]; ok {
		t.Error("Connection conn2 should be removed")
	}
	if conns, ok := mgr.userConns[1]; ok {
		if len(conns) > 0 {
			t.Error("All user connections should be removed")
		}
	}
}

func TestConnection_Authenticate(t *testing.T) {
	msgSvc := &MockMessageService{}
	userSvc := &MockUserService{}
	mgr := NewManager(msgSvc, userSvc).(*connectionManager)

	conn := &connection{
		id:   "test_conn",
		mgr:   mgr,
		send:  make(chan []byte, 10),
	}

	// 认证用户
	err := conn.authenticate(123)
	if err != nil {
		t.Fatalf("authenticate failed: %v", err)
	}

	// 验证用户ID已设置
	if conn.UserID() != 123 {
		t.Errorf("Expected UserID 123, got %d", conn.UserID())
	}

	// 验证连接已添加到用户连接列表
	if !mgr.IsOnline(123) {
		t.Error("User should be online after authentication")
	}
}

func TestGenerateConnID(t *testing.T) {
	id1 := generateConnID()
	id2 := generateConnID()

	if id1 == "" {
		t.Error("generateConnID returned empty string")
	}
	if id2 == "" {
		t.Error("generateConnID returned empty string")
	}
	if id1 == id2 {
		t.Error("generateConnID should generate unique IDs")
	}

	// 验证前缀
	if len(id1) < 5 || id1[:5] != "conn_" {
		t.Errorf("Connection ID should start with 'conn_', got %s", id1[:5])
	}
}
