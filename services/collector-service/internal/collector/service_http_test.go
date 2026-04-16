package collector

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunCreatesAndFinishesFetchJobOnSuccess(t *testing.T) {
	t.Parallel()

	var startCalled bool
	var finishPayload finishFetchJobPayload
	var ingestPayload IngestPayload

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/rss":
			w.Header().Set("Content-Type", "application/xml")
			_, _ = io.WriteString(w, `<?xml version="1.0"?><rss><channel><item><title>Hello</title><link>https://example.com/1</link><description>desc</description><pubDate>Mon, 02 Jan 2006 15:04:05 +0900</pubDate></item></channel></rss>`)
		case "/internal/fetch-jobs/start":
			startCalled = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"source_id":7,"job_id":11}`)
		case "/internal/ingest":
			if err := json.NewDecoder(r.Body).Decode(&ingestPayload); err != nil {
				t.Fatalf("decode ingest payload: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"source_id":7,"job_id":11,"fetched_count":1,"inserted_count":1,"duplicated_count":0}`)
		case "/internal/fetch-jobs/11/finish":
			if err := json.NewDecoder(r.Body).Decode(&finishPayload); err != nil {
				t.Fatalf("decode finish payload: %v", err)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewService(server.URL, []SourceConfig{{
		Name:            "Example",
		Type:            "rss",
		URL:             server.URL + "/rss",
		DefaultCategory: "cloud",
	}})

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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/rss":
			http.Error(w, "boom", http.StatusBadGateway)
		case r.URL.Path == "/internal/fetch-jobs/start":
			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"source_id":7,"job_id":11}`)
		case r.URL.Path == "/internal/fetch-jobs/11/finish":
			if err := json.NewDecoder(r.Body).Decode(&finishPayload); err != nil {
				t.Fatalf("decode finish payload: %v", err)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	service := NewService(server.URL, []SourceConfig{{
		Name:            "Example",
		Type:            "rss",
		URL:             server.URL + "/rss",
		DefaultCategory: "cloud",
	}})

	_, err := service.Run(context.Background())
	if err == nil || !strings.Contains(err.Error(), "fetch rss status=502") {
		t.Fatalf("unexpected error: %v", err)
	}
	if finishPayload.Status != "failed" || finishPayload.ErrorMessage == nil || !strings.Contains(*finishPayload.ErrorMessage, "fetch rss status=502") {
		t.Fatalf("unexpected finish payload: %+v", finishPayload)
	}
}
