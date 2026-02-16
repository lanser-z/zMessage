package user

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"zmessage/server/models"
)

// Handler HTTP处理器
type Handler struct {
	service Service
}

// NewHandler 创建处理器
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Register 注册
func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
		return
	}

	resp, err := h.service.Register(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case ErrUserExists:
			c.JSON(http.StatusConflict, gin.H{"error": "USER_ALREADY_EXISTS", "message": err.Error()})
		case ErrInvalidUsername, ErrInvalidPassword:
			c.JSON(http.StatusBadRequest, gin.H{"error": "USER_INVALID_" + getErrorType(err), "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Login 登录
func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
		return
	}

	resp, err := h.service.Login(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case ErrUserNotFound, ErrInvalidPassword:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "UNAUTHORIZED", "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMe 获取当前用户信息
func (h *Handler) GetMe(c *gin.Context) {
	userID := getUserID(c)

	user, err := h.service.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "USER_NOT_FOUND", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

// ListUsers 获取用户列表
func (h *Handler) ListUsers(c *gin.Context) {
	search := c.Query("search")
	limit := 50

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	users, err := h.service.GetUsers(c.Request.Context(), search, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR", "message": err.Error()})
		return
	}

	resp := make([]*userResponse, len(users))
	for i, user := range users {
		resp[i] = toUserResponse(user)
	}

	c.JSON(http.StatusOK, gin.H{"users": resp})
}

// GetUser 获取指定用户信息
func (h *Handler) GetUser(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "USER_NOT_FOUND", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

// UpdateMe 更新当前用户信息
func (h *Handler) UpdateMe(c *gin.Context) {
	userID := getUserID(c)

	var req UpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request", "message": err.Error()})
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), userID, &req)
	if err != nil {
		switch err {
		case ErrUserNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "USER_NOT_FOUND", "message": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "INTERNAL_ERROR", "message": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, toUserResponse(user))
}

// userResponse 用户响应（不含敏感信息）
type userResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar,omitempty"`
	Online   bool   `json:"online"`
	LastSeen int64  `json:"last_seen"`
}

// toUserResponse 转换为响应格式
func toUserResponse(user *models.User) *userResponse {
	resp := &userResponse{
		ID:       user.ID,
		Username: user.Username,
		Nickname: user.Nickname,
		Online:   user.Online,
		LastSeen: user.LastSeen,
	}

	// TODO: 处理头像URL（需要媒体模块）
	// 返回相对路径，前端通过 <base href> 控制前缀
	if user.AvatarID != nil {
		resp.Avatar = "api/media/" + strconv.FormatInt(*user.AvatarID, 10)
	}

	return resp
}

// getUserID 从上下文获取用户ID
func getUserID(c *gin.Context) int64 {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(int64); ok {
			return id
		}
	}
	return 0
}

// getErrorType 获取错误类型
func getErrorType(err error) string {
	switch err {
	case ErrInvalidUsername:
		return "USERNAME"
	case ErrInvalidPassword:
		return "PASSWORD"
	default:
		return "UNKNOWN"
	}
}
