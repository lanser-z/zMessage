package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"zmessage/server/modules/user"
)

// RegisterUsersRoutes 注册用户路由
func RegisterUsersRoutes(r *gin.Engine, svc user.Service) {
	users := r.Group("/api/users")
	users.Use(AuthMiddleware(svc))
	{
		users.GET("/me", handleGetMe(svc))
		users.GET("", handleGetUsers(svc))
		users.GET("/:id", handleGetUserByID(svc))
		users.PUT("/me", handleUpdateMe(svc))
	}
}

// handleGetMe 处理获取当前用户信息
func handleGetMe(svc user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := GetAuthContext(c)
		if auth == nil {
			return
		}

		// 获取用户信息
		user, err := svc.GetUserByID(c.Request.Context(), auth.UserID)
		if err != nil {
			InternalError(c, err)
			return
		}

		c.JSON(200, UserResponseType{
			ID:        user.ID,
			Username:  user.Username,
			Nickname:  user.Nickname,
			AvatarID:  user.AvatarID,
			CreatedAt: user.CreatedAt,
		})
	}
}

// handleGetUsers 处理获取用户列表
func handleGetUsers(svc user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := GetAuthContext(c)
		if auth == nil {
			return
		}

		// 获取查询参数
		search := c.DefaultQuery("search", "")
		limitStr := c.DefaultQuery("limit", "50")
		limit, _ := strconv.Atoi(limitStr)
		if limit > 50 {
			limit = 50
		}
		if limit <= 0 {
			limit = 50
		}

		// 调用用户服务
		users, err := svc.GetUsers(c.Request.Context(), search, limit)
		if err != nil {
			InternalError(c, err)
			return
		}

		// 转换为响应格式
		userList := make([]UserResponseType, len(users))
		for i, u := range users {
			userList[i] = UserResponseType{
				ID:        u.ID,
				Username:  u.Username,
				Nickname:  u.Nickname,
				AvatarID:  u.AvatarID,
				CreatedAt: u.CreatedAt,
			}
		}

		SuccessList(c, userList, len(userList))
	}
}

// handleGetUserByID 处理获取指定用户信息
func handleGetUserByID(svc user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := GetAuthContext(c)
		if auth == nil {
			return
		}

		// 获取用户ID
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			BadRequest(c, "无效的用户ID")
			return
		}

		// 获取用户信息
		user, err := svc.GetUserByID(c.Request.Context(), id)
		if err != nil {
			handleUserError(c, err)
			return
		}

		c.JSON(200, UserResponseType{
			ID:        user.ID,
			Username:  user.Username,
			Nickname:  user.Nickname,
			AvatarID:  user.AvatarID,
			CreatedAt: user.CreatedAt,
		})
	}
}

// handleUpdateMe 处理更新当前用户信息
func handleUpdateMe(svc user.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := GetAuthContext(c)
		if auth == nil {
			return
		}

		var req UpdateMeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			BadRequest(c, err.Error())
			return
		}

		// 调用用户服务更新
		var nickname *string
		if req.Nickname != "" {
			nickname = &req.Nickname
		}
		user, err := svc.UpdateUser(c.Request.Context(), auth.UserID, &user.UpdateRequest{
			Nickname: nickname,
			AvatarID: req.AvatarID,
		})
		if err != nil {
			InternalError(c, err)
			return
		}

		c.JSON(200, UserResponseType{
			ID:        user.ID,
			Username:  user.Username,
			Nickname:  user.Nickname,
			AvatarID:  user.AvatarID,
			CreatedAt: user.CreatedAt,
		})
	}
}

// UpdateMeRequest 更新用户信息请求
type UpdateMeRequest struct {
	Nickname  string `json:"nickname" binding:"omitempty,max=20"`
	AvatarID  *int64 `json:"avatar_id" binding:"omitempty"`
}

// handleUserError 处理用户服务错误
func handleUserError(c *gin.Context, err error) {
	switch err.Error() {
	case "USER_INVALID_USERNAME":
		BadRequest(c, "用户名格式无效")
	case "USER_INVALID_PASSWORD":
		BadRequest(c, "密码错误")
	case "USER_ALREADY_EXISTS":
		c.JSON(409, ErrorResponse{Error: "USER_ALREADY_EXISTS"})
	case "USER_NOT_FOUND":
		NotFound(c, "用户不存在")
	default:
		InternalError(c, err)
	}
}
