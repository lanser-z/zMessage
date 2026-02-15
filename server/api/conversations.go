package api

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"zmessage/server/modules/message"
	"zmessage/server/modules/user"
)

// RegisterConversationRoutes 注册会话路由
func RegisterConversationRoutes(r *gin.Engine, msgSvc message.Service, userSvc user.Service, wsMgr WSManager) {
	conv := r.Group("/api/conversations")
	conv.Use(AuthMiddleware(userSvc))
	{
		conv.GET("", handleGetConversations(msgSvc))
		conv.GET("/:id", handleGetConversation(msgSvc))
		conv.GET("/with/:user_id", handleGetConversationWithUser(msgSvc))
		conv.POST("/:id/read", handleMarkAsRead(msgSvc))
	}
}

// handleGetConversations 处理获取会话列表
func handleGetConversations(svc message.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		if limit > 20 {
			limit = 20
		}
		if limit <= 0 {
			limit = 20
		}

		// 调用消息服务
		conv, total, err := svc.GetConversations(auth.UserID, page, limit)
		if err != nil {
			InternalError(c, err)
			return
		}

		// 转换为响应格式
		convList := make([]ConversationResponse, len(conv))
		for i, cw := range conv {
			var participant *ParticipantResponse
			if cw.Participant.ID != auth.UserID {
				participant = &ParticipantResponse{
					ID:       cw.Participant.ID,
					Username: cw.Participant.Username,
					Nickname: cw.Participant.Nickname,
				}
			}

			var lastMsg *LastMessageResponse
			if cw.LastMessage != nil {
				lastMsg = &LastMessageResponse{
					ID:        cw.LastMessage.ID,
					Type:      cw.LastMessage.Type,
					Content:   cw.LastMessage.Content,
					SenderID:  cw.LastMessage.SenderID,
					CreatedAt: cw.LastMessage.CreatedAt,
				}
			}

			convList[i] = ConversationResponse{
				ID:           cw.ID,
				Participant:   participant,
				LastMessage:  lastMsg,
				UnreadCount:  cw.UnreadCount,
				UpdatedAt:    cw.UpdatedAt,
			}
		}

		SuccessList(c, convList, total)
	}
}

// handleGetConversation 处理获取会话详情
func handleGetConversation(svc message.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := GetAuthContext(c)
		if auth == nil {
			return
		}

		// 获取会话ID
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			BadRequest(c, "无效的会话ID")
			return
		}

		// 调用消息服务
		conv, err := svc.GetConversation(id, auth.UserID)
		if err != nil {
			handleMessageError(c, err)
			return
		}

		// 构建参与者信息
		var participant *ParticipantResponse
		if conv.Participant.ID != auth.UserID {
			participant = &ParticipantResponse{
				ID:       conv.Participant.ID,
				Username: conv.Participant.Username,
				Nickname: conv.Participant.Nickname,
			}
		}

		c.JSON(200, ConversationDetailResponse{
			ID:         conv.ID,
			Participant: participant,
			CreatedAt:  conv.CreatedAt,
			UpdatedAt:   conv.UpdatedAt,
		})
	}
}

// handleGetConversationWithUser 处理获取与指定用户的会话
func handleGetConversationWithUser(svc message.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := GetAuthContext(c)
		if auth == nil {
			return
		}

		// 获取对方用户ID
		otherIDStr := c.Param("user_id")
		otherID, err := strconv.ParseInt(otherIDStr, 10, 64)
		if err != nil {
			BadRequest(c, "无效的用户ID")
			return
		}

		if otherID == auth.UserID {
			BadRequest(c, "不能与自己创建会话")
			return
		}

		// 调用消息服务
		conv, err := svc.GetConversationWithUser(auth.UserID, otherID)
		if err != nil {
			InternalError(c, err)
			return
		}

		// 构建参与者信息
		var participant *ParticipantResponse
		if conv.Participant.ID != auth.UserID {
			participant = &ParticipantResponse{
				ID:       conv.Participant.ID,
				Username: conv.Participant.Username,
				Nickname: conv.Participant.Nickname,
			}
		}

		c.JSON(200, ConversationDetailResponse{
			ID:         conv.ID,
			Participant: participant,
			CreatedAt:  conv.CreatedAt,
			UpdatedAt:   conv.UpdatedAt,
		})
	}
}

// handleMarkAsRead 处理标记会话为已读
func handleMarkAsRead(svc message.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := GetAuthContext(c)
		if auth == nil {
			return
		}

		// 获取会话ID
		idStr := c.Param("id")
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			BadRequest(c, "无效的会话ID")
			return
		}

		// 调用消息服务
		err = svc.MarkAsRead(id, auth.UserID)
		if err != nil {
			InternalError(c, err)
			return
		}

		Success(c, map[string]bool{"success": true})
	}
}

// handleMessageError 处理消息服务错误
func handleMessageError(c *gin.Context, err error) {
	switch err.Error() {
	case "MSG_CONVERSATION_NOT_FOUND":
		NotFound(c, "会话不存在")
	case "MSG_USER_NOT_FOUND":
		NotFound(c, "用户不存在")
	default:
		InternalError(c, err)
	}
}

// ConversationResponse 会话响应
type ConversationResponse struct {
	ID           int64                `json:"id"`
	Participant   *ParticipantResponse   `json:"participant"`
	LastMessage  *LastMessageResponse  `json:"last_message"`
	UnreadCount  int                  `json:"unread_count"`
	UpdatedAt    int64                `json:"updated_at"`
}

// ConversationDetailResponse 会话详情响应
type ConversationDetailResponse struct {
	ID         int64              `json:"id"`
	Participant *ParticipantResponse `json:"participant"`
	CreatedAt  int64              `json:"created_at"`
	UpdatedAt  int64              `json:"updated_at"`
}

// ParticipantResponse 参与者响应
type ParticipantResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

// LastMessageResponse 最后消息响应
type LastMessageResponse struct {
	ID        int64  `json:"id"`
	Type      string `json:"type"`
	Content   string `json:"content"`
	SenderID  int64  `json:"sender_id"`
	CreatedAt int64  `json:"created_at"`
}
