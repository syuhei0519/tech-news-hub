package httpapi

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"tech-feed-hub/notification-service/internal/domain"
	"tech-feed-hub/notification-service/internal/repository"
	"tech-feed-hub/notification-service/internal/service"
	"tech-feed-hub/notification-service/internal/testutil"
)

func TestRouterNotificationEndpoints(t *testing.T) {
	// 通知一覧と read-status 更新の 200/400/404 を handler 境界でまとめて確認する。
	db, router := newIntegrationRouter(t)

	sourceID := testutil.InsertSource(t, db, "Router Source")
	startedAt := time.Date(2026, 4, 18, 8, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(5 * time.Minute)
	fetchJobID := testutil.InsertFetchJob(t, db, sourceID, "failed", startedAt, &finishedAt)

	notificationID := testutil.InsertNotification(t, db, domain.Notification{
		EventID:    "evt-router-1",
		EventType:  "collector.fetch.failed",
		Level:      "error",
		Title:      "Collector fetch failed",
		Body:       "fetch rss status=502",
		SourceID:   &sourceID,
		FetchJobID: &fetchJobID,
		IsRead:     false,
		CreatedAt:  time.Date(2026, 4, 18, 9, 0, 0, 0, time.UTC),
	})
	readAt := time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC)
	testutil.InsertNotification(t, db, domain.Notification{
		EventID:    "evt-router-2",
		EventType:  "article.ingested",
		Level:      "info",
		Title:      "Article ingested",
		Body:       "new article",
		SourceID:   &sourceID,
		FetchJobID: &fetchJobID,
		IsRead:     true,
		CreatedAt:  time.Date(2026, 4, 18, 10, 0, 0, 0, time.UTC),
		ReadAt:     &readAt,
	})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications?is_read=false&page=1&page_size=1", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for notifications list, got %d body=%s", rec.Code, rec.Body.String())
	}
	var listResp domain.ListNotificationsResult
	if err := json.Unmarshal(rec.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal notifications list: %v", err)
	}
	if listResp.Total != 1 || len(listResp.Items) != 1 || listResp.Items[0].ID != notificationID {
		t.Fatalf("unexpected notifications response: %+v", listResp)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/notifications?is_read=bad", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid bool query, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/notifications?page_size=101", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid page_size, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/"+int64String(notificationID)+"/read-status", strings.NewReader(`{"is_read":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for read-status update, got %d body=%s", rec.Code, rec.Body.String())
	}
	var updated domain.Notification
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("unmarshal updated notification: %v", err)
	}
	if !updated.IsRead || updated.ReadAt == nil {
		t.Fatalf("expected read notification with read_at, got %+v", updated)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/bad/read-status", strings.NewReader(`{"is_read":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid id, got %d body=%s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v1/notifications/999999/read-status", strings.NewReader(`{"is_read":true}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing notification, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestRouterNotificationEndpointsReturnServerErrorOnDBFailure(t *testing.T) {
	// DB 障害時は service / repository error を 500 として返す。
	db, router := newIntegrationRouter(t)
	if err := db.Close(); err != nil {
		t.Fatalf("close db for 500 test: %v", err)
	}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 after db close, got %d body=%s", rec.Code, rec.Body.String())
	}
}

func newIntegrationRouter(t *testing.T) (*sql.DB, http.Handler) {
	t.Helper()

	db := testutil.OpenMySQLForTest(t)
	testutil.ResetMySQLTables(t, db)

	notificationRepo := repository.NewNotificationRepository(db)
	notificationService := service.NewNotificationService(notificationRepo)

	return db, NewRouter(db, notificationService)
}

func int64String(value int64) string {
	return strconv.FormatInt(value, 10)
}
