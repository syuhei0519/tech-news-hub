package collector

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"tech-feed-hub/collector-service/internal/events"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newJSONResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func newXMLResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Header:     http.Header{"Content-Type": []string{"application/xml"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

type stubEventPublisher struct {
	publishFetchFailedFunc func(context.Context, events.FetchFailedPayload) error
}

func (p *stubEventPublisher) PublishFetchFailed(ctx context.Context, payload events.FetchFailedPayload) error {
	if p.publishFetchFailedFunc != nil {
		return p.publishFetchFailedFunc(ctx, payload)
	}
	return nil
}

func TestRunCreatesAndFinishesFetchJobOnSuccess(t *testing.T) {
	t.Parallel()

	var startPayload startFetchJobPayload
	var finishPayload finishFetchJobPayload
	var ingestPayload IngestPayload

	service := NewService("http://article-service")
	service.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/api/v1/sources":
			return newJSONResponse(http.StatusOK, `{"items":[{"id":7,"name":"Example","type":"rss","fetch_url":"http://feed-source/rss","fetch_method":"rss","interval_minutes":30,"default_category":"cloud","is_enabled":true}]}`), nil
		case "/rss":
			return newXMLResponse(http.StatusOK, `<?xml version="1.0"?><rss><channel><item><title>Hello</title><link>https://example.com/1</link><description>desc</description><pubDate>Mon, 02 Jan 2006 15:04:05 +0900</pubDate></item></channel></rss>`), nil
		case "/internal/fetch-jobs/start":
			if err := json.NewDecoder(r.Body).Decode(&startPayload); err != nil {
				t.Fatalf("decode start payload: %v", err)
			}
			return newJSONResponse(http.StatusAccepted, `{"source_id":7,"job_id":11}`), nil
		case "/internal/ingest":
			if err := json.NewDecoder(r.Body).Decode(&ingestPayload); err != nil {
				t.Fatalf("decode ingest payload: %v", err)
			}
			return newJSONResponse(http.StatusAccepted, `{"source_id":7,"job_id":11,"fetched_count":1,"inserted_count":1,"duplicated_count":0}`), nil
		case "/internal/fetch-jobs/11/finish":
			if err := json.NewDecoder(r.Body).Decode(&finishPayload); err != nil {
				t.Fatalf("decode finish payload: %v", err)
			}
			return newJSONResponse(http.StatusNoContent, ``), nil
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return nil, nil
		}
	})}

	results, err := service.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if startPayload.SourceID != 7 || startPayload.Source.IntervalMinutes != 30 || startPayload.Source.FetchMethod != "rss" {
		t.Fatalf("unexpected start payload: %+v", startPayload)
	}
	if ingestPayload.JobID != 11 || ingestPayload.SourceID != 7 {
		t.Fatalf("unexpected ingest payload ids: %+v", ingestPayload)
	}
	if ingestPayload.Source.FetchURL != "http://feed-source/rss" || ingestPayload.Source.IntervalMinutes != 30 {
		t.Fatalf("unexpected ingest source payload: %+v", ingestPayload.Source)
	}
	if finishPayload.Status != "success" || finishPayload.FetchedCount != 1 || finishPayload.InsertedCount != 1 || finishPayload.DuplicatedCount != 0 || finishPayload.ErrorMessage != nil {
		t.Fatalf("unexpected finish payload: %+v", finishPayload)
	}
	if len(results) != 1 || results[0].Status != "success" || results[0].JobID != 11 || results[0].SourceID != 7 {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestRunSendsNormalizedIngestContract(t *testing.T) {
	t.Parallel()

	var ingestPayload IngestPayload

	service := NewService("http://article-service")
	service.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/api/v1/sources":
			return newJSONResponse(http.StatusOK, `{"items":[{"id":7,"name":"Example","type":"rss","fetch_url":"http://feed-source/rss","fetch_method":"rss","interval_minutes":30,"default_category":"cloud","is_enabled":true}]}`), nil
		case "/rss":
			return newXMLResponse(http.StatusOK, `<?xml version="1.0"?><rss><channel>`+
				`<item><title>RFC1123Z</title><link>https://example.com/with-date</link><description><![CDATA[<p>desc</p>]]></description><pubDate>Mon, 02 Jan 2006 15:04:05 +0900</pubDate></item>`+
				`<item><title>Missing Date</title><link>https://example.com/no-date</link><description>plain text</description></item>`+
				`</channel></rss>`), nil
		case "/internal/fetch-jobs/start":
			return newJSONResponse(http.StatusAccepted, `{"source_id":7,"job_id":11}`), nil
		case "/internal/ingest":
			if err := json.NewDecoder(r.Body).Decode(&ingestPayload); err != nil {
				t.Fatalf("decode ingest payload: %v", err)
			}
			return newJSONResponse(http.StatusAccepted, `{"source_id":7,"job_id":11,"fetched_count":2,"inserted_count":1,"duplicated_count":1}`), nil
		case "/internal/fetch-jobs/11/finish":
			return newJSONResponse(http.StatusNoContent, ``), nil
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return nil, nil
		}
	})}

	if _, err := service.Run(context.Background()); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if ingestPayload.JobID != 11 || ingestPayload.SourceID != 7 {
		t.Fatalf("unexpected ingest payload ids: %+v", ingestPayload)
	}
	if ingestPayload.Source.Name != "Example" || ingestPayload.Source.DefaultCategory != "cloud" || ingestPayload.Source.FetchMethod != "rss" {
		t.Fatalf("unexpected ingest source contract: %+v", ingestPayload.Source)
	}
	if len(ingestPayload.Articles) != 2 {
		t.Fatalf("expected 2 normalized articles, got %+v", ingestPayload.Articles)
	}

	first := ingestPayload.Articles[0]
	if first.Title != "RFC1123Z" || first.URL != "https://example.com/with-date" || first.Excerpt != "desc" || first.Category != "cloud" {
		t.Fatalf("unexpected first normalized article: %+v", first)
	}
	if first.PublishedAt == nil || first.PublishedAt.Format(time.RFC3339) != "2006-01-02T06:04:05Z" {
		t.Fatalf("expected normalized published_at, got %+v", first.PublishedAt)
	}
	if len(first.Tags) != 1 || first.Tags[0] != "cloud" {
		t.Fatalf("unexpected first tags: %+v", first.Tags)
	}
	if first.DedupeKey != sha256Hex("https://example.com/with-date") {
		t.Fatalf("unexpected dedupe_key: %s", first.DedupeKey)
	}

	second := ingestPayload.Articles[1]
	if second.Title != "Missing Date" || second.PublishedAt != nil {
		t.Fatalf("expected nil published_at for missing date, got %+v", second)
	}
	if second.DedupeKey != sha256Hex("https://example.com/no-date") {
		t.Fatalf("unexpected second dedupe_key: %s", second.DedupeKey)
	}
}

func TestRunSkipsDisabledSources(t *testing.T) {
	t.Parallel()

	service := NewService("http://article-service")
	service.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/v1/sources" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		return newJSONResponse(http.StatusOK, `{"items":[{"id":7,"name":"Example","type":"rss","fetch_url":"http://feed-source/rss","fetch_method":"rss","interval_minutes":30,"default_category":"cloud","is_enabled":false}]}`), nil
	})}

	results, err := service.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected no results, got %+v", results)
	}
}

func TestRunReturnsErrorWhenSourceSyncFails(t *testing.T) {
	t.Parallel()

	service := NewService("http://article-service")
	service.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path != "/api/v1/sources" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		return newJSONResponse(http.StatusBadGateway, `{"error":"upstream down"}`), nil
	})}

	_, err := service.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "source sync failed: status=502") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunReturnsErrorWhenSourceSyncContainsInvalidSource(t *testing.T) {
	t.Parallel()

	// source の真実源は article-service 側なので、collector では不正な同期結果を即座に弾く。
	tests := []struct {
		name    string
		body    string
		wantErr string
	}{
		{
			name:    "invalid id",
			body:    `{"items":[{"id":0,"name":"Example","type":"rss","fetch_url":"http://feed-source/rss","fetch_method":"rss","interval_minutes":30,"default_category":"cloud","is_enabled":true}]}`,
			wantErr: "invalid id",
		},
		{
			name:    "unsupported type",
			body:    `{"items":[{"id":7,"name":"Example","type":"html","fetch_url":"http://feed-source/rss","fetch_method":"rss","interval_minutes":30,"default_category":"cloud","is_enabled":true}]}`,
			wantErr: "unsupported type",
		},
		{
			name:    "missing fetch url",
			body:    `{"items":[{"id":7,"name":"Example","type":"rss","fetch_url":"","fetch_method":"rss","interval_minutes":30,"default_category":"cloud","is_enabled":true}]}`,
			wantErr: "empty fetch_url",
		},
		{
			name:    "invalid interval",
			body:    `{"items":[{"id":7,"name":"Example","type":"rss","fetch_url":"http://feed-source/rss","fetch_method":"rss","interval_minutes":0,"default_category":"cloud","is_enabled":true}]}`,
			wantErr: "invalid interval_minutes",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			service := NewService("http://article-service")
			service.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path != "/api/v1/sources" {
					t.Fatalf("unexpected path: %s", r.URL.Path)
				}
				return newJSONResponse(http.StatusOK, tt.body), nil
			})}

			_, err := service.Run(context.Background())
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestRunFinishesFetchJobOnRSSFailure(t *testing.T) {
	t.Parallel()

	var finishPayload finishFetchJobPayload
	var publishedPayload events.FetchFailedPayload

	service := NewService("http://article-service")
	service.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/api/v1/sources":
			return newJSONResponse(http.StatusOK, `{"items":[{"id":7,"name":"Example","type":"rss","fetch_url":"http://feed-source/rss","fetch_method":"rss","interval_minutes":60,"default_category":"cloud","is_enabled":true}]}`), nil
		case "/rss":
			return newJSONResponse(http.StatusBadGateway, "boom\n"), nil
		case "/internal/fetch-jobs/start":
			return newJSONResponse(http.StatusAccepted, `{"source_id":7,"job_id":11}`), nil
		case "/internal/fetch-jobs/11/finish":
			if err := json.NewDecoder(r.Body).Decode(&finishPayload); err != nil {
				t.Fatalf("decode finish payload: %v", err)
			}
			return newJSONResponse(http.StatusNoContent, ``), nil
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return nil, nil
		}
	})}
	service.SetEventPublisher(&stubEventPublisher{
		publishFetchFailedFunc: func(_ context.Context, payload events.FetchFailedPayload) error {
			publishedPayload = payload
			return nil
		},
	})

	_, err := service.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "fetch rss status=502") {
		t.Fatalf("unexpected error: %v", err)
	}
	if finishPayload.Status != "failed" || finishPayload.ErrorMessage == nil || !strings.Contains(*finishPayload.ErrorMessage, "fetch rss status=502") {
		t.Fatalf("unexpected finish payload: %+v", finishPayload)
	}
	if publishedPayload.JobID != 11 || publishedPayload.SourceID != 7 || publishedPayload.SourceName != "Example" {
		t.Fatalf("unexpected published payload: %+v", publishedPayload)
	}
}

func TestRunIgnoresPublisherFailureOnRSSFailure(t *testing.T) {
	t.Parallel()

	service := NewService("http://article-service")
	service.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/api/v1/sources":
			return newJSONResponse(http.StatusOK, `{"items":[{"id":7,"name":"Example","type":"rss","fetch_url":"http://feed-source/rss","fetch_method":"rss","interval_minutes":60,"default_category":"cloud","is_enabled":true}]}`), nil
		case "/rss":
			return newJSONResponse(http.StatusBadGateway, "boom\n"), nil
		case "/internal/fetch-jobs/start":
			return newJSONResponse(http.StatusAccepted, `{"source_id":7,"job_id":11}`), nil
		case "/internal/fetch-jobs/11/finish":
			return newJSONResponse(http.StatusNoContent, ``), nil
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return nil, nil
		}
	})}
	service.SetEventPublisher(&stubEventPublisher{
		publishFetchFailedFunc: func(context.Context, events.FetchFailedPayload) error {
			return errors.New("broker down")
		},
	})

	_, err := service.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "fetch rss status=502") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunReturnsErrorOnMalformedFeedAndFinishesFetchJob(t *testing.T) {
	t.Parallel()

	var finishPayload finishFetchJobPayload

	service := NewService("http://article-service")
	service.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/api/v1/sources":
			return newJSONResponse(http.StatusOK, `{"items":[{"id":7,"name":"Example","type":"rss","fetch_url":"http://feed-source/rss","fetch_method":"rss","interval_minutes":60,"default_category":"cloud","is_enabled":true}]}`), nil
		case "/rss":
			return newXMLResponse(http.StatusOK, `<rss><channel><item><title>broken</channel></rss>`), nil
		case "/internal/fetch-jobs/start":
			return newJSONResponse(http.StatusAccepted, `{"source_id":7,"job_id":11}`), nil
		case "/internal/fetch-jobs/11/finish":
			if err := json.NewDecoder(r.Body).Decode(&finishPayload); err != nil {
				t.Fatalf("decode finish payload: %v", err)
			}
			return newJSONResponse(http.StatusNoContent, ``), nil
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
			return nil, nil
		}
	})}

	_, err := service.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "decode rss") {
		t.Fatalf("unexpected error: %v", err)
	}
	if finishPayload.Status != "failed" || finishPayload.ErrorMessage == nil || !strings.Contains(*finishPayload.ErrorMessage, "decode rss") {
		t.Fatalf("unexpected finish payload after malformed feed: %+v", finishPayload)
	}
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return fmt.Sprintf("%x", sum[:])
}
