package collector

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

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

	var startCalled bool
	var finishPayload finishFetchJobPayload
	var ingestPayload IngestPayload

	service := NewService("http://article-service", []SourceConfig{{
		Name:            "Example",
		Type:            "rss",
		URL:             "http://feed-source/rss",
		DefaultCategory: "cloud",
	}})
	service.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/rss":
			return newXMLResponse(http.StatusOK, `<?xml version="1.0"?><rss><channel><item><title>Hello</title><link>https://example.com/1</link><description>desc</description><pubDate>Mon, 02 Jan 2006 15:04:05 +0900</pubDate></item></channel></rss>`), nil
		case "/internal/fetch-jobs/start":
			startCalled = true
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

	if !startCalled {
		t.Fatal("expected start to be called")
	}
	if ingestPayload.JobID != 11 || ingestPayload.SourceID != 7 {
		t.Fatalf("unexpected ingest payload ids: %+v", ingestPayload)
	}
	if finishPayload.Status != "success" || finishPayload.FetchedCount != 1 || finishPayload.InsertedCount != 1 || finishPayload.DuplicatedCount != 0 || finishPayload.ErrorMessage != nil {
		t.Fatalf("unexpected finish payload: %+v", finishPayload)
	}
	if len(results) != 1 || results[0].Status != "success" || results[0].JobID != 11 || results[0].SourceID != 7 {
		t.Fatalf("unexpected results: %+v", results)
	}
}

func TestRunFinishesFetchJobOnRSSFailure(t *testing.T) {
	t.Parallel()

	var finishPayload finishFetchJobPayload
	var publishedPayload events.FetchFailedPayload

	service := NewService("http://article-service", []SourceConfig{{
		Name:            "Example",
		Type:            "rss",
		URL:             "http://feed-source/rss",
		DefaultCategory: "cloud",
	}})
	service.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case r.URL.Path == "/rss":
			return newJSONResponse(http.StatusBadGateway, "boom\n"), nil
		case r.URL.Path == "/internal/fetch-jobs/start":
			return newJSONResponse(http.StatusAccepted, `{"source_id":7,"job_id":11}`), nil
		case r.URL.Path == "/internal/fetch-jobs/11/finish":
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

	service := NewService("http://article-service", []SourceConfig{{
		Name:            "Example",
		Type:            "rss",
		URL:             "http://feed-source/rss",
		DefaultCategory: "cloud",
	}})
	service.client = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case r.URL.Path == "/rss":
			return newJSONResponse(http.StatusBadGateway, "boom\n"), nil
		case r.URL.Path == "/internal/fetch-jobs/start":
			return newJSONResponse(http.StatusAccepted, `{"source_id":7,"job_id":11}`), nil
		case r.URL.Path == "/internal/fetch-jobs/11/finish":
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
