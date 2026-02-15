package ws

import (
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"zmessage/server/modules/message"
	"zmessage/server/modules/user"
	"zmessage/server/pkg/protocol"
)

const (
	// PingInterval 心跳间隔
	PingInterval = 30 * time.Second
	// PongTimeout 心跳超时（增加到5分钟，避免频繁断开）
	PongTimeout = 300 * time.Second
	// MaxConnectionsPerUser 每用户最大连接数
	MaxConnectionsPerUser = 3
)

// NewManager 创建连接管理器
func NewManager(msgSvc message.Service, userSvc user.Service) Manager {
	mgr := &connectionManager{
		connections: make(map[string]*connection),
		userConns:   make(map[int64]map[string]*connection),
		msgHandler:  &handler{
			msgSvc:  msgSvc,
			userSvc: userSvc,
		},
	}
	// 注入manager到handler
	mgr.msgHandler.(*handler).SetManager(mgr)
	return mgr
}

// connectionManager 连接管理器实现
type connectionManager struct {
	connections map[string]*connection                 // connID -> connection
	userConns   map[int64]map[string]*connection   // userID -> connIDs
	mu          sync.RWMutex
	msgHandler   MessageHandler
}

// HandleConnection 处理新的WebSocket连接
func (m *connectionManager) HandleConnection(conn interface{}) {
	wsConn, ok := conn.(*websocket.Conn)
	if !ok {
		return
	}
	// 创建连接对象
	c := &connection{
		id:       generateConnID(),
		conn:     wsConn,
		send:     make(chan []byte, 256),
		mgr:      m,
		userID:   0,
	 createdAt: time.Now(),
	 pongTime:  time.Now(),
	}

	m.mu.Lock()
	m.connections[c.id] = c
	m.mu.Unlock()

	// 启动读协程
	go c.readPump()
	// 启动写协程
	go c.writePump()
}

// GetConnection 获取用户的连接（返回第一个）
func (m *connectionManager) GetConnection(userID int64) Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if conns, ok := m.userConns[userID]; ok && len(conns) > 0 {
		for _, c := range conns {
			return c
		}
	}
	return nil
}

// GetConnections 获取用户的所有连接
func (m *connectionManager) GetConnections(userID int64) []Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if conns, ok := m.userConns[userID]; ok {
		result := make([]Connection, 0, len(conns))
		for _, c := range conns {
			result = append(result, c)
		}
		return result
	}
	return nil
}

// BroadcastToUser 向用户的所有连接发送消息
func (m *connectionManager) BroadcastToUser(userID int64, msg *protocol.WSMessage) error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conns, ok := m.userConns[userID]
	if !ok || len(conns) == 0 {
		return nil // 用户离线，不报错
	}

	// 编码消息
	enc := NewEncoder()
	data, err := enc.Encode(msg)
	if err != nil {
		return err
	}

	for _, c := range conns {
		select {
		case c.send <- data:
			// 发送成功
		default:
			return fmt.Errorf("user %d channel full", userID)
		}
	}
	return nil
}

// IsOnline 检查用户是否在线
func (m *connectionManager) IsOnline(userID int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conns, ok := m.userConns[userID]
	return ok && len(conns) > 0
}

// GetOnlineUsers 获取在线用户列表
func (m *connectionManager) GetOnlineUsers() []int64 {
	m.mu.RLock()
	defer m.mu.RUnlock()

	users := make([]int64, 0)
	for userID := range m.userConns {
		if len(m.userConns[userID]) > 0 {
			users = append(users, userID)
		}
	}
	return users
}

// Disconnect 断开指定连接
func (m *connectionManager) Disconnect(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if c, ok := m.connections[connID]; ok {
		c.close()
		delete(m.connections, connID)

		// 从用户连接列表中移除
		if c.userID != 0 {
			if conns, ok := m.userConns[c.userID]; ok {
				delete(conns, connID)
				if len(conns) == 0 {
					delete(m.userConns, c.userID)
				}
			}
		}
	}
}

// DisconnectUser 断开用户的所有连接
func (m *connectionManager) DisconnectUser(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conns, ok := m.userConns[userID]; ok {
		for connID, c := range conns {
			c.close()
			delete(m.connections, connID)
		}
	}
	delete(m.userConns, userID)
}

// addConnection 将连接添加到用户连接列表
func (m *connectionManager) addConnection(c *connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.userConns[c.userID]; !ok {
		m.userConns[c.userID] = make(map[string]*connection)
	}
	m.userConns[c.userID][c.id] = c
}

// removeConnection 从用户连接列表中移除连接
func (m *connectionManager) removeConnection(c *connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conns, ok := m.userConns[c.userID]; ok {
		delete(conns, c.id)
		if len(conns) == 0 {
			delete(m.userConns, c.userID)
		}
	}
}

// generateConnID 生成连接ID
func generateConnID() string {
	return fmt.Sprintf("conn_%d", time.Now().UnixNano())
}

// connection WebSocket连接实现
type connection struct {
	id        string
	conn      *websocket.Conn
	send      chan []byte
	mgr       *connectionManager
	userID    int64
	mu        sync.Mutex
	createdAt time.Time
	pongTime  time.Time
}

func (c *connection) readPump() {
	defer c.close()

	c.conn.SetReadDeadline(time.Now().Add(PongTimeout))
	c.conn.SetPongHandler(func(string) error {
		c.mu.Lock()
		c.pongTime = time.Now()
		c.mu.Unlock()
		return nil
	})

	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err) || websocket.IsCloseError(err) {
				return // 正常关闭
			}
			return
		}

		if err := c.handleMessage(data); err != nil {
			c.sendError(fmt.Sprintf("handle message error: %v", err))
			return
		}
	}
}

func (c *connection) writePump() {
	defer c.close()

	ticker := time.NewTicker(PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case data, ok := <-c.send:
			if !ok {
				return
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				return
			}
		}
	}
}

func (c *connection) handleMessage(data []byte) error {
	msg, err := decodeWSMessage(data)
	if err != nil {
		return fmt.Errorf("decode message: %w", err)
	}

	return c.mgr.msgHandler.HandleMessage(c, msg)
}

func (c *connection) sendError(errMsg string) {
	enc := NewEncoder()
	data, _ := enc.Encode(&protocol.WSMessage{
		Type:    protocol.MsgError,
		Payload: encodeErrorPayload(errMsg),
	})
	select {
	case c.send <- data:
	default:
	}
}

func decodeWSMessage(data []byte) (*protocol.WSMessage, error) {
	dec := NewDecoder()
	return dec.Decode(data)
}

func encodeErrorPayload(message string) []byte {
	enc := NewEncoder()
	data, _ := enc.EncodePayload(&protocol.ErrorPayload{
		Code:    message,
		Message: "",
	})
	return data
}

// ID 获取连接ID
func (c *connection) ID() string {
	return c.id
}

// UserID 获取用户ID
func (c *connection) UserID() int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.userID
}

// Send 发送消息
func (c *connection) Send(msg *protocol.WSMessage) error {
	enc := NewEncoder()
	data, err := enc.Encode(msg)
	if err != nil {
		return err
	}

	select {
	case c.send <- data:
		return nil
	default:
		return fmt.Errorf("connection %s send buffer full", c.id)
	}
}

// authenticate 认证连接
func (c *connection) authenticate(userID int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查用户连接数限制
	c.mgr.mu.Lock()
	conns := c.mgr.userConns[userID]
	if len(conns) >= MaxConnectionsPerUser {
		c.mgr.mu.Unlock()
		return fmt.Errorf("max connections per user exceeded")
	}
	c.mgr.mu.Unlock()

	c.userID = userID
	c.mgr.addConnection(c)
	return nil
}

// close 关闭连接
func (c *connection) close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 先关闭 websocket
	c.conn.Close()

	// 安全地关闭 channel（避免重复关闭）
	if c.send != nil {
		select {
		case <-c.send:
			// channel 已关闭或为空
		default:
			close(c.send)
		}
		c.send = nil
	}

	// 从管理器中移除连接
	if c.userID != 0 {
		c.mgr.removeConnection(c)
	}
}
