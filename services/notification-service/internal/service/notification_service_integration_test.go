package service

import (
	"context"
	"testing"
	"time"

	"tech-feed-hub/notification-service/internal/domain"
	"tech-feed-hub/notification-service/internal/events"
	"tech-feed-hub/notification-service/internal/repository"
	"tech-feed-hub/notification-service/internal/testutil"
)

func TestNotificationServicePersistsEventDerivedNotifications(t *testing.T) {
	// consumer/service から永続化までを MySQL 実体で通し、
	// event 契約の変更で通知一覧に出る内容が壊れないことを確認する。
	db := testutil.OpenMySQLForTest(t)
	testutil.ResetMySQLTables(t, db)

	repo := repository.NewNotificationRepository(db)
	svc := NewNotificationService(repo)

	sourceID := testutil.InsertSource(t, db, "Event Source")
	startedAt := time.Date(2026, 4, 18, 8, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(5 * time.Minute)
	fetchJobID := testutil.InsertFetchJob(t, db, sourceID, "failed", startedAt, &finishedAt)

	occurredAt := time.Date(2026, 4, 18, 11, 0, 0, 0, time.UTC)
	title := "New article"
	if err := svc.HandleArticleIngested(context.Background(), events.Envelope[events.ArticleIngestedPayload]{
		EventID:    "evt-article-1",
		EventType:  events.EventTypeArticleIngested,
		OccurredAt: occurredAt,
		Payload: events.ArticleIngestedPayload{
			JobID:               fetchJobID,
			SourceID:            sourceID,
			SourceName:          "Event Source",
			InsertedCount:       2,
			RepresentativeTitle: &title,
		},
	}); err != nil {
		t.Fatalf("HandleArticleIngested returned error: %v", err)
	}

	if err := svc.HandleCollectorFetchFailed(context.Background(), events.Envelope[events.CollectorFetchFailedPayload]{
		EventID:    "evt-fetch-failed-1",
		EventType:  events.EventTypeCollectorFetchError,
		OccurredAt: occurredAt.Add(time.Minute),
		Payload: events.CollectorFetchFailedPayload{
			JobID:        fetchJobID,
			SourceID:     sourceID,
			SourceName:   "Event Source",
			ErrorMessage: "fetch rss status=502",
		},
	}); err != nil {
		t.Fatalf("HandleCollectorFetchFailed returned error: %v", err)
	}

	listed, err := repo.List(context.Background(), domain.ListNotificationsParams{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if listed.Total != 2 || len(listed.Items) != 2 {
		t.Fatalf("unexpected event-derived notifications: %+v", listed)
	}

	// 後着の fetch.failed が先頭に来る前提で、一覧の表示順と level/source 紐付けを固定する。
	latest := listed.Items[0]
	if latest.EventID != "evt-fetch-failed-1" || latest.Level != "error" || latest.SourceID == nil || *latest.SourceID != sourceID || latest.FetchJobID == nil || *latest.FetchJobID != fetchJobID {
		t.Fatalf("unexpected fetch-failed notification: %+v", latest)
	}

	// article.ingested は representative_title を本文に反映する現在仕様を守る。
	older := listed.Items[1]
	if older.EventID != "evt-article-1" || older.Level != "info" || older.Body != "最新記事: New article" {
		t.Fatalf("unexpected article-ingested notification: %+v", older)
	}
}
