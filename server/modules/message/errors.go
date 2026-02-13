package message

import (
	"fmt"
)

var (
	// ErrInvalidMessageType 无效的消息类型
	ErrInvalidMessageType = fmt.Errorf("invalid message type")

	// ErrInvalidMessageContent 无效的消息内容
	ErrInvalidMessageContent = fmt.Errorf("invalid message content")

	// ErrConversationNotFound 会话不存在
	ErrConversationNotFound = fmt.Errorf("conversation not found")

	// ErrMessageNotFound 消息不存在
	ErrMessageNotFound = fmt.Errorf("message not found")

	// ErrUserNotFound 用户不存在
	ErrUserNotFound = fmt.Errorf("user not found")

	// ErrSendToSelf 不能给自己发消息
	ErrSendToSelf = fmt.Errorf("cannot send message to yourself")

	// ErrAccessDenied 无权访问
	ErrAccessDenied = fmt.Errorf("access denied")
)
