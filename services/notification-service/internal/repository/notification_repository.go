package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	mysqlDriver "github.com/go-sql-driver/mysql"

	"tech-feed-hub/notification-service/internal/domain"
)

type NotificationRepository struct {
	db *sql.DB
}

func NewNotificationRepository(db *sql.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, input domain.CreateNotificationInput) (bool, error) {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO notifications (
			event_id,
			event_type,
			level,
			title,
			body,
			source_id,
			fetch_job_id,
			is_read,
			created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, FALSE, ?)`,
		input.EventID,
		input.EventType,
		input.Level,
		input.Title,
		input.Body,
		input.SourceID,
		input.FetchJobID,
		input.CreatedAt.UTC(),
	)
	if err != nil {
		var mysqlErr *mysqlDriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			return false, nil
		}
		return false, fmt.Errorf("insert notification: %w", err)
	}
	return true, nil
}

func (r *NotificationRepository) GetByID(ctx context.Context, id int64) (*domain.Notification, error) {
	row := r.db.QueryRowContext(
		ctx,
		`SELECT
			id,
			event_id,
			event_type,
			level,
			title,
			body,
			source_id,
			fetch_job_id,
			is_read,
			created_at,
			read_at
		FROM notifications
		WHERE id = ?`,
		id,
	)

	notification, err := scanNotification(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get notification by id: %w", err)
	}
	return notification, nil
}

func (r *NotificationRepository) List(ctx context.Context, params domain.ListNotificationsParams) (domain.ListNotificationsResult, error) {
	filter := make([]string, 0, 1)
	args := make([]any, 0, 4)
	if params.IsRead != nil {
		filter = append(filter, "is_read = ?")
		args = append(args, *params.IsRead)
	}

	whereClause := ""
	if len(filter) > 0 {
		whereClause = " WHERE " + strings.Join(filter, " AND ")
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM notifications" + whereClause
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return domain.ListNotificationsResult{}, fmt.Errorf("count notifications: %w", err)
	}

	offset := (params.Page - 1) * params.PageSize
	listArgs := append(append([]any{}, args...), params.PageSize, offset)
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			id,
			event_id,
			event_type,
			level,
			title,
			body,
			source_id,
			fetch_job_id,
			is_read,
			created_at,
			read_at
		FROM notifications`+whereClause+`
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?`,
		listArgs...,
	)
	if err != nil {
		return domain.ListNotificationsResult{}, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()

	items := make([]domain.Notification, 0, params.PageSize)
	for rows.Next() {
		item, err := scanNotification(rows)
		if err != nil {
			return domain.ListNotificationsResult{}, err
		}
		items = append(items, *item)
	}
	if err := rows.Err(); err != nil {
		return domain.ListNotificationsResult{}, fmt.Errorf("iterate notifications: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + params.PageSize - 1) / params.PageSize
	}

	return domain.ListNotificationsResult{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *NotificationRepository) UpdateReadStatus(ctx context.Context, id int64, isRead bool) (*domain.Notification, error) {
	var readAt any
	if isRead {
		readAt = time.Now().UTC()
	}

	result, err := r.db.ExecContext(
		ctx,
		`UPDATE notifications
		SET is_read = ?, read_at = ?
		WHERE id = ?`,
		isRead,
		readAt,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("update notification read status: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("notification rows affected: %w", err)
	}
	if affected == 0 {
		return nil, nil
	}

	return r.GetByID(ctx, id)
}

type scanner interface {
	Scan(dest ...any) error
}

func scanNotification(row scanner) (*domain.Notification, error) {
	var (
		notification domain.Notification
		sourceID     sql.NullInt64
		fetchJobID   sql.NullInt64
		readAt       sql.NullTime
	)

	if err := row.Scan(
		&notification.ID,
		&notification.EventID,
		&notification.EventType,
		&notification.Level,
		&notification.Title,
		&notification.Body,
		&sourceID,
		&fetchJobID,
		&notification.IsRead,
		&notification.CreatedAt,
		&readAt,
	); err != nil {
		return nil, err
	}

	if sourceID.Valid {
		notification.SourceID = &sourceID.Int64
	}
	if fetchJobID.Valid {
		notification.FetchJobID = &fetchJobID.Int64
	}
	if readAt.Valid {
		notification.ReadAt = &readAt.Time
	}
	return &notification, nil
}
