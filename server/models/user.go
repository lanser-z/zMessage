package models

// User 用户模型
type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"` // 不对外暴露
	Nickname     string `json:"nickname"`
	AvatarID     *int64 `json:"avatar_id,omitempty"`
	CreatedAt    int64  `json:"created_at"`
	LastSeen     int64  `json:"last_seen"`
	Online       bool   `json:"online,omitempty"` // 运行时状态，不存储
}
