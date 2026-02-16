package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"zmessage/server/models"
	"zmessage/server/modules/share"
	"zmessage/server/modules/user"
)

var shareSvc share.Service

// RegisterShareRoutes 注册分享路由
func RegisterShareRoutes(r *gin.Engine, svc share.Service, userSvc user.Service) {
	shareSvc = svc

	// 公开路由（无需认证）- 查看分享
	r.GET("/api/shared/:token", handleGetSharedContent)

	// 需要认证的路由
	api := r.Group("/api")
	api.Use(AuthMiddleware(userSvc))
	{
		// 创建分享
		api.POST("/conversations/:id/share", handleCreateShare)

		// 我的分享列表
		api.GET("/shares", handleGetMyShares)

		// 删除分享
		api.DELETE("/shares/:id", handleDeleteShare)
	}
}

// handleCreateShare 处理创建分享
func handleCreateShare(c *gin.Context) {
	auth := GetAuthContext(c)
	if auth == nil {
		return
	}

	// 获取会话ID
	convIDStr := c.Param("id")
	convID, err := strconv.ParseInt(convIDStr, 10, 64)
	if err != nil {
		BadRequest(c, "无效的会话ID")
		return
	}

	// 解析请求体
	var req models.ShareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		BadRequest(c, "请求参数错误")
		return
	}
	req.ConversationID = convID

	// 设置默认值
	if req.ExpireDays == 0 {
		req.ExpireDays = 7 // 默认7天
	}
	if req.MessageRange == "" {
		req.MessageRange = "recent"
	}
	if req.MessageRange == "recent" && req.RecentCount == 0 {
		req.RecentCount = 50 // 默认50条
	}

	// 创建分享
	detail, err := shareSvc.CreateShare(&req, auth.UserID)
	if err != nil {
		if err == share.ErrConversationNotFound || err == share.ErrAccessDenied {
			Forbidden(c, err.Error())
			return
		}
		InternalError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"id":           detail.ID,
		"share_token":  detail.ShareToken,
		"share_url":    "/shared/" + detail.ShareToken,
		"expire_at":    detail.ExpireAt,
		"message_count": detail.MessageCount,
	})
}

// handleGetSharedContent 处理获取分享内容（公开接口）
func handleGetSharedContent(c *gin.Context) {
	token := c.Param("token")

	// 增加访问次数
	_ = shareSvc.IncrementViewCount(token)

	// 获取分享详情
	detail, err := shareSvc.GetShare(token)
	if err != nil {
		if err == share.ErrShareNotFound || err == share.ErrShareExpired {
			NotFound(c, "分享不存在或已过期")
			return
		}
		InternalError(c, err)
		return
	}

	// 获取分页参数
	beforeIDStr := c.Query("before_id")
	beforeID, _ := strconv.ParseInt(beforeIDStr, 10, 64)
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit > 100 {
		limit = 100
	}
	if limit <= 0 {
		limit = 50
	}

	// 获取消息
	messages, hasMore, err := shareSvc.GetShareMessages(token, beforeID, limit)
	if err != nil {
		InternalError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"share": gin.H{
			"id":               detail.ID,
			"share_token":      detail.ShareToken,
			"conversation_id":  detail.ConversationID,
			"participants":     detail.Participants,
			"message_count":    detail.MessageCount,
			"first_message_id": detail.FirstMessageID,
			"last_message_id":  detail.LastMessageID,
			"expire_at":        detail.ExpireAt,
			"created_at":       detail.CreatedAt,
			"view_count":       detail.ViewCount,
			"is_expired":       detail.IsExpired,
			"created_by_nickname": detail.CreatedByNickname,
		},
		"messages": messages,
		"has_more": hasMore,
	})
}

// handleGetMyShares 处理获取我的分享列表
func handleGetMyShares(c *gin.Context) {
	auth := GetAuthContext(c)
	if auth == nil {
		return
	}

	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit > 50 {
		limit = 50
	}
	if limit <= 0 {
		limit = 20
	}

	// 获取分享列表
	shares, total, err := shareSvc.ListShares(auth.UserID, page, limit)
	if err != nil {
		InternalError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"shares": shares,
		"total":  total,
	})
}

// handleDeleteShare 处理删除分享
func handleDeleteShare(c *gin.Context) {
	auth := GetAuthContext(c)
	if auth == nil {
		return
	}

	// 获取分享ID
	shareIDStr := c.Param("id")
	shareID, err := strconv.ParseInt(shareIDStr, 10, 64)
	if err != nil {
		BadRequest(c, "无效的分享ID")
		return
	}

	// 删除分享
	err = shareSvc.DeleteShare(shareID, auth.UserID)
	if err != nil {
		if err == share.ErrShareNotFound {
			NotFound(c, "分享不存在")
			return
		}
		if err == share.ErrAccessDenied {
			Forbidden(c, "无权删除该分享")
			return
		}
		InternalError(c, err)
		return
	}

	c.JSON(200, gin.H{
		"success": true,
	})
}
