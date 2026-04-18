package httpapi

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"tech-feed-hub/article-service/internal/domain"
	"tech-feed-hub/article-service/internal/repository"
	"tech-feed-hub/article-service/internal/service"
	"tech-feed-hub/article-service/internal/testutil"
)

func TestRouterArticleEndpointsAndCSV(t *testing.T) {
	// 記事系 endpoint の 400/404/200 と CSV content-type をまとめて確認する。
	db, router := newIntegrationRouter(t)

	sourceID := testutil.InsertSource(t, db, domain.Source{
		Name:            "Router Source",
		Type:            "rss",
		FetchURL:        "https://example.com/router.xml",
		FetchMethod:     "rss",
		IntervalMinutes: 30,
		DefaultCategory: "k8s",
		IsEnabled:       true,
	})
	published := time.Date(2026, 4, 16, 8, 0, 0, 0, time.UTC)
	articleID := testutil.InsertArticle(t, db, domain.Article{
		Title:       "Router article",
		URL:         "https://example.com/router-article",
		SourceID:    sourceID,
		PublishedAt: &published,
		FetchedAt:   published.Add(time.Minute),
		Excerpt:     "router excerpt",
		Category:    "k8s",
		Tags:        []string{"router"},
		IsRead:      false,
		IsFavorite:  true,
		DedupeKey:   "router-article",
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/articles?is_read=bad", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid bool query, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/articles?q=Router&is_favorite=true", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for article list, got %d body=%s", rec.Code, rec.Body.String())
	}
	var listResp struct {
		Items []domain.Article `json:"items"`
		Total int              `json:"total"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal article list: %v", err)
	}
	if listResp.Total != 1 || len(listResp.Items) != 1 || listResp.Items[0].ID != articleID {
		t.Fatalf("unexpected list response: %+v", listResp)
	}

	// 詳細と状態更新は未存在 article を 404 にそろえる必要がある。
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/articles/999999", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing article, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/articles/999999/favorite-status", strings.NewReader(`{"is_favorite":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing favorite target, got %d body=%s", rec.Code, rec.Body.String())
	}

	// CSV endpoint は成功時ヘッダと本文の両方を確認する。
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/articles/export.csv?source_id=1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for csv export, got %d body=%s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Type"); got != "text/csv; charset=utf-8" {
		t.Fatalf("unexpected csv content type: %s", got)
	}
	if body := rec.Body.String(); !strings.Contains(body, "title,url,source_name,category,published_at,fetched_at,is_read,is_favorite,excerpt,tags") || !strings.Contains(body, "Router article") {
		t.Fatalf("unexpected csv body: %s", body)
	}

	// query parse failure は service まで入る前に 400 を返す。
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/articles/export.csv?from=bad", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid from query, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouterSourceEndpointsAndServerError(t *testing.T) {
	// source 系 endpoint の 409 と、repository 障害時の 500 を確認する。
	db, router := newIntegrationRouter(t)

	sourceID := testutil.InsertSource(t, db, domain.Source{
		Name:            "Existing Source",
		Type:            "rss",
		FetchURL:        "https://example.com/existing.xml",
		FetchMethod:     "rss",
		IntervalMinutes: 60,
		DefaultCategory: "ops",
		IsEnabled:       true,
	})
	published := time.Date(2026, 4, 16, 9, 0, 0, 0, time.UTC)
	testutil.InsertArticle(t, db, domain.Article{
		Title:       "Dependent",
		URL:         "https://example.com/dependent",
		SourceID:    sourceID,
		PublishedAt: &published,
		FetchedAt:   published,
		Category:    "ops",
		DedupeKey:   "dependent",
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/sources", strings.NewReader(`{"name":"Existing Source","type":"rss","fetch_url":"https://example.com/another.xml","fetch_method":"rss","interval_minutes":15,"default_category":"ops","is_enabled":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate source, got %d body=%s", rec.Code, rec.Body.String())
	}

	// 参照中 source の削除失敗も 409 に統一する。
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/sources/"+int64String(sourceID), nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for referenced source delete, got %d body=%s", rec.Code, rec.Body.String())
	}

	if err := db.Close(); err != nil {
		t.Fatalf("close db for 500 test: %v", err)
	}

	// DB 障害は writeServiceError 経由で 500 になる。
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/sources", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 after db close, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouterFetchJobEndpointsAndIngestConflict(t *testing.T) {
	// fetch job 系 endpoint の 400/404/202/204/409 を一通り確認する。
	db, router := newIntegrationRouter(t)

	sourceID := testutil.InsertSource(t, db, domain.Source{
		Name:            "Job Source",
		Type:            "rss",
		FetchURL:        "https://example.com/jobs.xml",
		FetchMethod:     "rss",
		IntervalMinutes: 20,
		DefaultCategory: "jobs",
		IsEnabled:       true,
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/fetch-jobs", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing source_id, got %d body=%s", rec.Code, rec.Body.String())
	}

	// start は未存在 source を 404 にし、既存 source は 202 で job を返す。
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/internal/fetch-jobs/start", strings.NewReader(`{"source_id":999999}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing start source, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/internal/fetch-jobs/start", strings.NewReader(`{"source_id":`+int64String(sourceID)+`}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for start fetch job, got %d body=%s", rec.Code, rec.Body.String())
	}
	var startResp struct {
		SourceID int64 `json:"source_id"`
		JobID    int64 `json:"job_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &startResp); err != nil {
		t.Fatalf("unmarshal start fetch job response: %v", err)
	}
	if startResp.SourceID != sourceID || startResp.JobID < 1 {
		t.Fatalf("unexpected start response: %+v", startResp)
	}

	// finish は初回のみ成功し、完了済み job の再完了は conflict になる。
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/internal/fetch-jobs/"+int64String(startResp.JobID)+"/finish", bytes.NewBufferString(`{"status":"failed","fetched_count":3,"inserted_count":1,"duplicated_count":2,"error_message":"network error"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for finish fetch job, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/internal/fetch-jobs/"+int64String(startResp.JobID)+"/finish", bytes.NewBufferString(`{"status":"success","fetched_count":3,"inserted_count":1,"duplicated_count":2}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for finishing completed job, got %d body=%s", rec.Code, rec.Body.String())
	}

	// ingest も完了済み job に対しては conflict を返す。
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/internal/ingest", bytes.NewBufferString(`{"job_id":`+int64String(startResp.JobID)+`,"source_id":`+int64String(sourceID)+`,"source":{"name":"Job Source","default_category":"jobs"},"articles":[{"title":"New article","url":"https://example.com/new","fetched_at":"2026-04-16T10:00:00Z","dedupe_key":"new-article"}]}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for ingest on finished job, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouterInternalEndpointsValidateContract(t *testing.T) {
	// collector が叩く internal API の 400 shape を固定し、service 間契約の破壊を router 境界で検知する。
	db, router := newIntegrationRouter(t)

	sourceID := testutil.InsertSource(t, db, domain.Source{
		Name:            "Contract Source",
		Type:            "rss",
		FetchURL:        "https://example.com/contract.xml",
		FetchMethod:     "rss",
		IntervalMinutes: 20,
		DefaultCategory: "contracts",
		IsEnabled:       true,
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/fetch-jobs/start", strings.NewReader(`{"source":{"name":"Contract Source"}}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "source.default_category is required") {
		t.Fatalf("expected 400 for invalid start request, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/internal/fetch-jobs/start", strings.NewReader(`{"source_id":`+int64String(sourceID)+`}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected 202 for start fetch job, got %d body=%s", rec.Code, rec.Body.String())
	}
	var startResp struct {
		JobID int64 `json:"job_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &startResp); err != nil {
		t.Fatalf("unmarshal start fetch job response: %v", err)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/internal/ingest", bytes.NewBufferString(`{"job_id":`+int64String(startResp.JobID)+`,"source_id":`+int64String(sourceID)+`,"source":{"name":"Contract Source","default_category":"contracts"},"articles":[{"url":"https://example.com/new","fetched_at":"2026-04-16T10:00:00Z","dedupe_key":"new-article"}]}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "articles[0].title is required") {
		t.Fatalf("expected 400 for invalid ingest request, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/internal/fetch-jobs/"+int64String(startResp.JobID)+"/finish", bytes.NewBufferString(`{"status":"failed","fetched_count":1,"inserted_count":0,"duplicated_count":1}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest || !strings.Contains(rec.Body.String(), "error_message is required") {
		t.Fatalf("expected 400 for invalid finish request, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func newIntegrationRouter(t *testing.T) (*sql.DB, http.Handler) {
	t.Helper()

	// router test は本番と同じ repository/service 構成を立てて handler 境界を見る。
	db := testutil.OpenMySQLForTest(t)
	testutil.ResetMySQLTables(t, db)

	articleRepo := repository.NewArticleRepository(db)
	sourceRepo := repository.NewSourceRepository(db)
	jobRepo := repository.NewFetchJobRepository(db)
	articleService := service.NewArticleService(articleRepo, sourceRepo, jobRepo)

	return db, NewRouter(db, articleService)
}

func int64String(value int64) string {
	return strconv.FormatInt(value, 10)
}
