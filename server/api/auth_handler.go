package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"zmessage/server/modules/user"
)

// RegisterAuthRoutes 注册认证路由
func RegisterAuthRoutes(r *gin.Engine, svc user.Service) {
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", handleRegister(svc))
		auth.POST("/login", handleLogin(svc))
	}
}

// handleRegister 处理用户注册
func handleRegister(svc user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequest(c, err.Error())
			return
		}

		// 调用用户服务注册
		resp, err := svc.Register(c.Request.Context(), &user.RegisterRequest{
			Username: req.Username,
			Password: req.Password,
			Nickname: req.Nickname,
		})
		if err != nil {
			handleUserError(c, err)
			return
		}

		c.JSON(200, RegisterResponse{
			User: UserInfo{
				ID:       resp.User.ID,
				Username: resp.User.Username,
				Nickname: resp.User.Nickname,
				AvatarID: resp.User.AvatarID,
				CreatedAt: resp.User.CreatedAt,
			},
			Token: resp.Token,
		})
	}
}

// handleLogin 处理用户登录
func handleLogin(svc user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequest(c, err.Error())
			return
		}

		// 调试：打印用户名
		fmt.Printf("[LOGIN] Attempting login for user: %s\n", req.Username)

		// 调用用户服务登录
		resp, err := svc.Login(c.Request.Context(), &user.LoginRequest{
			Username: req.Username,
			Password: req.Password,
		})
		if err != nil {
			fmt.Printf("[LOGIN] Login failed: %v\n", err)
			handleUserError(c, err)
			return
		}

		clientIP := c.ClientIP()
		fmt.Printf("[LOGIN] Login successful for user: %s (ID: %d) from %s\n", resp.User.Username, resp.User.ID, clientIP)
		c.JSON(200, LoginResponse{
			User: UserInfo{
				ID:       resp.User.ID,
				Username: resp.User.Username,
				Nickname: resp.User.Nickname,
				AvatarID: resp.User.AvatarID,
				CreatedAt: resp.User.CreatedAt,
			},
			Token: resp.Token,
		})
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Nickname string `json:"nickname" binding:"max=20"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required,min=3,max=20"`
	Password string `json:"password" binding:"required,min=6,max=32"`
}

// RegisterResponse 注册响应
type RegisterResponse struct {
	User  UserInfo `json:"user"`
	Token string    `json:"token"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	User  UserInfo `json:"user"`
	Token string    `json:"token"`
}

// UserInfo 用户信息响应
type UserInfo struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	AvatarID  *int64 `json:"avatar_id,omitempty"`
	CreatedAt int64  `json:"created_at"`
}
