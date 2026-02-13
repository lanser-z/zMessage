package ws

import (
	"fmt"

	"zmessage/server/models"
	"zmessage/server/modules/message"
	"zmessage/server/modules/user"
	"zmessage/server/pkg/protocol"
)

// NewHandler 创建消息处理器
func NewHandler(msgSvc message.Service, userSvc user.Service) MessageHandler {
	return &handler{
		msgSvc:  msgSvc,
		userSvc: userSvc,
		mgr:     nil, // 需要通过SetManager注入
	}
}

// handler 消息处理器实现
type handler struct {
	msgSvc  message.Service
	userSvc user.Service
	mgr     Manager
}

// SetManager 设置连接管理器
func (h *handler) SetManager(mgr Manager) {
	h.mgr = mgr
}

// HandleMessage 处理消息（路由分发）
func (h *handler) HandleMessage(conn Connection, msg *protocol.WSMessage) error {
	switch msg.Type {
	case protocol.MsgAuth:
		return h.handleAuth(conn, msg)
	case protocol.MsgChat:
		return h.handleChat(conn, msg)
	case protocol.MsgAck:
		return h.handleAck(conn, msg)
	case protocol.MsgSyncReq:
		return h.handleSync(conn, msg)
	case protocol.MsgPresence:
		return h.handlePresence(conn, msg)
	case protocol.MsgPing:
		return h.handlePing(conn)
	default:
		return fmt.Errorf("unknown message type: %d", msg.Type)
	}
}

// handleAuth 处理认证消息
func (h *handler) handleAuth(conn Connection, msg *protocol.WSMessage) error {
	var payload protocol.AuthPayload
	if err := decodePayload(msg.Payload, &payload); err != nil {
		return err
	}

	// 验证Token
	userID, err := h.userSvc.ValidateToken(payload.Token)
	if err != nil {
		conn.Send(&protocol.WSMessage{
			Type:    protocol.MsgAuthRsp,
			Seq:     msg.Seq,
			Payload: h.encodeError("invalid_token"),
		})
		return nil
	}

	// 认证成功，设置用户ID
	if err := conn.authenticate(userID); err != nil {
		return err
	}

	// 设置在线状态
	h.userSvc.OnlineStatus().SetOnline(userID, true)

	// 返回成功响应
	conn.Send(&protocol.WSMessage{
		Type:    protocol.MsgAuthRsp,
		Seq:     msg.Seq,
		Payload: h.encodeSuccess(&protocol.AuthResponsePayload{
			Success: true,
			UserID:  userID,
		}),
	})
	return nil
}

// handleChat 处理聊天消息
func (h *handler) handleChat(conn Connection, msg *protocol.WSMessage) error {
	var payload protocol.ChatPayload
	if err := decodePayload(msg.Payload, &payload); err != nil {
		return err
	}

	from := conn.UserID()
	if from == 0 {
		conn.Send(&protocol.WSMessage{
			Type:    protocol.MsgError,
			Seq:     msg.Seq,
			Payload: h.encodeError("not_authenticated"),
		})
		return nil
	}

	// 发送消息
	sentMsg, err := h.msgSvc.SendMessage(&message.SendMessageRequest{
		From:    from,
		To:      payload.To,
		Type:    payload.Type,
		Content: payload.Content,
	})
	if err != nil {
		conn.Send(&protocol.WSMessage{
			Type:    protocol.MsgError,
			Seq:     msg.Seq,
			Payload: h.encodeError("send_failed"),
		})
		return nil
	}

	// 推送给接收者
	if h.mgr != nil {
		if toConn := h.mgr.GetConnection(payload.To); toConn != nil {
			toConn.Send(&protocol.WSMessage{
				Type: protocol.MsgChatPush,
				Payload: h.encodeChatPush(&protocol.ChatPushPayload{
					MessageID: sentMsg.ID,
					From:       from,
					To:         payload.To,
					Type:       payload.Type,
					Content:    sentMsg.Content,
					CreatedAt:  sentMsg.CreatedAt,
				}),
			})
		}
	}

	return nil
}

// handleAck 处理确认消息
func (h *handler) handleAck(conn Connection, msg *protocol.WSMessage) error {
	var payload protocol.AckPayload
	if err := decodePayload(msg.Payload, &payload); err != nil {
		return err
	}

	// 更新消息状态
	if err := h.msgSvc.UpdateMessageStatus(payload.MessageID, payload.Status); err != nil {
		return fmt.Errorf("update message status: %w", err)
	}

	return nil
}

// handleSync 处理同步请求
func (h *handler) handleSync(conn Connection, msg *protocol.WSMessage) error {
	var payload protocol.SyncRequestPayload
	if err := decodePayload(msg.Payload, &payload); err != nil {
		return err
	}

	from := conn.UserID()
	if from == 0 {
		conn.Send(&protocol.WSMessage{
			Type:    protocol.MsgError,
			Seq:     msg.Seq,
			Payload: h.encodeError("not_authenticated"),
		})
		return nil
	}

	// 获取离线消息
	messages, err := h.msgSvc.GetOfflineMessages(from, payload.LastMessageID, 100)
	if err != nil {
		conn.Send(&protocol.WSMessage{
			Type:    protocol.MsgError,
			Seq:     msg.Seq,
			Payload: h.encodeError("sync_failed"),
		})
		return nil
	}

	// 返回同步响应
	conn.Send(&protocol.WSMessage{
		Type: protocol.MsgSyncRsp,
		Seq:  msg.Seq,
		Payload: h.encodeSync(&protocol.SyncResponsePayload{
			Messages: h.encodeMessages(messages),
			HasMore:   len(messages) >= 100,
		}),
	})
	return nil
}

// handlePresence 处理在线状态
func (h *handler) handlePresence(conn Connection, msg *protocol.WSMessage) error {
	var payload protocol.PresencePayload
	if err := decodePayload(msg.Payload, &payload); err != nil {
		return err
	}

	from := conn.UserID()
	if from == 0 {
		return nil
	}

	// 更新用户在线状态
	h.userSvc.OnlineStatus().SetOnline(from, payload.Status == "online")

	// 广播在线状态变化给其他在线用户
	if h.mgr != nil {
		users := h.mgr.GetOnlineUsers()
		for _, uid := range users {
			if uid != from {
				if userConn := h.mgr.GetConnection(uid); userConn != nil {
					userConn.Send(&protocol.WSMessage{
						Type: protocol.MsgPresencePush,
						Payload: h.encodePresence(&protocol.PresencePushPayload{
							UserID: uid,
							Status: payload.Status,
						}),
					})
				}
			}
		}
	}

	return nil
}

// handlePing 处理心跳
func (h *handler) handlePing(conn Connection) error {
	// 响应Pong
	conn.Send(&protocol.WSMessage{
		Type: protocol.MsgPong,
	})
	return nil
}

// encodePayload 编码负载
func decodePayload(data []byte, dest interface{}) error {
	dec := NewDecoder()
	return dec.DecodePayload(data, dest)
}

// encodeSuccess 编码成功响应
func (h *handler) encodeSuccess(resp interface{}) []byte {
	enc := NewEncoder()
	data, _ := enc.EncodePayload(resp)
	return data
}

// encodeError 编码错误响应
func (h *handler) encodeError(err string) []byte {
	enc := NewEncoder()
	data, _ := enc.EncodePayload(&protocol.ErrorPayload{
		Code:    err,
		Message: "",
	})
	return data
}

// encodeChatPush 编码聊天推送消息
func (h *handler) encodeChatPush(msg *protocol.ChatPushPayload) []byte {
	enc := NewEncoder()
	data, _ := enc.EncodePayload(msg)
	return data
}

// encodeSync 编码同步响应消息
func (h *handler) encodeSync(msg *protocol.SyncResponsePayload) []byte {
	enc := NewEncoder()
	data, _ := enc.EncodePayload(msg)
	return data
}

// encodeMessages 编码消息列表
func (h *handler) encodeMessages(messages []*models.Message) []protocol.ChatPushPayload {
	result := make([]protocol.ChatPushPayload, len(messages))
	for i, msg := range messages {
		result[i] = protocol.ChatPushPayload{
			MessageID: msg.ID,
			From:       msg.SenderID,
			To:         msg.ReceiverID,
			Type:       msg.Type,
			Content:    msg.Content,
			CreatedAt:  msg.CreatedAt,
		}
	}
	return result
}

// encodePresence 编码在线状态
func (h *handler) encodePresence(msg *protocol.PresencePushPayload) []byte {
	enc := NewEncoder()
	data, _ := enc.EncodePayload(msg)
	return data
}
