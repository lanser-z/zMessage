package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"zmessage/server/modules/message"
	"zmessage/server/modules/user"
)

// RegisterMessageRoutes 注册消息路由
func RegisterMessageRoutes(r *gin.Engine, msgSvc message.Service, userSvc user.Service, wsMgr WSManager) {
	msg := r.Group("/api/conversations/:id/messages")
	msg.Use(AuthMiddleware(userSvc))
	{
		msg.GET("", handleGetMessages(msgSvc))
		msg.POST("", handleSendMessage(msgSvc))
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

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// handleSendMessage 处理发送消息
func handleSendMessage(svc message.Service) gin.HandlerFunc {
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

		// 解析请求体
		var req SendMessageRequest
		if err := c.BindJSON(&req); err != nil {
			BadRequest(c, "无效的请求格式")
			return
		}

		// 验证消息类型和内容
		if req.Type == "" || req.Content == "" {
			BadRequest(c, "消息类型和内容不能为空")
			return
		}

		// 获取会话信息以确定接收者
		conv, err := svc.GetConversation(convID, auth.UserID)
		if err != nil {
			InternalError(c, err)
			return
		}

		// 确定接收者ID
		var receiverID int64
		if conv.UserAID == auth.UserID {
			receiverID = conv.UserBID
		} else {
			receiverID = conv.UserAID
		}

		// 调用消息服务发送
		msg, err := svc.SendMessage(&message.SendMessageRequest{
			From:    auth.UserID,
			To:      receiverID,
			Type:    req.Type,
			Content: req.Content,
		})
		if err != nil {
			InternalError(c, err)
			return
		}

		c.JSON(200, MessageResponse{
			ID:            msg.ID,
			ConversationID: msg.ConversationID,
			SenderID:      msg.SenderID,
			ReceiverID:    msg.ReceiverID,
			Type:          msg.Type,
			Content:       msg.Content,
			Status:        msg.Status,
			CreatedAt:     msg.CreatedAt,
		})
	}
}
