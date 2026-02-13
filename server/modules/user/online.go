package user

import "sync"

// onlineStatusManager 在线状态管理器
type onlineStatusManager struct {
	users map[int64]bool
	mu    sync.RWMutex
}

// NewOnlineStatusManager 创建在线状态管理器
func NewOnlineStatusManager() OnlineStatusManager {
	return &onlineStatusManager{
		users: make(map[int64]bool),
	}
}

// SetOnline 设置用户在线状态
func (m *onlineStatusManager) SetOnline(userID int64, online bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.users[userID] = online
}

// IsOnline 检查用户是否在线
func (m *onlineStatusManager) IsOnline(userID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.users[userID]
}

// GetOnlineUsers 获取在线用户列表
func (m *onlineStatusManager) GetOnlineUsers() []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var users []int64
	for userID, online := range m.users {
		if online {
			users = append(users, userID)
		}
	}
	return users
}
