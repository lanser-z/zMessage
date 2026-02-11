package dal

import (
	"database/sql"
	"fmt"
	"zmessage/server/models"
)

type conversationDAL struct {
	db DB
}

func NewConversationDAL(db DB) ConversationDAL {
	return &conversationDAL{db: db}
}

func (d *conversationDAL) Create(conv *models.Conversation) error {
	// 确保 user_a_id < user_b_id 以保证唯一性
	if conv.UserAID > conv.UserBID {
		conv.UserAID, conv.UserBID = conv.UserBID, conv.UserAID
	}

	query := `
		INSERT INTO conversations (user_a_id, user_b_id, created_at, updated_at)
		VALUES (?, ?, ?, ?)
	`
	result, err := d.db.Exec(query, conv.UserAID, conv.UserBID, conv.CreatedAt, conv.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create conversation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	conv.ID = id
	return nil
}

func (d *conversationDAL) GetByID(id int64) (*models.Conversation, error) {
	query := `
		SELECT id, user_a_id, user_b_id, created_at, updated_at
		FROM conversations WHERE id = ?
	`
	conv := &models.Conversation{}
	err := d.db.QueryRow(query, id).Scan(
		&conv.ID,
		&conv.UserAID,
		&conv.UserBID,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get conversation by id: %w", err)
	}
	return conv, nil
}

func (d *conversationDAL) GetByUsers(userA, userB int64) (*models.Conversation, error) {
	// 确保 user_a_id < user_b_id
	if userA > userB {
		userA, userB = userB, userA
	}

	query := `
		SELECT id, user_a_id, user_b_id, created_at, updated_at
		FROM conversations WHERE user_a_id = ? AND user_b_id = ?
	`
	conv := &models.Conversation{}
	err := d.db.QueryRow(query, userA, userB).Scan(
		&conv.ID,
		&conv.UserAID,
		&conv.UserBID,
		&conv.CreatedAt,
		&conv.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get conversation by users: %w", err)
	}
	return conv, nil
}

func (d *conversationDAL) GetByUser(userID int64, page, limit int) ([]*models.Conversation, int, error) {
	// 获取总数
	countQuery := `
		SELECT COUNT(*)
		FROM conversations
		WHERE user_a_id = ? OR user_b_id = ?
	`
	var total int
	err := d.db.QueryRow(countQuery, userID, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count conversations: %w", err)
	}

	if total == 0 {
		return []*models.Conversation{}, 0, nil
	}

	// 获取列表
	offset := (page - 1) * limit
	query := `
		SELECT id, user_a_id, user_b_id, created_at, updated_at
		FROM conversations
		WHERE user_a_id = ? OR user_b_id = ?
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`
	rows, err := d.db.Query(query, userID, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	var convs []*models.Conversation
	for rows.Next() {
		conv := &models.Conversation{}
		err := rows.Scan(
			&conv.ID,
			&conv.UserAID,
			&conv.UserBID,
			&conv.CreatedAt,
			&conv.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("scan conversation: %w", err)
		}
		convs = append(convs, conv)
	}

	return convs, total, nil
}

func (d *conversationDAL) Update(conv *models.Conversation) error {
	query := `
		UPDATE conversations
		SET user_a_id = ?, user_b_id = ?, updated_at = ?
		WHERE id = ?
	`
	result, err := d.db.Exec(query, conv.UserAID, conv.UserBID, conv.UpdatedAt, conv.ID)
	if err != nil {
		return fmt.Errorf("update conversation: %w", err)
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

func (d *conversationDAL) UpdateTime(id int64, updatedAt int64) error {
	query := `UPDATE conversations SET updated_at = ? WHERE id = ?`
	result, err := d.db.Exec(query, updatedAt, id)
	if err != nil {
		return fmt.Errorf("update conversation time: %w", err)
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

func (d *conversationDAL) Delete(id int64) error {
	query := `DELETE FROM conversations WHERE id = ?`
	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete conversation: %w", err)
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
