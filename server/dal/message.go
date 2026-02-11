package dal

import (
	"database/sql"
	"fmt"
	"zmessage/server/models"
)

type messageDAL struct {
	db DB
}

func NewMessageDAL(db DB) MessageDAL {
	return &messageDAL{db: db}
}

func (d *messageDAL) Create(msg *models.Message) error {
	query := `
		INSERT INTO messages (conversation_id, sender_id, receiver_id, type, content, status, created_at, synced_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := d.db.Exec(query,
		msg.ConversationID,
		msg.SenderID,
		msg.ReceiverID,
		msg.Type,
		msg.Content,
		msg.Status,
		msg.CreatedAt,
		msg.SyncedAt,
	)
	if err != nil {
		return fmt.Errorf("create message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	msg.ID = id
	return nil
}

func (d *messageDAL) GetByID(id int64) (*models.Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, receiver_id, type, content, status, created_at, synced_at
		FROM messages WHERE id = ?
	`
	msg := &models.Message{}
	err := d.db.QueryRow(query, id).Scan(
		&msg.ID,
		&msg.ConversationID,
		&msg.SenderID,
		&msg.ReceiverID,
		&msg.Type,
		&msg.Content,
		&msg.Status,
		&msg.CreatedAt,
		&msg.SyncedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get message by id: %w", err)
	}
	return msg, nil
}

func (d *messageDAL) GetByConversation(convID int64, beforeID int64, limit int) ([]*models.Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, receiver_id, type, content, status, created_at, synced_at
		FROM messages
		WHERE conversation_id = ?
	`
	args := []interface{}{convID}

	if beforeID > 0 {
		query += ` AND id < ?`
		args = append(args, beforeID)
	}

	query += ` ORDER BY created_at DESC LIMIT ?`
	args = append(args, limit)

	rows, err := d.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("get messages by conversation: %w", err)
	}
	defer rows.Close()

	var msgs []*models.Message
	for rows.Next() {
		msg := &models.Message{}
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Type,
			&msg.Content,
			&msg.Status,
			&msg.CreatedAt,
			&msg.SyncedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

func (d *messageDAL) GetOfflineMessages(userID int64, lastID int64, limit int) ([]*models.Message, error) {
	query := `
		SELECT id, conversation_id, sender_id, receiver_id, type, content, status, created_at, synced_at
		FROM messages
		WHERE receiver_id = ? AND id > ?
		ORDER BY id ASC
		LIMIT ?
	`
	rows, err := d.db.Query(query, userID, lastID, limit)
	if err != nil {
		return nil, fmt.Errorf("get offline messages: %w", err)
	}
	defer rows.Close()

	var msgs []*models.Message
	for rows.Next() {
		msg := &models.Message{}
		err := rows.Scan(
			&msg.ID,
			&msg.ConversationID,
			&msg.SenderID,
			&msg.ReceiverID,
			&msg.Type,
			&msg.Content,
			&msg.Status,
			&msg.CreatedAt,
			&msg.SyncedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		msgs = append(msgs, msg)
	}

	return msgs, nil
}

func (d *messageDAL) Update(msg *models.Message) error {
	query := `
		UPDATE messages
		SET status = ?, synced_at = ?
		WHERE id = ?
	`
	result, err := d.db.Exec(query, msg.Status, msg.SyncedAt, msg.ID)
	if err != nil {
		return fmt.Errorf("update message: %w", err)
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

func (d *messageDAL) UpdateStatus(id int64, status string) error {
	query := `UPDATE messages SET status = ? WHERE id = ?`
	result, err := d.db.Exec(query, status, id)
	if err != nil {
		return fmt.Errorf("update message status: %w", err)
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

func (d *messageDAL) UpdateMessagesStatus(convID int64, receiverID int64, status string) error {
	query := `
		UPDATE messages
		SET status = ?
		WHERE conversation_id = ? AND receiver_id = ? AND status != 'read'
	`
	_, err := d.db.Exec(query, status, convID, receiverID)
	if err != nil {
		return fmt.Errorf("update messages status: %w", err)
	}
	return nil
}

func (d *messageDAL) CountUnread(convID int64, userID int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages
		WHERE conversation_id = ? AND receiver_id = ? AND status != 'read'
	`
	var count int
	err := d.db.QueryRow(query, convID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count unread messages: %w", err)
	}
	return count, nil
}

func (d *messageDAL) CountTotalUnread(userID int64) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM messages
		WHERE receiver_id = ? AND status != 'read'
	`
	var count int
	err := d.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count total unread: %w", err)
	}
	return count, nil
}

func (d *messageDAL) Delete(id int64) error {
	query := `DELETE FROM messages WHERE id = ?`
	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete message: %w", err)
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
