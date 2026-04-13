package repository

import (
	"context"
	"database/sql"
	"fmt"

	"tech-feed-hub/article-service/internal/domain"
)

type SourceRepository struct {
	db *sql.DB
}

func NewSourceRepository(db *sql.DB) *SourceRepository {
	return &SourceRepository{db: db}
}

func (r *SourceRepository) EnsureSource(ctx context.Context, source domain.Source) (int64, error) {
	query := `
		INSERT INTO sources (name, type, fetch_url, fetch_method, interval_minutes, default_category, is_enabled)
		VALUES (?, ?, ?, ?, ?, ?, TRUE)
		ON DUPLICATE KEY UPDATE
			type = VALUES(type),
			fetch_url = VALUES(fetch_url),
			fetch_method = VALUES(fetch_method),
			interval_minutes = VALUES(interval_minutes),
			default_category = VALUES(default_category),
			updated_at = CURRENT_TIMESTAMP`

	result, err := r.db.ExecContext(ctx, query,
		source.Name,
		source.Type,
		source.FetchURL,
		source.FetchMethod,
		source.IntervalMinutes,
		source.DefaultCategory,
	)
	if err != nil {
		return 0, fmt.Errorf("ensure source: %w", err)
	}

	id, _ := result.LastInsertId()
	if id > 0 {
		return id, nil
	}

	var existingID int64
	if err := r.db.QueryRowContext(ctx, "SELECT id FROM sources WHERE name = ?", source.Name).Scan(&existingID); err != nil {
		return 0, fmt.Errorf("find ensured source: %w", err)
	}
	return existingID, nil
}

func (r *SourceRepository) UpdateFetchStatus(ctx context.Context, id int64, status string, errMsg *string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE sources
		SET last_fetched_at = CURRENT_TIMESTAMP,
		    last_fetch_status = ?,
		    last_error_message = ?
		WHERE id = ?`, status, errMsg, id)
	if err != nil {
		return fmt.Errorf("update source status: %w", err)
	}
	return nil
}
