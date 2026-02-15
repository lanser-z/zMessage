package dal

import (
	"database/sql"
	"fmt"
	"zmessage/server/models"
)

type mediaDAL struct {
	db DB
}

func NewMediaDAL(db DB) MediaDAL {
	return &mediaDAL{db: db}
}

func (d *mediaDAL) Create(media *models.Media) error {
	query := `
		INSERT INTO media (owner_id, type, original_path, thumbnail_path, size, mime_type, width, height, duration, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := d.db.Exec(query,
		media.OwnerID,
		media.Type,
		media.OriginalPath,
		media.ThumbnailPath,
		media.Size,
		media.MimeType,
		media.Width,
		media.Height,
		media.Duration,
		media.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("create media: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get last insert id: %w", err)
	}

	media.ID = id
	return nil
}

func (d *mediaDAL) GetByID(id int64) (*models.Media, error) {
	query := `
		SELECT id, owner_id, type, original_path, thumbnail_path, size, mime_type, width, height, duration, created_at
		FROM media WHERE id = ?
	`
	media := &models.Media{}
	err := d.db.QueryRow(query, id).Scan(
		&media.ID,
		&media.OwnerID,
		&media.Type,
		&media.OriginalPath,
		&media.ThumbnailPath,
		&media.Size,
		&media.MimeType,
		&media.Width,
		&media.Height,
		&media.Duration,
		&media.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get media by id: %w", err)
	}
	return media, nil
}

func (d *mediaDAL) GetByOwner(ownerID int64) ([]*models.Media, error) {
	query := `
		SELECT id, owner_id, type, original_path, thumbnail_path, size, mime_type, width, height, duration, created_at
		FROM media WHERE owner_id = ?
		ORDER BY created_at DESC
	`
	rows, err := d.db.Query(query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("get media by owner: %w", err)
	}
	defer rows.Close()

	var mediaList []*models.Media
	for rows.Next() {
		media := &models.Media{}
		err := rows.Scan(
			&media.ID,
			&media.OwnerID,
			&media.Type,
			&media.OriginalPath,
			&media.ThumbnailPath,
			&media.Size,
			&media.MimeType,
			&media.Width,
			&media.Height,
			&media.Duration,
			&media.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan media: %w", err)
		}
		mediaList = append(mediaList, media)
	}

	return mediaList, nil
}

func (d *mediaDAL) Update(media *models.Media) error {
	query := `
		UPDATE media
		SET original_path = ?, thumbnail_path = ?, width = ?, height = ?, duration = ?
		WHERE id = ?
	`
	result, err := d.db.Exec(query,
		media.OriginalPath,
		media.ThumbnailPath,
		media.Width,
		media.Height,
		media.Duration,
		media.ID,
	)
	if err != nil {
		return fmt.Errorf("update media: %w", err)
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

func (d *mediaDAL) Delete(id int64) error {
	query := `DELETE FROM media WHERE id = ?`
	result, err := d.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("delete media: %w", err)
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
