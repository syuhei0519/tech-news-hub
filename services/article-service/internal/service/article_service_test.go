package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"tech-feed-hub/article-service/internal/domain"
)

type stubArticleRepo struct {
	bulkUpsertFunc func(ctx context.Context, sourceID int64, articles []domain.Article) (int, int, error)
}

func (r *stubArticleRepo) List(context.Context, domain.ListArticlesParams) (domain.ListArticlesResult, error) {
	return domain.ListArticlesResult{}, nil
}

func (r *stubArticleRepo) GetByID(context.Context, int64) (*domain.Article, error) {
	return nil, nil
}

func (r *stubArticleRepo) BulkUpsert(ctx context.Context, sourceID int64, articles []domain.Article) (int, int, error) {
	return r.bulkUpsertFunc(ctx, sourceID, articles)
}

type stubSourceRepo struct {
	ensureSourceFunc      func(ctx context.Context, source domain.Source) (int64, error)
	updateFetchStatusFunc func(ctx context.Context, id int64, status string, errMsg *string) error
}

func (r *stubSourceRepo) List(context.Context) ([]domain.Source, error) {
	return nil, nil
}

func (r *stubSourceRepo) GetByID(context.Context, int64) (*domain.Source, error) {
	return nil, nil
}

func (r *stubSourceRepo) Create(context.Context, domain.Source) (*domain.Source, error) {
	return nil, nil
}

func (r *stubSourceRepo) Update(context.Context, domain.Source) (*domain.Source, error) {
	return nil, nil
}

func (r *stubSourceRepo) Delete(context.Context, int64) (bool, error) {
	return false, nil
}

func (r *stubSourceRepo) EnsureSource(ctx context.Context, source domain.Source) (int64, error) {
	return r.ensureSourceFunc(ctx, source)
}

func (r *stubSourceRepo) UpdateFetchStatus(ctx context.Context, id int64, status string, errMsg *string) error {
	return r.updateFetchStatusFunc(ctx, id, status, errMsg)
}

type stubJobRepo struct {
	getByIDFunc func(ctx context.Context, id int64) (*domain.FetchJob, error)
	listFunc    func(ctx context.Context, params domain.ListFetchJobsParams) (domain.ListFetchJobsResult, error)
	createFunc  func(ctx context.Context, sourceID int64) (int64, error)
	finishFunc  func(ctx context.Context, jobID int64, status string, fetchedCount int, insertedCount int, duplicatedCount int, errMsg *string) error
}

func (r *stubJobRepo) Create(ctx context.Context, sourceID int64) (int64, error) {
	return r.createFunc(ctx, sourceID)
}

func (r *stubJobRepo) GetByID(ctx context.Context, id int64) (*domain.FetchJob, error) {
	return r.getByIDFunc(ctx, id)
}

func (r *stubJobRepo) List(ctx context.Context, params domain.ListFetchJobsParams) (domain.ListFetchJobsResult, error) {
	return r.listFunc(ctx, params)
}

func (r *stubJobRepo) Finish(ctx context.Context, jobID int64, status string, fetchedCount int, insertedCount int, duplicatedCount int, errMsg *string) error {
	return r.finishFunc(ctx, jobID, status, fetchedCount, insertedCount, duplicatedCount, errMsg)
}

func TestStartFetchJobEnsuresSourceAndCreatesJob(t *testing.T) {
	t.Parallel()

	service := &ArticleService{
		articleRepo: &stubArticleRepo{},
		sourceRepo: &stubSourceRepo{
			ensureSourceFunc: func(_ context.Context, source domain.Source) (int64, error) {
				if source.Name != "Kubernetes" {
					t.Fatalf("unexpected source: %+v", source)
				}
				return 12, nil
			},
			updateFetchStatusFunc: func(context.Context, int64, string, *string) error { return nil },
		},
		jobRepo: &stubJobRepo{
			createFunc: func(_ context.Context, sourceID int64) (int64, error) {
				if sourceID != 12 {
					t.Fatalf("unexpected source id: %d", sourceID)
				}
				return 34, nil
			},
			getByIDFunc: func(context.Context, int64) (*domain.FetchJob, error) { return nil, nil },
			listFunc: func(context.Context, domain.ListFetchJobsParams) (domain.ListFetchJobsResult, error) {
				return domain.ListFetchJobsResult{}, nil
			},
			finishFunc: func(context.Context, int64, string, int, int, int, *string) error { return nil },
		},
	}

	result, err := service.StartFetchJob(context.Background(), StartFetchJobInput{
		Source: IngestSourceInput{
			Name:            "Kubernetes",
			Type:            "rss",
			FetchURL:        "https://example.com/feed.xml",
			FetchMethod:     "rss",
			IntervalMinutes: 60,
			DefaultCategory: "k8s",
		},
	})
	if err != nil {
		t.Fatalf("StartFetchJob returned error: %v", err)
	}
	if result.SourceID != 12 || result.JobID != 34 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestFinishFetchJobUpdatesJobAndSourceStatus(t *testing.T) {
	t.Parallel()

	var updatedStatus string
	var updatedMessage *string

	service := &ArticleService{
		articleRepo: &stubArticleRepo{},
		sourceRepo: &stubSourceRepo{
			ensureSourceFunc: func(context.Context, domain.Source) (int64, error) { return 0, nil },
			updateFetchStatusFunc: func(_ context.Context, id int64, status string, errMsg *string) error {
				if id != 9 {
					t.Fatalf("unexpected source id: %d", id)
				}
				updatedStatus = status
				updatedMessage = errMsg
				return nil
			},
		},
		jobRepo: &stubJobRepo{
			createFunc: func(context.Context, int64) (int64, error) { return 0, nil },
			getByIDFunc: func(_ context.Context, id int64) (*domain.FetchJob, error) {
				return &domain.FetchJob{ID: id, SourceID: 9, Status: "running"}, nil
			},
			listFunc: func(context.Context, domain.ListFetchJobsParams) (domain.ListFetchJobsResult, error) {
				return domain.ListFetchJobsResult{}, nil
			},
			finishFunc: func(_ context.Context, jobID int64, status string, fetchedCount int, insertedCount int, duplicatedCount int, errMsg *string) error {
				if jobID != 21 || status != "failed" || fetchedCount != 5 || insertedCount != 2 || duplicatedCount != 3 {
					t.Fatalf("unexpected finish args: job=%d status=%s counts=%d/%d/%d", jobID, status, fetchedCount, insertedCount, duplicatedCount)
				}
				if errMsg == nil || *errMsg != "network error" {
					t.Fatalf("unexpected error message: %v", errMsg)
				}
				return nil
			},
		},
	}

	err := service.FinishFetchJob(context.Background(), 21, FinishFetchJobInput{
		Status:          "failed",
		FetchedCount:    5,
		InsertedCount:   2,
		DuplicatedCount: 3,
		ErrorMessage:    testStringPtr("network error"),
	})
	if err != nil {
		t.Fatalf("FinishFetchJob returned error: %v", err)
	}
	if updatedStatus != "failed" || updatedMessage == nil || *updatedMessage != "network error" {
		t.Fatalf("unexpected source update: status=%s err=%v", updatedStatus, updatedMessage)
	}
}

func testStringPtr(value string) *string {
	return &value
}

func TestIngestRejectsFinishedJob(t *testing.T) {
	t.Parallel()

	service := &ArticleService{
		articleRepo: &stubArticleRepo{
			bulkUpsertFunc: func(context.Context, int64, []domain.Article) (int, int, error) {
				t.Fatal("BulkUpsert should not be called")
				return 0, 0, nil
			},
		},
		sourceRepo: &stubSourceRepo{
			ensureSourceFunc:      func(context.Context, domain.Source) (int64, error) { return 0, nil },
			updateFetchStatusFunc: func(context.Context, int64, string, *string) error { return nil },
		},
		jobRepo: &stubJobRepo{
			createFunc: func(context.Context, int64) (int64, error) { return 0, nil },
			getByIDFunc: func(context.Context, int64) (*domain.FetchJob, error) {
				finishedAt := time.Now().UTC()
				return &domain.FetchJob{ID: 10, SourceID: 3, Status: "success", FinishedAt: &finishedAt}, nil
			},
			listFunc: func(context.Context, domain.ListFetchJobsParams) (domain.ListFetchJobsResult, error) {
				return domain.ListFetchJobsResult{}, nil
			},
			finishFunc: func(context.Context, int64, string, int, int, int, *string) error { return nil },
		},
	}

	_, err := service.Ingest(context.Background(), IngestRequest{
		JobID:    10,
		SourceID: 3,
		Source: IngestSourceInput{
			DefaultCategory: "k8s",
		},
	})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected ErrConflict, got: %v", err)
	}
}

func TestListFetchJobsRequiresSourceID(t *testing.T) {
	t.Parallel()

	service := &ArticleService{
		articleRepo: &stubArticleRepo{},
		sourceRepo: &stubSourceRepo{
			ensureSourceFunc:      func(context.Context, domain.Source) (int64, error) { return 0, nil },
			updateFetchStatusFunc: func(context.Context, int64, string, *string) error { return nil },
		},
		jobRepo: &stubJobRepo{
			createFunc:  func(context.Context, int64) (int64, error) { return 0, nil },
			getByIDFunc: func(context.Context, int64) (*domain.FetchJob, error) { return nil, nil },
			listFunc: func(context.Context, domain.ListFetchJobsParams) (domain.ListFetchJobsResult, error) {
				return domain.ListFetchJobsResult{}, nil
			},
			finishFunc: func(context.Context, int64, string, int, int, int, *string) error { return nil },
		},
	}

	_, err := service.ListFetchJobs(context.Background(), domain.ListFetchJobsParams{})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected ErrValidation, got: %v", err)
	}
}
