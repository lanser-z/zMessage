package user

import (
	"context"
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/go-playground/validator/v10"
	"zmessage/server/dal"
	"zmessage/server/models"
	"zmessage/server/pkg/jwt"
	"zmessage/server/pkg/password"
)

var (
	// ErrInvalidUsername 用户名格式无效
	ErrInvalidUsername = fmt.Errorf("USER_INVALID_USERNAME")

	// ErrInvalidPassword 密码格式无效
	ErrInvalidPassword = fmt.Errorf("USER_INVALID_PASSWORD")

	// ErrUserExists 用户名已存在
	ErrUserExists = fmt.Errorf("USER_ALREADY_EXISTS")

	// ErrUserNotFound 用户不存在
	ErrUserNotFound = fmt.Errorf("USER_NOT_FOUND")

	// ErrInvalidToken Token无效
	ErrInvalidToken = fmt.Errorf("USER_INVALID_TOKEN")

	// ErrTokenExpired Token过期
	ErrTokenExpired = fmt.Errorf("USER_TOKEN_EXPIRED")
)

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20"`
	Password string `json:"password" validate:"required,min=6,max=32"`
	Nickname string `json:"nickname" validate:"omitempty,max=50"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UpdateRequest 更新用户信息请求
type UpdateRequest struct {
	Nickname *string `json:"nickname" validate:"omitempty,max=50"`
	AvatarID *int64  `json:"avatar_id"`
}

// AuthResponse 认证响应
type AuthResponse struct {
	User  *models.User `json:"user"`
	Token string        `json:"token"`
}

// service 用户服务实现
type service struct {
	dal      dal.Manager
	jwt      *jwt.Manager
	password password.Hasher
	online   OnlineStatusManager
	validate  *validator.Validate
}

// NewService 创建用户服务
func NewService(dalMgr dal.Manager, jwtSecret string) Service {
	return &service{
		dal:      dalMgr,
		jwt:      jwt.NewManager(jwtSecret),
		password: password.NewHasher(),
		online:   NewOnlineStatusManager(),
		validate:  validator.New(),
	}
}

// Register 用户注册
func (s *service) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	// 验证请求
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 验证用户名格式（只允许字母数字下划线）
	if !isValidUsername(req.Username) {
		return nil, ErrInvalidUsername
	}

	// 检查用户名是否已存在
	existing, err := s.dal.User().GetByUsername(req.Username)
	if err == nil && existing != nil {
		return nil, ErrUserExists
	}

	// 哈希密码
	hash, err := s.password.Hash(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// 设置默认昵称
	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}

	// 创建用户
	now := time.Now().Unix()
	user := &models.User{
		Username:     req.Username,
		PasswordHash: hash,
		Nickname:     nickname,
		CreatedAt:    now,
		LastSeen:     now,
	}

	if err := s.dal.User().Create(user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// 生成Token
	token, err := s.jwt.GenerateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

// Login 用户登录
func (s *service) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	// 验证请求
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 查询用户
	user, err := s.dal.User().GetByUsername(req.Username)
	if err != nil {
		if err == dal.ErrNotFound {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	// 验证密码
	if !s.password.Verify(user.PasswordHash, req.Password) {
		return nil, ErrInvalidPassword
	}

	// 更新最后活跃时间
	now := time.Now().Unix()
	if err := s.dal.User().UpdateLastSeen(user.ID, now); err != nil {
		return nil, fmt.Errorf("update last seen: %w", err)
	}
	user.LastSeen = now

	// 生成Token
	token, err := s.jwt.GenerateToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	// 设置在线状态
	s.online.SetOnline(user.ID, true)

	return &AuthResponse{
		User:  user,
		Token: token,
	}, nil
}

// GetUserByID 根据ID获取用户
func (s *service) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	user, err := s.dal.User().GetByID(id)
	if err != nil {
		if err == dal.ErrNotFound {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	// 填充在线状态
	user.Online = s.online.IsOnline(user.ID)

	return user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *service) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	user, err := s.dal.User().GetByUsername(username)
	if err != nil {
		if err == dal.ErrNotFound {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	// 填充在线状态
	user.Online = s.online.IsOnline(user.ID)

	return user, nil
}

// GetUsers 获取用户列表
func (s *service) GetUsers(ctx context.Context, search string, limit int) ([]*models.User, error) {
	users, err := s.dal.User().List(search, limit)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	// 填充在线状态
	syncGroup := sync.WaitGroup{}
	for _, user := range users {
		syncGroup.Add(1)
		go func(u *models.User) {
			defer syncGroup.Done()
			u.Online = s.online.IsOnline(u.ID)
		}(user)
	}
	syncGroup.Wait()

	return users, nil
}

// UpdateUser 更新用户信息
func (s *service) UpdateUser(ctx context.Context, id int64, req *UpdateRequest) (*models.User, error) {
	// 验证请求
	if err := s.validate.Struct(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// 获取用户
	user, err := s.dal.User().GetByID(id)
	if err != nil {
		if err == dal.ErrNotFound {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}

	// 更新字段
	if req.Nickname != nil {
		user.Nickname = *req.Nickname
	}
	if req.AvatarID != nil {
		user.AvatarID = req.AvatarID
	}

	// 保存
	if err := s.dal.User().Update(user); err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return user, nil
}

// UpdateLastSeen 更新最后活跃时间
func (s *service) UpdateLastSeen(ctx context.Context, id int64) error {
	return s.dal.User().UpdateLastSeen(id, time.Now().Unix())
}

// ValidateToken 验证Token
func (s *service) ValidateToken(token string) (int64, error) {
	userID, err := s.jwt.ValidateToken(token)
	if err != nil {
		return 0, ErrInvalidToken
	}
	return userID, nil
}

// GenerateToken 生成Token
func (s *service) GenerateToken(userID int64) (string, error) {
	return s.jwt.GenerateToken(userID)
}

// OnlineStatus 获取在线状态管理器
func (s *service) OnlineStatus() OnlineStatusManager {
	return s.online
}

// isValidUsername 验证用户名格式
func isValidUsername(username string) bool {
	// 3-20字符，只允许字母数字下划线
	if len(username) < 3 || len(username) > 20 {
		return false
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, username)
	return matched
}
