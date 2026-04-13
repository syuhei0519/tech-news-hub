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
	"net/http"
	"strings"
	"time"
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
	FetchedCount    int    `json:"fetched_count"`
	InsertedCount   int    `json:"inserted_count"`
	DuplicatedCount int    `json:"duplicated_count"`
}

type Service struct {
	client            *http.Client
	articleServiceURL string
	sources           []SourceConfig
}

func NewService(articleServiceURL string, sources []SourceConfig) *Service {
	return &Service{
		client:            &http.Client{Timeout: 20 * time.Second},
		articleServiceURL: articleServiceURL,
		sources:           sources,
	}
}

func (s *Service) Run(ctx context.Context) ([]RunResult, error) {
	results := make([]RunResult, 0, len(s.sources))
	for _, source := range s.sources {
		result, err := s.collectSource(ctx, source)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *Service) collectSource(ctx context.Context, source SourceConfig) (RunResult, error) {
	items, err := s.fetchRSS(ctx, source)
	if err != nil {
		return RunResult{}, err
	}

	payload := IngestPayload{
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
		return RunResult{}, fmt.Errorf("marshal ingest payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.articleServiceURL+"/internal/ingest", bytes.NewReader(body))
	if err != nil {
		return RunResult{}, fmt.Errorf("create ingest request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return RunResult{}, fmt.Errorf("send ingest request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return RunResult{}, fmt.Errorf("ingest failed: status=%d body=%s", resp.StatusCode, string(raw))
	}

	var ingestResult struct {
		InsertedCount   int `json:"inserted_count"`
		DuplicatedCount int `json:"duplicated_count"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ingestResult); err != nil {
		return RunResult{}, fmt.Errorf("decode ingest response: %w", err)
	}

	return RunResult{
		SourceName:      source.Name,
		FetchedCount:    len(items),
		InsertedCount:   ingestResult.InsertedCount,
		DuplicatedCount: ingestResult.DuplicatedCount,
	}, nil
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
