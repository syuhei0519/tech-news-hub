package service

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"tech-feed-hub/article-service/internal/domain"
	"tech-feed-hub/article-service/internal/repository"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
)

type serviceError struct {
	kind    error
	message string
}

func (e *serviceError) Error() string {
	return e.message
}

func (e *serviceError) Unwrap() error {
	return e.kind
}

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

type SourceInput struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	FetchURL        string `json:"fetch_url"`
	FetchMethod     string `json:"fetch_method"`
	IntervalMinutes int    `json:"interval_minutes"`
	DefaultCategory string `json:"default_category"`
	IsEnabled       bool   `json:"is_enabled"`
}

func (s *ArticleService) ListSources(ctx context.Context) ([]domain.Source, error) {
	return s.sourceRepo.List(ctx)
}

func (s *ArticleService) GetSource(ctx context.Context, id int64) (*domain.Source, error) {
	source, err := s.sourceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if source == nil {
		return nil, newServiceError(ErrNotFound, "source not found")
	}
	return source, nil
}

func (s *ArticleService) CreateSource(ctx context.Context, input SourceInput) (*domain.Source, error) {
	source, err := sanitizeSourceInput(input)
	if err != nil {
		return nil, err
	}
	created, err := s.sourceRepo.Create(ctx, source)
	if err != nil {
		return nil, mapSourceError(err)
	}
	return created, nil
}

func (s *ArticleService) UpdateSource(ctx context.Context, id int64, input SourceInput) (*domain.Source, error) {
	source, err := sanitizeSourceInput(input)
	if err != nil {
		return nil, err
	}
	source.ID = id
	updated, err := s.sourceRepo.Update(ctx, source)
	if err != nil {
		return nil, mapSourceError(err)
	}
	if updated == nil {
		return nil, newServiceError(ErrNotFound, "source not found")
	}
	return updated, nil
}

func (s *ArticleService) DeleteSource(ctx context.Context, id int64) error {
	deleted, err := s.sourceRepo.Delete(ctx, id)
	if err != nil {
		return mapSourceError(err)
	}
	if !deleted {
		return newServiceError(ErrNotFound, "source not found")
	}
	return nil
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

func sanitizeSourceInput(input SourceInput) (domain.Source, error) {
	source := domain.Source{
		Name:            strings.TrimSpace(input.Name),
		Type:            strings.TrimSpace(strings.ToLower(input.Type)),
		FetchURL:        strings.TrimSpace(input.FetchURL),
		FetchMethod:     strings.TrimSpace(strings.ToLower(input.FetchMethod)),
		IntervalMinutes: input.IntervalMinutes,
		DefaultCategory: strings.TrimSpace(input.DefaultCategory),
		IsEnabled:       input.IsEnabled,
	}

	switch {
	case source.Name == "":
		return domain.Source{}, newServiceError(ErrValidation, "name is required")
	case len(source.Name) > 255:
		return domain.Source{}, newServiceError(ErrValidation, "name must be 255 characters or fewer")
	case source.Type == "":
		return domain.Source{}, newServiceError(ErrValidation, "type is required")
	case source.Type != "rss":
		return domain.Source{}, newServiceError(ErrValidation, "type must be rss")
	case source.FetchURL == "":
		return domain.Source{}, newServiceError(ErrValidation, "fetch_url is required")
	case len(source.FetchURL) > 2048:
		return domain.Source{}, newServiceError(ErrValidation, "fetch_url must be 2048 characters or fewer")
	case source.FetchMethod == "":
		return domain.Source{}, newServiceError(ErrValidation, "fetch_method is required")
	case source.FetchMethod != "rss":
		return domain.Source{}, newServiceError(ErrValidation, "fetch_method must be rss")
	case source.IntervalMinutes < 1:
		return domain.Source{}, newServiceError(ErrValidation, "interval_minutes must be greater than or equal to 1")
	case source.IntervalMinutes > 10080:
		return domain.Source{}, newServiceError(ErrValidation, "interval_minutes must be less than or equal to 10080")
	case source.DefaultCategory == "":
		return domain.Source{}, newServiceError(ErrValidation, "default_category is required")
	case len(source.DefaultCategory) > 128:
		return domain.Source{}, newServiceError(ErrValidation, "default_category must be 128 characters or fewer")
	}

	parsedURL, err := url.ParseRequestURI(source.FetchURL)
	if err != nil {
		return domain.Source{}, newServiceError(ErrValidation, "fetch_url must be a valid absolute URL")
	}
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return domain.Source{}, newServiceError(ErrValidation, "fetch_url must use http or https")
	}

	return source, nil
}

func mapSourceError(err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, "duplicate"):
		return newServiceError(ErrConflict, "source already exists")
	case strings.Contains(message, "referenced"):
		return newServiceError(ErrConflict, "source is still referenced by related records")
	default:
		return err
	}
}

func newServiceError(kind error, message string) error {
	return &serviceError{kind: kind, message: message}
}
