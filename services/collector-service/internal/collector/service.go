package collector

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"tech-feed-hub/collector-service/internal/events"
)

type SourceConfig struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	URL             string `json:"url"`
	DefaultCategory string `json:"default_category"`
}

type rssFeed struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type IngestPayload struct {
	JobID    int64           `json:"job_id"`
	SourceID int64           `json:"source_id"`
	Source   IngestSource    `json:"source"`
	Articles []IngestArticle `json:"articles"`
}

type IngestSource struct {
	Name            string `json:"name"`
	Type            string `json:"type"`
	FetchURL        string `json:"fetch_url"`
	FetchMethod     string `json:"fetch_method"`
	IntervalMinutes int    `json:"interval_minutes"`
	DefaultCategory string `json:"default_category"`
}

type IngestArticle struct {
	Title       string     `json:"title"`
	URL         string     `json:"url"`
	PublishedAt *time.Time `json:"published_at"`
	FetchedAt   time.Time  `json:"fetched_at"`
	Excerpt     string     `json:"excerpt"`
	Category    string     `json:"category"`
	Tags        []string   `json:"tags"`
	DedupeKey   string     `json:"dedupe_key"`
}

type RunResult struct {
	SourceName      string `json:"source_name"`
	SourceID        int64  `json:"source_id"`
	JobID           int64  `json:"job_id"`
	Status          string `json:"status"`
	FetchedCount    int    `json:"fetched_count"`
	InsertedCount   int    `json:"inserted_count"`
	DuplicatedCount int    `json:"duplicated_count"`
	ErrorMessage    string `json:"error_message,omitempty"`
}

type Service struct {
	client            *http.Client
	articleServiceURL string
	sources           []SourceConfig
	publisher         events.EventPublisher
}

func NewService(articleServiceURL string, sources []SourceConfig) *Service {
	return &Service{
		client:            &http.Client{Timeout: 20 * time.Second},
		articleServiceURL: articleServiceURL,
		sources:           sources,
		publisher:         events.NopPublisher{},
	}
}

func (s *Service) SetEventPublisher(publisher events.EventPublisher) {
	if publisher == nil {
		s.publisher = events.NopPublisher{}
		return
	}
	s.publisher = publisher
}

func (s *Service) Run(ctx context.Context) ([]RunResult, error) {
	results := make([]RunResult, 0, len(s.sources))
	for _, source := range s.sources {
		// source ごとに順に処理し、どの source で止まったかを呼び出し側から判断しやすくする。
		result, err := s.collectSource(ctx, source)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *Service) collectSource(ctx context.Context, source SourceConfig) (RunResult, error) {
	// 「開始済みなのに履歴が残らない」を避けるため、外部取得より先に job を起票する。
	started, err := s.startFetchJob(ctx, source)
	if err != nil {
		return RunResult{}, err
	}

	items, err := s.fetchRSS(ctx, source)
	if err != nil {
		return RunResult{}, s.handleFailure(ctx, started, source, err)
	}

	payload := IngestPayload{
		JobID:    started.JobID,
		SourceID: started.SourceID,
		Source: IngestSource{
			Name:            source.Name,
			Type:            source.Type,
			FetchURL:        source.URL,
			FetchMethod:     "rss",
			IntervalMinutes: 60,
			DefaultCategory: source.DefaultCategory,
		},
		Articles: items,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return RunResult{}, s.handleFailure(ctx, started, source, fmt.Errorf("marshal ingest payload: %w", err))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.articleServiceURL+"/internal/ingest", bytes.NewReader(body))
	if err != nil {
		return RunResult{}, s.handleFailure(ctx, started, source, fmt.Errorf("create ingest request: %w", err))
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return RunResult{}, s.handleFailure(ctx, started, source, fmt.Errorf("send ingest request: %w", err))
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		message := fmt.Sprintf("ingest failed: status=%d body=%s", resp.StatusCode, string(raw))
		return RunResult{}, s.handleFailure(ctx, started, source, fmt.Errorf("%s", message))
	}

	var ingestResult struct {
		InsertedCount   int `json:"inserted_count"`
		DuplicatedCount int `json:"duplicated_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ingestResult); err != nil {
		return RunResult{}, s.handleFailure(ctx, started, source, fmt.Errorf("decode ingest response: %w", err))
	}

	// article-service 側で集計した件数を finish に渡し、履歴一覧と source 状態の数字を揃える。
	if err := s.finishFetchJob(ctx, started.JobID, finishFetchJobPayload{
		Status:          "success",
		FetchedCount:    len(items),
		InsertedCount:   ingestResult.InsertedCount,
		DuplicatedCount: ingestResult.DuplicatedCount,
	}); err != nil {
		return RunResult{}, err
	}

	return RunResult{
		SourceName:      source.Name,
		SourceID:        started.SourceID,
		JobID:           started.JobID,
		Status:          "success",
		FetchedCount:    len(items),
		InsertedCount:   ingestResult.InsertedCount,
		DuplicatedCount: ingestResult.DuplicatedCount,
	}, nil
}

type startFetchJobPayload struct {
	Source IngestSource `json:"source"`
}

type startFetchJobResult struct {
	SourceID int64 `json:"source_id"`
	JobID    int64 `json:"job_id"`
}

func (s *Service) startFetchJob(ctx context.Context, source SourceConfig) (startFetchJobResult, error) {
	// source 情報は start API にも渡し、article-service 側で source 解決と job 作成を一貫して行う。
	payload := startFetchJobPayload{
		Source: IngestSource{
			Name:            source.Name,
			Type:            source.Type,
			FetchURL:        source.URL,
			FetchMethod:     "rss",
			IntervalMinutes: 60,
			DefaultCategory: source.DefaultCategory,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return startFetchJobResult{}, fmt.Errorf("marshal fetch job start payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.articleServiceURL+"/internal/fetch-jobs/start", bytes.NewReader(body))
	if err != nil {
		return startFetchJobResult{}, fmt.Errorf("create fetch job start request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return startFetchJobResult{}, fmt.Errorf("send fetch job start request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return startFetchJobResult{}, fmt.Errorf("fetch job start failed: status=%d body=%s", resp.StatusCode, string(raw))
	}

	var result startFetchJobResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return startFetchJobResult{}, fmt.Errorf("decode fetch job start response: %w", err)
	}
	return result, nil
}

type finishFetchJobPayload struct {
	Status          string  `json:"status"`
	FetchedCount    int     `json:"fetched_count"`
	InsertedCount   int     `json:"inserted_count"`
	DuplicatedCount int     `json:"duplicated_count"`
	ErrorMessage    *string `json:"error_message"`
}

func (s *Service) handleFailure(ctx context.Context, started startFetchJobResult, source SourceConfig, cause error) error {
	// 取得失敗も UI から見えることが重要なので、失敗時は必ず finish を打つ。
	if err := s.finishFetchJob(ctx, started.JobID, finishFetchJobPayload{
		Status:       "failed",
		ErrorMessage: stringPtr(cause.Error()),
	}); err != nil {
		return err
	}

	if err := s.publisher.PublishFetchFailed(ctx, events.FetchFailedPayload{
		JobID:        started.JobID,
		SourceID:     started.SourceID,
		SourceName:   source.Name,
		ErrorMessage: cause.Error(),
	}); err != nil {
		log.Printf("publish collector.fetch.failed event: %v", err)
	}

	return cause
}

func (s *Service) finishFetchJob(ctx context.Context, jobID int64, payload finishFetchJobPayload) error {
	// finish API を単一の出口にして、collector 側の失敗種別に関係なく同じ履歴更新経路へ流す。
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal fetch job finish payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/internal/fetch-jobs/%d/finish", s.articleServiceURL, jobID), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create fetch job finish request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("send fetch job finish request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("fetch job finish failed: status=%d body=%s", resp.StatusCode, string(raw))
	}
	return nil
}

func (s *Service) fetchRSS(ctx context.Context, source SourceConfig) ([]IngestArticle, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, source.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create rss request: %w", err)
	}
	req.Header.Set("User-Agent", "tech-feed-hub/0.1")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch rss: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch rss status=%d", resp.StatusCode)
	}

	var feed rssFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("decode rss: %w", err)
	}

	now := time.Now().UTC()
	articles := make([]IngestArticle, 0, len(feed.Channel.Items))
	for _, item := range feed.Channel.Items {
		article := IngestArticle{
			Title:     strings.TrimSpace(item.Title),
			URL:       strings.TrimSpace(item.Link),
			FetchedAt: now,
			Excerpt:   strings.TrimSpace(stripHTML(item.Description)),
			Category:  source.DefaultCategory,
			Tags:      []string{source.DefaultCategory},
			DedupeKey: dedupeKey(item.Link),
		}
		if article.Title == "" || article.URL == "" {
			continue
		}

		if publishedAt, err := parsePubDate(item.PubDate); err == nil {
			article.PublishedAt = &publishedAt
		}
		articles = append(articles, article)
	}

	return articles, nil
}

func parsePubDate(raw string) (time.Time, error) {
	for _, layout := range []string{time.RFC1123Z, time.RFC1123, time.RFC3339} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported pubDate: %s", raw)
}

func dedupeKey(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])
}

func stripHTML(value string) string {
	replacer := strings.NewReplacer(
		"<p>", " ",
		"</p>", " ",
		"<br>", " ",
		"<br/>", " ",
		"<br />", " ",
		"&nbsp;", " ",
		"&amp;", "&",
	)
	cleaned := replacer.Replace(value)
	cleaned = strings.NewReplacer("<", " ", ">", " ").Replace(cleaned)
	return strings.Join(strings.Fields(cleaned), " ")
}

func stringPtr(value string) *string {
	return &value
}
