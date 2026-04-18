package repository

import (
	"context"
	"testing"
	"time"

	"tech-feed-hub/article-service/internal/domain"
	"tech-feed-hub/article-service/internal/testutil"
)

func TestFetchJobRepositoryCreateFinishAndList(t *testing.T) {
	db := testutil.OpenMySQLForTest(t)
	testutil.ResetMySQLTables(t, db)

	repo := NewFetchJobRepository(db)
	sourceID := testutil.InsertSource(t, db, domain.Source{
		Name:            "Fetch Source",
		Type:            "rss",
		FetchURL:        "https://example.com/fetch.xml",
		FetchMethod:     "rss",
		IntervalMinutes: 20,
		DefaultCategory: "news",
		IsEnabled:       true,
	})

	jobID, err := repo.Create(context.Background(), sourceID)
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	job, err := repo.GetByID(context.Background(), jobID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if job == nil || job.Status != "running" || job.SourceID != sourceID {
		t.Fatalf("unexpected created job: %+v", job)
	}

	errMsg := "collector timeout"
	if err := repo.Finish(context.Background(), jobID, "failed", 12, 5, 7, &errMsg); err != nil {
		t.Fatalf("Finish returned error: %v", err)
	}

	finished, err := repo.GetByID(context.Background(), jobID)
	if err != nil {
		t.Fatalf("GetByID after finish returned error: %v", err)
	}
	if finished == nil || finished.FinishedAt == nil || finished.Status != "failed" || finished.ErrorMessage == nil || *finished.ErrorMessage != errMsg {
		t.Fatalf("unexpected finished job: %+v", finished)
	}

	olderStart := time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC)
	newerStart := time.Date(2026, 4, 11, 9, 0, 0, 0, time.UTC)
	testutil.InsertFetchJob(t, db, domain.FetchJob{
		SourceID:        sourceID,
		StartedAt:       olderStart,
		Status:          "success",
		FetchedCount:    3,
		InsertedCount:   2,
		DuplicatedCount: 1,
	})
	testutil.InsertFetchJob(t, db, domain.FetchJob{
		SourceID:        sourceID,
		StartedAt:       newerStart,
		Status:          "success",
		FetchedCount:    4,
		InsertedCount:   4,
		DuplicatedCount: 0,
	})

	listed, err := repo.List(context.Background(), domain.ListFetchJobsParams{
		SourceID: sourceID,
		Status:   "success",
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if listed.Total != 2 || listed.TotalPages != 2 || len(listed.Items) != 1 {
		t.Fatalf("unexpected list meta: %+v", listed)
	}
	if !listed.Items[0].StartedAt.Equal(newerStart) {
		t.Fatalf("expected newest success job first, got %+v", listed.Items[0])
	}
}
