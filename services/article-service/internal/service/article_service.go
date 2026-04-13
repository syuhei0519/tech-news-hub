package service

import (
	"context"
	"fmt"
	"time"

	"tech-feed-hub/article-service/internal/domain"
	"tech-feed-hub/article-service/internal/repository"
)

type ArticleService struct {
	articleRepo *repository.ArticleRepository
	sourceRepo  *repository.SourceRepository
	jobRepo     *repository.FetchJobRepository
}

func NewArticleService(articleRepo *repository.ArticleRepository, sourceRepo *repository.SourceRepository, jobRepo *repository.FetchJobRepository) *ArticleService {
	return &ArticleService{
		articleRepo: articleRepo,
		sourceRepo:  sourceRepo,
		jobRepo:     jobRepo,
	}
}

func (s *ArticleService) ListArticles(ctx context.Context, params domain.ListArticlesParams) (domain.ListArticlesResult, error) {
	return s.articleRepo.List(ctx, params)
}

func (s *ArticleService) GetArticle(ctx context.Context, id int64) (*domain.Article, error) {
	return s.articleRepo.GetByID(ctx, id)
}

type IngestSourceInput struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	FetchURL        string `json:"fetch_url"`
	FetchMethod     string `json:"fetch_method"`
	IntervalMinutes int    `json:"interval_minutes"`
	DefaultCategory string `json:"default_category"`
}

type IngestArticleInput struct {
	Title       string     `json:"title"`
	URL         string     `json:"url"`
	PublishedAt *time.Time `json:"published_at"`
	FetchedAt   time.Time  `json:"fetched_at"`
	Excerpt     string     `json:"excerpt"`
	Category    string     `json:"category"`
	Tags        []string   `json:"tags"`
	DedupeKey   string     `json:"dedupe_key"`
}

type IngestRequest struct {
	Source   IngestSourceInput    `json:"source"`
	Articles []IngestArticleInput `json:"articles"`
}

type IngestResult struct {
	SourceID        int64 `json:"source_id"`
	JobID           int64 `json:"job_id"`
	FetchedCount    int   `json:"fetched_count"`
	InsertedCount   int   `json:"inserted_count"`
	DuplicatedCount int   `json:"duplicated_count"`
}

func (s *ArticleService) Ingest(ctx context.Context, req IngestRequest) (IngestResult, error) {
	sourceID, err := s.sourceRepo.EnsureSource(ctx, domain.Source{
		Name:            req.Source.Name,
		Type:            req.Source.Type,
		FetchURL:        req.Source.FetchURL,
		FetchMethod:     req.Source.FetchMethod,
		IntervalMinutes: req.Source.IntervalMinutes,
		DefaultCategory: req.Source.DefaultCategory,
	})
	if err != nil {
		return IngestResult{}, err
	}

	jobID, err := s.jobRepo.Create(ctx, sourceID)
	if err != nil {
		return IngestResult{}, err
	}

	articles := make([]domain.Article, 0, len(req.Articles))
	for _, item := range req.Articles {
		category := item.Category
		if category == "" {
			category = req.Source.DefaultCategory
		}
		articles = append(articles, domain.Article{
			Title:       item.Title,
			URL:         item.URL,
			PublishedAt: item.PublishedAt,
			FetchedAt:   item.FetchedAt,
			Excerpt:     item.Excerpt,
			Category:    category,
			Tags:        item.Tags,
			DedupeKey:   item.DedupeKey,
		})
	}

	inserted, duplicated, ingestErr := s.articleRepo.BulkUpsert(ctx, sourceID, articles)
	if ingestErr != nil {
		errMsg := ingestErr.Error()
		_ = s.jobRepo.Finish(ctx, jobID, "failed", len(articles), inserted, duplicated, &errMsg)
		_ = s.sourceRepo.UpdateFetchStatus(ctx, sourceID, "failed", &errMsg)
		return IngestResult{}, fmt.Errorf("bulk upsert: %w", ingestErr)
	}

	if err := s.jobRepo.Finish(ctx, jobID, "success", len(articles), inserted, duplicated, nil); err != nil {
		return IngestResult{}, err
	}
	if err := s.sourceRepo.UpdateFetchStatus(ctx, sourceID, "success", nil); err != nil {
		return IngestResult{}, err
	}

	return IngestResult{
		SourceID:        sourceID,
		JobID:           jobID,
		FetchedCount:    len(articles),
		InsertedCount:   inserted,
		DuplicatedCount: duplicated,
	}, nil
}
