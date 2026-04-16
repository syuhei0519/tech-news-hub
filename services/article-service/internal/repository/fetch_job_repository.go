package repository

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"tech-feed-hub/article-service/internal/domain"
)

type FetchJobRepository struct {
	db *sql.DB
}

func NewFetchJobRepository(db *sql.DB) *FetchJobRepository {
	return &FetchJobRepository{db: db}
}

func (r *FetchJobRepository) Create(ctx context.Context, sourceID int64) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO fetch_jobs (source_id, started_at, status)
		VALUES (?, ?, 'running')`, sourceID, time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("create fetch job: %w", err)
	}
	id, _ := result.LastInsertId()
	return id, nil
}

func (r *FetchJobRepository) GetByID(ctx context.Context, id int64) (*domain.FetchJob, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, source_id, started_at, finished_at, status, fetched_count, inserted_count, duplicated_count, error_message
		FROM fetch_jobs
		WHERE id = ?`, id)

	job, err := scanFetchJob(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get fetch job: %w", err)
	}
	return &job, nil
}

func (r *FetchJobRepository) List(ctx context.Context, params domain.ListFetchJobsParams) (domain.ListFetchJobsResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 || params.PageSize > 100 {
		params.PageSize = 20
	}

	where := []string{"1=1"}
	args := make([]any, 0, 2)
	if params.SourceID > 0 {
		where = append(where, "source_id = ?")
		args = append(args, params.SourceID)
	}
	if params.Status != "" {
		where = append(where, "status = ?")
		args = append(args, params.Status)
	}

	whereClause := strings.Join(where, " AND ")
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM fetch_jobs WHERE %s", whereClause)
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return domain.ListFetchJobsResult{}, fmt.Errorf("count fetch jobs: %w", err)
	}

	offset := (params.Page - 1) * params.PageSize
	listQuery := fmt.Sprintf(`
		SELECT id, source_id, started_at, finished_at, status, fetched_count, inserted_count, duplicated_count, error_message
		FROM fetch_jobs
		WHERE %s
		ORDER BY started_at DESC, id DESC
		LIMIT ? OFFSET ?`, whereClause)
	listArgs := append(append([]any{}, args...), params.PageSize, offset)
	rows, err := r.db.QueryContext(ctx, listQuery, listArgs...)
	if err != nil {
		return domain.ListFetchJobsResult{}, fmt.Errorf("list fetch jobs: %w", err)
	}
	defer rows.Close()

	items := make([]domain.FetchJob, 0)
	for rows.Next() {
		job, err := scanFetchJob(rows)
		if err != nil {
			return domain.ListFetchJobsResult{}, err
		}
		items = append(items, job)
	}
	if err := rows.Err(); err != nil {
		return domain.ListFetchJobsResult{}, fmt.Errorf("iterate fetch jobs: %w", err)
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(params.PageSize)))
	}

	return domain.ListFetchJobsResult{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (r *FetchJobRepository) Finish(ctx context.Context, jobID int64, status string, fetchedCount int, insertedCount int, duplicatedCount int, errMsg *string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE fetch_jobs
		SET finished_at = ?, status = ?, fetched_count = ?, inserted_count = ?, duplicated_count = ?, error_message = ?
		WHERE id = ?`,
		time.Now().UTC(),
		status,
		fetchedCount,
		insertedCount,
		duplicatedCount,
		errMsg,
		jobID,
	)
	if err != nil {
		return fmt.Errorf("finish fetch job: %w", err)
	}
	return nil
}

func scanFetchJob(scanner interface {
	Scan(dest ...any) error
}) (domain.FetchJob, error) {
	var job domain.FetchJob
	var finishedAt sql.NullTime
	var errorMessage sql.NullString

	if err := scanner.Scan(
		&job.ID,
		&job.SourceID,
		&job.StartedAt,
		&finishedAt,
		&job.Status,
		&job.FetchedCount,
		&job.InsertedCount,
		&job.DuplicatedCount,
		&errorMessage,
	); err != nil {
		return domain.FetchJob{}, err
	}

	if finishedAt.Valid {
		job.FinishedAt = &finishedAt.Time
	}
	if errorMessage.Valid {
		job.ErrorMessage = &errorMessage.String
	}
	return job, nil
}
