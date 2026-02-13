package api

import (
	"github.com/gin-gonic/gin"
	"zmessage/server/modules/message"
)

// UserServiceAdapter 用户服务适配器
type UserServiceAdapter struct {
	svc message.Service
}

func (a *UserServiceAdapter) ValidateToken(token string) (int64, error) {
	return a.svc.(interface{ ValidateToken(string) (int64, error) }).ValidateToken(token)
}

// MediaServiceAdapter 媒体服务适配器
type MediaServiceAdapter struct {
	svc message.Service
}

func (a *MediaServiceAdapter) ValidateToken(token string) (int64, error) {
	return a.svc.(interface{ ValidateToken(string) (int64, error) }).ValidateToken(token)
}

// GetAuthContext 获取认证上下文（从gin.Context）
func GetAuthContext(c *gin.Context) *AuthContext {
	userID, exists := c.Get("auth")
	if !exists {
		return nil
	}
	return userID.(*AuthContext)
}

// UserResponseType 用户信息响应类型
type UserResponseType struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarID  *int64 `json:"avatar_id,omitempty"`
	CreatedAt int64  `json:"created_at"`
}

// UserResponse 用户信息响应
type UserResponse struct {
	User  UserResponseType `json:"user"`
	Token string    `json:"token"`
}
