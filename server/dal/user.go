package dal

import (
	"database/sql"
	"fmt"
	"zmessage/server/models"
)

type userDAL struct {
	db DB
}

func NewUserDAL(db DB) UserDAL {
	return &userDAL{db: db}
}

func (d *userDAL) Create(user *models.User) error {
	query := `
		INSERT INTO users (username, password_hash, nickname, avatar_id, created_at, last_seen)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	result, err := d.db.Exec(query,
		user.Username,
		user.PasswordHash,
		user.Nickname,
		user.AvatarID,
		user.CreatedAt,
		user.LastSeen,
	)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	user.ID = id
	return nil
}

func (d *userDAL) GetByID(id int64) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, nickname, avatar_id, created_at, last_seen
		FROM users WHERE id = ?
	`
	user := &models.User{}
	err := d.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Nickname,
		&user.AvatarID,
		&user.CreatedAt,
		&user.LastSeen,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return user, nil
}

func (d *userDAL) GetByUsername(username string) (*models.User, error) {
	query := `
		SELECT id, username, password_hash, nickname, avatar_id, created_at, last_seen
		FROM users WHERE username = ?
	`
	user := &models.User{}
	err := d.db.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Nickname,
		&user.AvatarID,
		&user.CreatedAt,
		&user.LastSeen,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return user, nil
}

func (d *userDAL) List(search string, limit int) ([]*models.User, error) {
	query := `
		SELECT id, username, nickname, avatar_id, created_at, last_seen
		FROM users
		WHERE ? = '' OR username LIKE ? OR nickname LIKE ?
		ORDER BY created_at DESC
		LIMIT ?
	`
	searchPattern := ""
	if search != "" {
		searchPattern = "%" + search + "%"
	}

	rows, err := d.db.Query(query, search, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Nickname,
			&user.AvatarID,
			&user.CreatedAt,
			&user.LastSeen,
		)
		if err != nil {
			return nil, fmt.Errorf("scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (d *userDAL) Update(user *models.User) error {
	query := `
		UPDATE users
		SET nickname = ?, avatar_id = ?, last_seen = ?
		WHERE id = ?
	`
	result, err := d.db.Exec(query, user.Nickname, user.AvatarID, user.LastSeen, user.ID)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
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

func (d *userDAL) UpdateLastSeen(id int64, lastSeen int64) error {
	query := `UPDATE users SET last_seen = ? WHERE id = ?`
	result, err := d.db.Exec(query, lastSeen, id)
	if err != nil {
		return fmt.Errorf("update last seen: %w", err)
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

func (d *userDAL) Delete(id int64) error {
	query := `DELETE FROM users WHERE id = ?`
	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
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

func (d *userDAL) Count() (int, error) {
	query := `SELECT COUNT(*) FROM users`
	var count int
	err := d.db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}
