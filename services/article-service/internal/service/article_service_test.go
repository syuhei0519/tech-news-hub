package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"tech-feed-hub/article-service/internal/domain"
)

type stubArticleRepo struct {
	listFunc                 func(ctx context.Context, params domain.ListArticlesParams) (domain.ListArticlesResult, error)
	getByIDFunc              func(ctx context.Context, id int64) (*domain.Article, error)
	updateReadStatusFunc     func(ctx context.Context, id int64, isRead bool) (*domain.Article, error)
	updateFavoriteStatusFunc func(ctx context.Context, id int64, isFavorite bool) (*domain.Article, error)
	bulkUpsertFunc           func(ctx context.Context, sourceID int64, articles []domain.Article) (int, int, error)
}

func (r *stubArticleRepo) List(ctx context.Context, params domain.ListArticlesParams) (domain.ListArticlesResult, error) {
	if r.listFunc != nil {
		return r.listFunc(ctx, params)
	}
	return domain.ListArticlesResult{}, nil
}

func (r *stubArticleRepo) GetByID(ctx context.Context, id int64) (*domain.Article, error) {
	if r.getByIDFunc != nil {
		return r.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (r *stubArticleRepo) UpdateReadStatus(ctx context.Context, id int64, isRead bool) (*domain.Article, error) {
	if r.updateReadStatusFunc != nil {
		return r.updateReadStatusFunc(ctx, id, isRead)
	}
	return nil, nil
}

func (r *stubArticleRepo) UpdateFavoriteStatus(ctx context.Context, id int64, isFavorite bool) (*domain.Article, error) {
	if r.updateFavoriteStatusFunc != nil {
		return r.updateFavoriteStatusFunc(ctx, id, isFavorite)
	}
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

func TestListArticlesPassesReadAndFavoriteFilters(t *testing.T) {
	t.Parallel()

	isRead := false
	isFavorite := true

	service := &ArticleService{
		articleRepo: &stubArticleRepo{
			listFunc: func(_ context.Context, params domain.ListArticlesParams) (domain.ListArticlesResult, error) {
				if params.IsRead == nil || *params.IsRead != isRead {
					t.Fatalf("unexpected is_read: %#v", params.IsRead)
				}
				if params.IsFavorite == nil || *params.IsFavorite != isFavorite {
					t.Fatalf("unexpected is_favorite: %#v", params.IsFavorite)
				}
				return domain.ListArticlesResult{}, nil
			},
		},
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

	_, err := service.ListArticles(context.Background(), domain.ListArticlesParams{
		IsRead:     &isRead,
		IsFavorite: &isFavorite,
	})
	if err != nil {
		t.Fatalf("ListArticles returned error: %v", err)
	}
}

func TestUpdateReadStatusReturnsUpdatedArticle(t *testing.T) {
	t.Parallel()

	service := &ArticleService{
		articleRepo: &stubArticleRepo{
			updateReadStatusFunc: func(_ context.Context, id int64, isRead bool) (*domain.Article, error) {
				if id != 7 || !isRead {
					t.Fatalf("unexpected args: id=%d isRead=%t", id, isRead)
				}
				return &domain.Article{ID: id, IsRead: isRead}, nil
			},
		},
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

	article, err := service.UpdateReadStatus(context.Background(), 7, UpdateReadStatusInput{IsRead: true})
	if err != nil {
		t.Fatalf("UpdateReadStatus returned error: %v", err)
	}
	if article == nil || article.ID != 7 || !article.IsRead {
		t.Fatalf("unexpected article: %+v", article)
	}
}

func TestUpdateFavoriteStatusReturnsNotFoundWhenArticleMissing(t *testing.T) {
	t.Parallel()

	service := &ArticleService{
		articleRepo: &stubArticleRepo{
			updateFavoriteStatusFunc: func(context.Context, int64, bool) (*domain.Article, error) {
				return nil, nil
			},
		},
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

	_, err := service.UpdateFavoriteStatus(context.Background(), 8, UpdateFavoriteStatusInput{IsFavorite: true})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}
