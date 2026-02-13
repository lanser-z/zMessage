package api

import (
	"github.com/gin-gonic/gin"
	"zmessage/server/modules/media"
	"zmessage/server/modules/message"
	"zmessage/server/modules/user"
)

// Handler API处理器接口
type Handler interface {
	// RegisterRoutes 注册路由
	RegisterRoutes(router *gin.Engine)
}

// Service API服务接口
type Service interface {
	// UserService 获取用户服务
	UserService() user.Service
	// MessageService 获取消息服务
	MessageService() message.Service
	// MediaService 获取媒体服务
	MediaService() media.Service
	// WSManager 获取WebSocket管理器
	WSManager() WSManager
}

// WSManager WebSocket管理器接口
type WSManager interface {
	// HandleConnection 处理WebSocket连接
	HandleConnection(conn interface{})
	// BroadcastToUser 向用户广播消息
	BroadcastToUser(userID int64, message interface{}) error
	// IsOnline 检查用户是否在线
	IsOnline(userID int64) bool
}

// AuthContext 认证上下文
type AuthContext struct {
	UserID int64
}

// Response 统一响应格式
type Response struct {
	Code    int         `json:"code"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data,omitempty"`
}

// ResponseWithData 带数据的响应格式
type ResponseWithData struct {
	Code    int         `json:"code"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data"`
}

// ListResponse 列表响应格式
type ListResponse struct {
	Code    int         `json:"code"`
	Message  string      `json:"message,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Total    int         `json:"total,omitempty"`
}

// ErrorResponse 错误响应格式
type ErrorResponse struct {
	Error string `json:"error"`
}

// 中间件
// AuthMiddleware 认证中间件
func AuthMiddleware(svc user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, ErrorResponse{Error: "USER_UNAUTHORIZED"})
			c.Abort()
			return
		}

		// 移除 "Bearer " 前缀
		tokenStr := token
		if len(token) > 7 && token[:7] == "Bearer " {
			tokenStr = token[7:]
		}

		// 验证Token
		userID, err := svc.ValidateToken(tokenStr)
		if err != nil {
			c.JSON(401, ErrorResponse{Error: "USER_INVALID_TOKEN"})
			c.Abort()
			return
		}

		// 将用户ID存入上下文
		c.Set("auth", &AuthContext{UserID: userID})
		c.Next()
	}
}

// 辅助函数
// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(200, Response{Code: 0, Message: "success", Data: data})
}

// SuccessWithData 成功响应（带数据）
func SuccessWithData(c *gin.Context, data interface{}) {
	c.JSON(200, ResponseWithData{Code: 0, Message: "success", Data: data})
}

// SuccessList 列表响应
func SuccessList(c *gin.Context, data interface{}, total int) {
	c.JSON(200, ListResponse{Code: 0, Message: "success", Data: data, Total: total})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{Error: message})
}

// BadRequest 400错误
func BadRequest(c *gin.Context, message string) {
	Error(c, 400, message)
}

// Unauthorized 401错误
func Unauthorized(c *gin.Context, message string) {
	Error(c, 401, message)
}

// NotFound 404错误
func NotFound(c *gin.Context, message string) {
	Error(c, 404, message)
}

// InternalError 500错误
func InternalError(c *gin.Context, err error) {
	Error(c, 500, "INTERNAL_ERROR")
}

