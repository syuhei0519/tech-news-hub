package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"
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
