package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"zmessage/server/modules/message"
)

// RegisterMessageRoutes 注册消息路由
func RegisterMessageRoutes(r *gin.Engine, svc message.Service, wsMgr WSManager) {
	msg := r.Group("/api/conversations/:id/messages")
	// 简单的token验证，实际应该使用user.Service
	msg.Use(func(c *gin.Context) {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(401, ErrorResponse{Error: "USER_UNAUTHORIZED"})
			c.Abort()
			return
		}
		// TODO: 实际项目中应该调用user.Service.ValidateToken
		c.Set("auth", &AuthContext{UserID: 1}) // 临时使用固定用户ID
		c.Next()
	})
	{
		msg.GET("", handleGetMessages(svc))
	}
}

// handleGetMessages 处理获取会话消息历史
func handleGetMessages(svc message.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := GetAuthContext(c)
		if auth == nil {
			return
		}

		// 获取会话ID
		idStr := c.Param("id")
		convID, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			BadRequest(c, "无效的会话ID")
			return
		}

		// 获取查询参数
		beforeStr := c.DefaultQuery("before_id", "0")
		beforeID, _ := strconv.ParseInt(beforeStr, 10, 64)
		limitStr := c.DefaultQuery("limit", "50")
		limit, _ := strconv.Atoi(limitStr)
		if limit > 100 {
			limit = 100
		}
		if limit <= 0 {
			limit = 50
		}

		// 调用消息服务
		messages, hasMore, err := svc.GetMessages(convID, auth.UserID, beforeID, limit)
		if err != nil {
			InternalError(c, err)
			return
		}

		// 转换为响应格式
		msgList := make([]MessageResponse, len(messages))
		for i, m := range messages {
			msgList[i] = MessageResponse{
				ID:            m.ID,
				ConversationID: m.ConversationID,
				SenderID:      m.SenderID,
				ReceiverID:    m.ReceiverID,
				Type:          m.Type,
				Content:       m.Content,
				Status:        m.Status,
				CreatedAt:     m.CreatedAt,
			}
		}

		c.JSON(200, MessagesResponse{
			Messages: msgList,
			HasMore:  hasMore,
		})
	}
}

// MessagesResponse 消息列表响应
type MessagesResponse struct {
	Messages []MessageResponse `json:"messages"`
	HasMore  bool            `json:"has_more"`
}

// MessageResponse 消息响应
type MessageResponse struct {
	ID            int64  `json:"id"`
	ConversationID int64  `json:"conversation_id"`
	SenderID      int64  `json:"sender_id"`
	ReceiverID    int64  `json:"receiver_id"`
	Type          string `json:"type"`
	Content       string `json:"content"`
	Status        string `json:"status"`
	CreatedAt     int64  `json:"created_at"`
}
