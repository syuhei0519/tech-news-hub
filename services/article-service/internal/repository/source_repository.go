package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"tech-feed-hub/article-service/internal/domain"
)

type SourceRepository struct {
	db *sql.DB
}

func NewSourceRepository(db *sql.DB) *SourceRepository {
	return &SourceRepository{db: db}
}

func (r *SourceRepository) List(ctx context.Context) ([]domain.Source, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, type, fetch_url, fetch_method, interval_minutes, default_category,
		       is_enabled, last_fetched_at, last_fetch_status, last_error_message, created_at, updated_at
		FROM sources
		-- 一覧画面では最近触られた source を上に出す。
		ORDER BY updated_at DESC, id DESC`)
	if err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	defer rows.Close()

	sources := make([]domain.Source, 0)
	for rows.Next() {
		source, err := scanSource(rows)
		if err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sources: %w", err)
	}
	return sources, nil
}

func (r *SourceRepository) GetByID(ctx context.Context, id int64) (*domain.Source, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, type, fetch_url, fetch_method, interval_minutes, default_category,
		       is_enabled, last_fetched_at, last_fetch_status, last_error_message, created_at, updated_at
		FROM sources
		WHERE id = ?`, id)

	source, err := scanSource(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &source, nil
}

func (r *SourceRepository) Create(ctx context.Context, source domain.Source) (*domain.Source, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO sources (name, type, fetch_url, fetch_method, interval_minutes, default_category, is_enabled)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		source.Name,
		source.Type,
		source.FetchURL,
		source.FetchMethod,
		source.IntervalMinutes,
		source.DefaultCategory,
		source.IsEnabled,
	)
	if err != nil {
		return nil, mapSourceWriteError("create source", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("create source last insert id: %w", err)
	}
	// DB の default 値と timestamp を含んだ正規形を返す。
	return r.GetByID(ctx, id)
}

func (r *SourceRepository) Update(ctx context.Context, source domain.Source) (*domain.Source, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE sources
		SET name = ?, type = ?, fetch_url = ?, fetch_method = ?, interval_minutes = ?,
		    default_category = ?, is_enabled = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?`,
		source.Name,
		source.Type,
		source.FetchURL,
		source.FetchMethod,
		source.IntervalMinutes,
		source.DefaultCategory,
		source.IsEnabled,
		source.ID,
	)
	if err != nil {
		return nil, mapSourceWriteError("update source", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("update source rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil, nil
	}
	return r.GetByID(ctx, source.ID)
}

func (r *SourceRepository) Delete(ctx context.Context, id int64) (bool, error) {
	result, err := r.db.ExecContext(ctx, `DELETE FROM sources WHERE id = ?`, id)
	if err != nil {
		return false, mapSourceWriteError("delete source", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("delete source rows affected: %w", err)
	}
	return rowsAffected > 0, nil
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

type sourceScanner interface {
	Scan(dest ...any) error
}

func scanSource(scanner sourceScanner) (domain.Source, error) {
	var source domain.Source
	var lastFetchedAt sql.NullTime
	var lastFetchStatus sql.NullString
	var lastErrorMessage sql.NullString

	// source 一覧と詳細で同じ scan ロジックを使い、nullable 項目の扱いを揃える。
	err := scanner.Scan(
		&source.ID,
		&source.Name,
		&source.Type,
		&source.FetchURL,
		&source.FetchMethod,
		&source.IntervalMinutes,
		&source.DefaultCategory,
		&source.IsEnabled,
		&lastFetchedAt,
		&lastFetchStatus,
		&lastErrorMessage,
		&source.CreatedAt,
		&source.UpdatedAt,
	)
	if err != nil {
		return domain.Source{}, err
	}
	if lastFetchedAt.Valid {
		source.LastFetchedAt = &lastFetchedAt.Time
	}
	if lastFetchStatus.Valid {
		source.LastFetchStatus = &lastFetchStatus.String
	}
	if lastErrorMessage.Valid {
		source.LastErrorMsg = &lastErrorMessage.String
	}
	return source, nil
}

func mapSourceWriteError(operation string, err error) error {
	var mysqlErr *mysqlDriver.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1062:
			return fmt.Errorf("%s: duplicate: %w", operation, err)
		case 1451:
			return fmt.Errorf("%s: referenced: %w", operation, err)
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}
