package dal

import (
	"database/sql"
	"zmessage/server/models"
)

// DB 数据库接口
type DB interface {
	Close() error
	Begin() (*sql.Tx, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

// Manager 数据库管理器接口
type Manager interface {
	// DB 获取数据库连接
	DB() DB

	// User 用户数据访问
	User() UserDAL

	// Conversation 会话数据访问
	Conversation() ConversationDAL

	// Message 消息数据访问
	Message() MessageDAL

	// Media 媒体数据访问
	Media() MediaDAL

	// SharedConversation 分享数据访问
	SharedConversation() SharedConversationDAL

	// Close 关闭数据库连接
	Close() error
}

// UserDAL 用户数据访问接口
type UserDAL interface {
	Create(user *models.User) error
	GetByID(id int64) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	List(search string, limit int) ([]*models.User, error)
	Update(user *models.User) error
	UpdateLastSeen(id int64, lastSeen int64) error
	Delete(id int64) error
	Count() (int, error)
}

// ConversationDAL 会话数据访问接口
type ConversationDAL interface {
	Create(conv *models.Conversation) error
	GetByID(id int64) (*models.Conversation, error)
	GetByUsers(userA, userB int64) (*models.Conversation, error)
	GetByUser(userID int64, page, limit int) ([]*models.Conversation, int, error)
	Update(conv *models.Conversation) error
	UpdateTime(id int64, updatedAt int64) error
	Delete(id int64) error
}

// MessageDAL 消息数据访问接口
type MessageDAL interface {
	Create(msg *models.Message) error
	GetByID(id int64) (*models.Message, error)
	GetByConversation(convID int64, beforeID int64, limit int) ([]*models.Message, error)
	GetOfflineMessages(userID int64, lastID int64, limit int) ([]*models.Message, error)
	Update(msg *models.Message) error
	UpdateStatus(id int64, status string) error
	UpdateMessagesStatus(convID int64, receiverID int64, status string) error
	CountUnread(convID int64, userID int64) (int, error)
	CountTotalUnread(userID int64) (int, error)
	Delete(id int64) error
}

// MediaDAL 媒体数据访问接口
type MediaDAL interface {
	Create(media *models.Media) error
	GetByID(id int64) (*models.Media, error)
	GetByOwner(ownerID int64) ([]*models.Media, error)
	Update(media *models.Media) error
	Delete(id int64) error
}

// SharedConversationDAL 分享数据访问接口
type SharedConversationDAL interface {
	Create(sc *models.SharedConversation) error
	GetByID(id int64) (*models.SharedConversation, error)
	GetByToken(token string) (*models.SharedConversation, error)
	GetByConversation(convID int64) ([]*models.SharedConversation, error)
	GetByCreator(creatorID int64, page, limit int) ([]*models.SharedConversation, int, error)
	UpdateViewCount(id int64, count int) error
	Delete(id int64) error
	DeleteExpired() error
}
