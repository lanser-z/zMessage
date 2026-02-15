package sse

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"zmessage/server/modules/user"
)

// Handler SSE 处理器
type Handler struct {
	userSvc user.Service
}

// NewHandler 创建 SSE 处理器
func NewHandler(userSvc user.Service) *Handler {
	return &Handler{userSvc: userSvc}
}

// PushMessage 推送消息的通道类型
type PushMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// clients 存储所有 SSE 客户端连接
// 每个用户可以有多个连接（多个标签页/设备）
var (
	clients  = make(map[int64]map[chan *PushMessage]bool) // userID -> channels
	clientsMu sync.RWMutex
)

// Broadcast 广播消息给指定用户的所有连接
func Broadcast(userID int64, msgType string, data interface{}) {
	pushMsg := &PushMessage{
		Type: msgType,
		Data: data,
	}

	fmt.Printf("[SSE Broadcast] Broadcasting to user %d, type: %s, data: %+v\n", userID, msgType, data)

	clientsMu.RLock()
	userChans, ok := clients[userID]
	clientsMu.RUnlock()

	if !ok {
		fmt.Printf("[SSE Broadcast] User %d has no active connections\n", userID)
		return
	}

	sentCount := 0
	for ch := range userChans {
		select {
		case ch <- pushMsg:
			sentCount++
			fmt.Printf("[SSE Broadcast] Sent to one channel (total: %d)\n", sentCount)
		default:
			// 如果通道满了，忽略
			fmt.Printf("[SSE Broadcast] Channel full, message dropped\n")
		}
	}
	fmt.Printf("[SSE Broadcast] Completed broadcasting to user %d, sent to %d channels\n", userID, sentCount)
}

// Subscribe SSE 订阅端点
func (h *Handler) Subscribe(c *gin.Context) {
	// 从 URL 参数获取 token
	token := c.Query("token")
	if token == "" {
		c.JSON(401, gin.H{"error": "token required"})
		return
	}

	// 验证 token 获取用户 ID
	userID, err := h.userSvc.ValidateToken(token)
	if err != nil {
		c.JSON(401, gin.H{"error": "invalid token"})
		return
	}

	// 设置 SSE 头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("X-Accel-Buffering", "no") // 禁用 Nginx 缓冲

	// 获取 flusher
	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(500, gin.H{"error": "streaming not supported"})
		return
	}

	// 创建消息通道
	msgChan := make(chan *PushMessage, 100)

	// 注册连接
	clientsMu.Lock()
	if clients[userID] == nil {
		clients[userID] = make(map[chan *PushMessage]bool)
	}
	clients[userID][msgChan] = true
	connCount := len(clients[userID])
	clientsMu.Unlock()

	// 发送连接成功事件
	fmt.Fprintf(c.Writer, "event: connected\n")
	fmt.Fprintf(c.Writer, "data: {\"user_id\":%d,\"conn_count\":%d}\n\n", userID, connCount)
	flusher.Flush()

	// 监听连接断开
	notify := c.Request.Context().Done()

	// 保持连接并推送消息
	defer func() {
		close(msgChan)

		clientsMu.Lock()
		if clients[userID] != nil {
			delete(clients[userID], msgChan)
			// 如果该用户没有连接了，清理map
			if len(clients[userID]) == 0 {
				delete(clients, userID)
			}
		}
		clientsMu.Unlock()

		fmt.Printf("[SSE] User %d disconnected (remaining: %d)\n", userID, len(clients[userID]))
	}()

	clientIP := c.ClientIP()
	fmt.Printf("[SSE] User %d subscribed from %s (total connections: %d)\n", userID, clientIP, connCount)

	// 发送心跳（每 15 秒）
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 发送心跳
			fmt.Fprintf(c.Writer, "event: heartbeat\n")
			fmt.Fprintf(c.Writer, "data: {\"timestamp\":%d,\"interval\":15}\n\n", time.Now().Unix())
			flusher.Flush()
			fmt.Printf("[SSE] Heartbeat sent to user %d\n", userID)
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			// 发送消息
			jsonData, _ := json.Marshal(msg.Data)
			fmt.Printf("[SSE] Sending to user %d: event=%s, data=%s\n", userID, msg.Type, string(jsonData))
			fmt.Fprintf(c.Writer, "event: %s\n", msg.Type)
			fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
			flusher.Flush()
			fmt.Printf("[SSE] Message sent to user %d successfully\n", userID)
		case <-notify:
			return
		}
	}
}
