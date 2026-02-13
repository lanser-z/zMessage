package user

import (
	"context"
	"zmessage/server/models"
)

// Service 用户服务接口
type Service interface {
	// Register 用户注册
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)

	// Login 用户登录
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)

	// GetUserByID 根据ID获取用户
	GetUserByID(ctx context.Context, id int64) (*models.User, error)

	// GetUserByUsername 根据用户名获取用户
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)

	// GetUsers 获取用户列表
	GetUsers(ctx context.Context, search string, limit int) ([]*models.User, error)

	// UpdateUser 更新用户信息
	UpdateUser(ctx context.Context, id int64, req *UpdateRequest) (*models.User, error)

	// UpdateLastSeen 更新最后活跃时间
	UpdateLastSeen(ctx context.Context, id int64) error

	// ValidateToken 验证Token，返回用户ID
	ValidateToken(token string) (int64, error)

	// GenerateToken 生成Token
	GenerateToken(userID int64) (string, error)

	// OnlineStatus 获取在线状态管理器
	OnlineStatus() OnlineStatusManager
}

// OnlineStatusManager 在线状态管理器接口
type OnlineStatusManager interface {
	// SetOnline 设置用户在线状态
	SetOnline(userID int64, online bool)

	// IsOnline 检查用户是否在线
	IsOnline(userID int64) bool

	// GetOnlineUsers 获取在线用户列表
	GetOnlineUsers() []int64
}
