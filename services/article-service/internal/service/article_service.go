package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/url"
	"strings"
	"time"

	"tech-feed-hub/article-service/internal/domain"
	"tech-feed-hub/article-service/internal/events"
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
	articleRepo articleRepository
	sourceRepo  sourceRepository
	jobRepo     fetchJobRepository
	publisher   events.NotificationPublisher
}

type articleRepository interface {
	List(ctx context.Context, params domain.ListArticlesParams) (domain.ListArticlesResult, error)
	Export(ctx context.Context, params domain.ExportArticlesParams) ([]domain.Article, error)
	GetByID(ctx context.Context, id int64) (*domain.Article, error)
	UpdateReadStatus(ctx context.Context, id int64, isRead bool) (*domain.Article, error)
	UpdateFavoriteStatus(ctx context.Context, id int64, isFavorite bool) (*domain.Article, error)
	BulkUpsert(ctx context.Context, sourceID int64, articles []domain.Article) (inserted int, duplicated int, err error)
}

type sourceRepository interface {
	List(ctx context.Context) ([]domain.Source, error)
	GetByID(ctx context.Context, id int64) (*domain.Source, error)
	Create(ctx context.Context, source domain.Source) (*domain.Source, error)
	Update(ctx context.Context, source domain.Source) (*domain.Source, error)
	Delete(ctx context.Context, id int64) (bool, error)
	EnsureSource(ctx context.Context, source domain.Source) (int64, error)
	UpdateFetchStatus(ctx context.Context, id int64, status string, errMsg *string) error
}

type fetchJobRepository interface {
	Create(ctx context.Context, sourceID int64) (int64, error)
	GetByID(ctx context.Context, id int64) (*domain.FetchJob, error)
	List(ctx context.Context, params domain.ListFetchJobsParams) (domain.ListFetchJobsResult, error)
	Finish(ctx context.Context, jobID int64, status string, fetchedCount int, insertedCount int, duplicatedCount int, errMsg *string) error
}

func NewArticleService(articleRepo *repository.ArticleRepository, sourceRepo *repository.SourceRepository, jobRepo *repository.FetchJobRepository) *ArticleService {
	return &ArticleService{
		articleRepo: articleRepo,
		sourceRepo:  sourceRepo,
		jobRepo:     jobRepo,
		publisher:   events.NopPublisher{},
	}
}

func (s *ArticleService) SetNotificationPublisher(publisher events.NotificationPublisher) {
	if publisher == nil {
		s.publisher = events.NopPublisher{}
		return
	}
	s.publisher = publisher
}

func (s *ArticleService) ListFetchJobs(ctx context.Context, params domain.ListFetchJobsParams) (domain.ListFetchJobsResult, error) {
	// 履歴一覧は source 詳細画面から使う前提のため、source 単位での絞り込みを必須にする。
	if params.SourceID < 1 {
		return domain.ListFetchJobsResult{}, newServiceError(ErrValidation, "source_id is required")
	}
	if params.Status != "" && params.Status != "running" && params.Status != "success" && params.Status != "failed" {
		return domain.ListFetchJobsResult{}, newServiceError(ErrValidation, "status must be running, success, or failed")
	}
	return s.jobRepo.List(ctx, params)
}

func (s *ArticleService) ListArticles(ctx context.Context, params domain.ListArticlesParams) (domain.ListArticlesResult, error) {
	if err := validateArticleFilters(params.ArticleFilterParams); err != nil {
		return domain.ListArticlesResult{}, err
	}
	return s.articleRepo.List(ctx, params)
}

func (s *ArticleService) ExportArticlesCSV(ctx context.Context, params domain.ExportArticlesParams) ([]byte, error) {
	if err := validateArticleFilters(params.ArticleFilterParams); err != nil {
		return nil, err
	}
	if params.Limit < 1 {
		params.Limit = 1000
	}

	articles, err := s.articleRepo.Export(ctx, params)
	if err != nil {
		return nil, err
	}
	if len(articles) > params.Limit {
		return nil, newServiceError(ErrValidation, fmt.Sprintf("export result exceeds limit %d", params.Limit))
	}

	var buf bytes.Buffer
	buf.Write([]byte{0xEF, 0xBB, 0xBF})
	writer := csv.NewWriter(&buf)
	if err := writer.Write([]string{
		"title", "url", "source_name", "category", "published_at",
		"fetched_at", "is_read", "is_favorite", "excerpt", "tags",
	}); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	for _, article := range articles {
		record := []string{
			sanitizeCSVCell(article.Title),
			sanitizeCSVCell(article.URL),
			sanitizeCSVCell(article.SourceName),
			sanitizeCSVCell(article.Category),
			formatCSVTime(article.PublishedAt),
			formatCSVTime(&article.FetchedAt),
			formatCSVBool(article.IsRead),
			formatCSVBool(article.IsFavorite),
			sanitizeCSVCell(article.Excerpt),
			sanitizeCSVCell(strings.Join(article.Tags, ";")),
		}
		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("write csv record: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}
	return buf.Bytes(), nil
}

func (s *ArticleService) GetArticle(ctx context.Context, id int64) (*domain.Article, error) {
	return s.articleRepo.GetByID(ctx, id)
}

type UpdateReadStatusInput struct {
	IsRead bool `json:"is_read"`
}

type UpdateFavoriteStatusInput struct {
	IsFavorite bool `json:"is_favorite"`
}

func (s *ArticleService) UpdateReadStatus(ctx context.Context, id int64, input UpdateReadStatusInput) (*domain.Article, error) {
	if id < 1 {
		return nil, newServiceError(ErrValidation, "article id is required")
	}

	// 初期は記事テーブル内の単純 state 更新で閉じ、ユーザー別の抽象化は持ち込まない。
	article, err := s.articleRepo.UpdateReadStatus(ctx, id, input.IsRead)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, newServiceError(ErrNotFound, "article not found")
	}
	return article, nil
}

func (s *ArticleService) UpdateFavoriteStatus(ctx context.Context, id int64, input UpdateFavoriteStatusInput) (*domain.Article, error) {
	if id < 1 {
		return nil, newServiceError(ErrValidation, "article id is required")
	}

	// favorite も article-service の責務内で完結させ、gateway 側に状態判断を持たせない。
	article, err := s.articleRepo.UpdateFavoriteStatus(ctx, id, input.IsFavorite)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, newServiceError(ErrNotFound, "article not found")
	}
	return article, nil
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
	// 更新時は handler 側で path id を受けるため、payload からは受けずここで上書きする。
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
	JobID    int64                `json:"job_id"`
	SourceID int64                `json:"source_id"`
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
	if req.JobID < 1 {
		return IngestResult{}, newServiceError(ErrValidation, "job_id is required")
	}
	if req.SourceID < 1 {
		return IngestResult{}, newServiceError(ErrValidation, "source_id is required")
	}

	// job の作成と完了は collector の前後処理に寄せ、ingest 自体は記事登録だけに責務を絞る。
	job, err := s.jobRepo.GetByID(ctx, req.JobID)
	if err != nil {
		return IngestResult{}, err
	}
	if job == nil {
		return IngestResult{}, newServiceError(ErrNotFound, "fetch job not found")
	}
	if job.SourceID != req.SourceID {
		return IngestResult{}, newServiceError(ErrValidation, "job source mismatch")
	}
	if job.Status != "running" {
		return IngestResult{}, newServiceError(ErrConflict, "fetch job is already finished")
	}

	articles := make([]domain.Article, 0, len(req.Articles))
	for _, item := range req.Articles {
		category := item.Category
		// source ごとの既定カテゴリで最低限の分類を維持する。
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

	inserted, duplicated, ingestErr := s.articleRepo.BulkUpsert(ctx, req.SourceID, articles)
	if ingestErr != nil {
		return IngestResult{}, fmt.Errorf("bulk upsert: %w", ingestErr)
	}

	result := IngestResult{
		SourceID:        req.SourceID,
		JobID:           req.JobID,
		FetchedCount:    len(articles),
		InsertedCount:   inserted,
		DuplicatedCount: duplicated,
	}

	if inserted > 0 {
		var representativeTitle *string
		for _, article := range articles {
			if strings.TrimSpace(article.Title) == "" {
				continue
			}
			title := article.Title
			representativeTitle = &title
			break
		}
		if err := s.publisher.PublishArticleIngested(ctx, events.ArticleIngestedPayload{
			JobID:               req.JobID,
			SourceID:            req.SourceID,
			SourceName:          req.Source.Name,
			InsertedCount:       inserted,
			RepresentativeTitle: representativeTitle,
		}); err != nil {
			log.Printf("publish article.ingested event: %v", err)
		}
	}

	return result, nil
}

type StartFetchJobInput struct {
	SourceID int64             `json:"source_id"`
	Source   IngestSourceInput `json:"source"`
}

type StartFetchJobResult struct {
	SourceID int64 `json:"source_id"`
	JobID    int64 `json:"job_id"`
}

func (s *ArticleService) StartFetchJob(ctx context.Context, input StartFetchJobInput) (StartFetchJobResult, error) {
	var sourceID int64
	if input.SourceID > 0 {
		// source 一覧 API で解決済みの ID が来た場合は、それを job 作成にそのまま使う。
		source, err := s.sourceRepo.GetByID(ctx, input.SourceID)
		if err != nil {
			return StartFetchJobResult{}, err
		}
		if source == nil {
			return StartFetchJobResult{}, newServiceError(ErrNotFound, "source not found")
		}
		sourceID = source.ID
	} else {
		// 後方互換のため、source 情報だけの caller も引き続き受け付ける。
		if strings.TrimSpace(input.Source.Name) == "" {
			return StartFetchJobResult{}, newServiceError(ErrValidation, "source_id or source is required")
		}

		var err error
		sourceID, err = s.sourceRepo.EnsureSource(ctx, domain.Source{
			Name:            input.Source.Name,
			Type:            input.Source.Type,
			FetchURL:        input.Source.FetchURL,
			FetchMethod:     input.Source.FetchMethod,
			IntervalMinutes: input.Source.IntervalMinutes,
			DefaultCategory: input.Source.DefaultCategory,
		})
		if err != nil {
			return StartFetchJobResult{}, err
		}
	}

	jobID, err := s.jobRepo.Create(ctx, sourceID)
	if err != nil {
		return StartFetchJobResult{}, err
	}

	return StartFetchJobResult{
		SourceID: sourceID,
		JobID:    jobID,
	}, nil
}

type FinishFetchJobInput struct {
	Status          string  `json:"status"`
	FetchedCount    int     `json:"fetched_count"`
	InsertedCount   int     `json:"inserted_count"`
	DuplicatedCount int     `json:"duplicated_count"`
	ErrorMessage    *string `json:"error_message"`
}

func (s *ArticleService) FinishFetchJob(ctx context.Context, jobID int64, input FinishFetchJobInput) error {
	if jobID < 1 {
		return newServiceError(ErrValidation, "job_id is required")
	}
	if input.Status != "success" && input.Status != "failed" {
		return newServiceError(ErrValidation, "status must be success or failed")
	}
	if input.FetchedCount < 0 || input.InsertedCount < 0 || input.DuplicatedCount < 0 {
		return newServiceError(ErrValidation, "counts must be greater than or equal to 0")
	}

	job, err := s.jobRepo.GetByID(ctx, jobID)
	if err != nil {
		return err
	}
	if job == nil {
		return newServiceError(ErrNotFound, "fetch job not found")
	}
	if job.Status != "running" {
		return newServiceError(ErrConflict, "fetch job is already finished")
	}

	// source 一覧と詳細で最後の取得状態を即座に出せるよう、job 完了と source 状態更新を同じ service に集約する。
	errorMessage := normalizeOptionalString(input.ErrorMessage)
	if input.Status == "failed" && errorMessage == nil {
		return newServiceError(ErrValidation, "error_message is required when status is failed")
	}
	if input.Status == "success" {
		errorMessage = nil
	}

	if err := s.jobRepo.Finish(ctx, jobID, input.Status, input.FetchedCount, input.InsertedCount, input.DuplicatedCount, errorMessage); err != nil {
		return err
	}
	if err := s.sourceRepo.UpdateFetchStatus(ctx, job.SourceID, input.Status, errorMessage); err != nil {
		return err
	}
	return nil
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
	// 現行 collector の取得経路に合わせ、http/https だけを受け付ける。
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
		// article / fetch_jobs から参照される source は即時削除できない。
		return newServiceError(ErrConflict, "source is still referenced by related records")
	default:
		return err
	}
}

func newServiceError(kind error, message string) error {
	return &serviceError{kind: kind, message: message}
}

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func validateArticleFilters(params domain.ArticleFilterParams) error {
	if params.PublishedFrom != nil && params.PublishedTo != nil && params.PublishedFrom.After(*params.PublishedTo) {
		return newServiceError(ErrValidation, "from must be less than or equal to to")
	}
	return nil
}

func formatCSVTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func formatCSVBool(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func sanitizeCSVCell(value string) string {
	if value == "" {
		return ""
	}
	switch value[0] {
	case '=', '+', '-', '@', '\t', '\r', '\n':
		return "'" + value
	default:
		return value
	}
}
