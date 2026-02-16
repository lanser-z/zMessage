package share

import "fmt"

var (
	// ErrShareNotFound 分享不存在
	ErrShareNotFound = fmt.Errorf("share not found")

	// ErrShareExpired 分享已过期
	ErrShareExpired = fmt.Errorf("share expired")

	// ErrAccessDenied 无权访问
	ErrAccessDenied = fmt.Errorf("access denied")

	// ErrInvalidMessageRange 无效的消息范围
	ErrInvalidMessageRange = fmt.Errorf("invalid message range")

	// ErrConversationNotFound 会话不存在
	ErrConversationNotFound = fmt.Errorf("conversation not found")
)
