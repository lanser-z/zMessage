package dal

import (
	"database/sql"
	"fmt"
	"time"
	"zmessage/server/models"
)

type sharedConversationDAL struct {
	db DB
}

func NewSharedConversationDAL(db DB) SharedConversationDAL {
	return &sharedConversationDAL{db: db}
}

func (d *sharedConversationDAL) Create(sc *models.SharedConversation) error {
	query := `
		INSERT INTO shared_conversations
		(conversation_id, share_token, created_by, expire_at, created_at, first_message_id, last_message_id, message_count, view_count)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := d.db.Exec(query,
		sc.ConversationID,
		sc.ShareToken,
		sc.CreatedBy,
		sc.ExpireAt,
		sc.CreatedAt,
		sc.FirstMessageID,
		sc.LastMessageID,
		sc.MessageCount,
		sc.ViewCount,
	)
	if err != nil {
		return fmt.Errorf("create shared conversation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	sc.ID = id
	return nil
}

func (d *sharedConversationDAL) GetByID(id int64) (*models.SharedConversation, error) {
	query := `
		SELECT id, conversation_id, share_token, created_by, expire_at, created_at,
		       first_message_id, last_message_id, message_count, view_count
		FROM shared_conversations WHERE id = ?
	`
	sc := &models.SharedConversation{}
	err := d.db.QueryRow(query, id).Scan(
		&sc.ID,
		&sc.ConversationID,
		&sc.ShareToken,
		&sc.CreatedBy,
		&sc.ExpireAt,
		&sc.CreatedAt,
		&sc.FirstMessageID,
		&sc.LastMessageID,
		&sc.MessageCount,
		&sc.ViewCount,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get shared conversation by id: %w", err)
	}
	return sc, nil
}

func (d *sharedConversationDAL) GetByToken(token string) (*models.SharedConversation, error) {
	query := `
		SELECT id, conversation_id, share_token, created_by, expire_at, created_at,
		       first_message_id, last_message_id, message_count, view_count
		FROM shared_conversations WHERE share_token = ?
	`
	sc := &models.SharedConversation{}
	err := d.db.QueryRow(query, token).Scan(
		&sc.ID,
		&sc.ConversationID,
		&sc.ShareToken,
		&sc.CreatedBy,
		&sc.ExpireAt,
		&sc.CreatedAt,
		&sc.FirstMessageID,
		&sc.LastMessageID,
		&sc.MessageCount,
		&sc.ViewCount,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get shared conversation by token: %w", err)
	}

	// 检查是否过期
	if sc.ExpireAt > 0 && time.Now().Unix() > sc.ExpireAt {
		return nil, ErrNotFound
	}

	return sc, nil
}

func (d *sharedConversationDAL) GetByConversation(convID int64) ([]*models.SharedConversation, error) {
	query := `
		SELECT id, conversation_id, share_token, created_by, expire_at, created_at,
		       first_message_id, last_message_id, message_count, view_count
		FROM shared_conversations WHERE conversation_id = ?
		ORDER BY created_at DESC
	`
	rows, err := d.db.Query(query, convID)
	if err != nil {
		return nil, fmt.Errorf("get shared conversations: %w", err)
	}
	defer rows.Close()

	var scs []*models.SharedConversation
	for rows.Next() {
		sc := &models.SharedConversation{}
		err := rows.Scan(
			&sc.ID,
			&sc.ConversationID,
			&sc.ShareToken,
			&sc.CreatedBy,
			&sc.ExpireAt,
			&sc.CreatedAt,
			&sc.FirstMessageID,
			&sc.LastMessageID,
			&sc.MessageCount,
			&sc.ViewCount,
		)
		if err != nil {
			return nil, fmt.Errorf("scan shared conversation: %w", err)
		}
		scs = append(scs, sc)
	}

	return scs, nil
}

func (d *sharedConversationDAL) GetByCreator(creatorID int64, page, limit int) ([]*models.SharedConversation, int, error) {
	// 获取总数
	var total int
	countQuery := `SELECT COUNT(*) FROM shared_conversations WHERE created_by = ?`
	err := d.db.QueryRow(countQuery, creatorID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count shares: %w", err)
	}

	// 获取分页数据
	offset := (page - 1) * limit
	query := `
		SELECT id, conversation_id, share_token, created_by, expire_at, created_at,
		       first_message_id, last_message_id, message_count, view_count
		FROM shared_conversations
		WHERE created_by = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := d.db.Query(query, creatorID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("get shares by creator: %w", err)
	}
	defer rows.Close()

	var scs []*models.SharedConversation
	for rows.Next() {
		sc := &models.SharedConversation{}
		err := rows.Scan(
			&sc.ID,
			&sc.ConversationID,
			&sc.ShareToken,
			&sc.CreatedBy,
			&sc.ExpireAt,
			&sc.CreatedAt,
			&sc.FirstMessageID,
			&sc.LastMessageID,
			&sc.MessageCount,
			&sc.ViewCount,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan share: %w", err)
		}
		scs = append(scs, sc)
	}

	return scs, total, nil
}

func (d *sharedConversationDAL) UpdateViewCount(id int64, count int) error {
	query := `UPDATE shared_conversations SET view_count = ? WHERE id = ?`
	_, err := d.db.Exec(query, count, id)
	if err != nil {
		return fmt.Errorf("update view count: %w", err)
	}
	return nil
}

func (d *sharedConversationDAL) Delete(id int64) error {
	query := `DELETE FROM shared_conversations WHERE id = ?`
	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete shared conversation: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (d *sharedConversationDAL) DeleteExpired() error {
	query := `DELETE FROM shared_conversations WHERE expire_at > 0 AND expire_at < ?`
	_, err := d.db.Exec(query, time.Now().Unix())
	if err != nil {
		return fmt.Errorf("delete expired shared conversations: %w", err)
	}
	return nil
}
