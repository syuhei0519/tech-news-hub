package repository

import (
	"context"
	"testing"
	"time"

	"tech-feed-hub/notification-service/internal/domain"
	"tech-feed-hub/notification-service/internal/testutil"
)

func TestNotificationRepositoryListAndReadStatus(t *testing.T) {
	// 通知一覧の filter / paging と read-status 更新時の read_at 契約を MySQL 実体で固定する。
	db := testutil.OpenMySQLForTest(t)
	testutil.ResetMySQLTables(t, db)

	repo := NewNotificationRepository(db)

	sourceID := testutil.InsertSource(t, db, "Notification Source")
	startedAt := time.Date(2026, 4, 18, 8, 0, 0, 0, time.UTC)
	finishedAt := startedAt.Add(5 * time.Minute)
	fetchJobID := testutil.InsertFetchJob(t, db, sourceID, "success", startedAt, &finishedAt)

	readAt := time.Date(2026, 4, 18, 10, 30, 0, 0, time.UTC)
	olderCreatedAt := time.Date(2026, 4, 18, 9, 0, 0, 0, time.UTC)
	newerCreatedAt := olderCreatedAt.Add(time.Hour)

	olderID := testutil.InsertNotification(t, db, domain.Notification{
		EventID:    "evt-older",
		EventType:  "collector.fetch.failed",
		Level:      "error",
		Title:      "Older notification",
		Body:       "older body",
		SourceID:   &sourceID,
		FetchJobID: &fetchJobID,
		IsRead:     true,
		CreatedAt:  olderCreatedAt,
		ReadAt:     &readAt,
	})
	newerID := testutil.InsertNotification(t, db, domain.Notification{
		EventID:    "evt-newer",
		EventType:  "article.ingested",
		Level:      "info",
		Title:      "Newer notification",
		Body:       "newer body",
		SourceID:   &sourceID,
		FetchJobID: &fetchJobID,
		IsRead:     false,
		CreatedAt:  newerCreatedAt,
	})

	// unread filter は新しい通知だけを返し、total と total_pages も filter 後件数で計算される必要がある。
	isRead := false
	listed, err := repo.List(context.Background(), domain.ListNotificationsParams{
		IsRead:   &isRead,
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if listed.Total != 1 || listed.TotalPages != 1 || len(listed.Items) != 1 {
		t.Fatalf("unexpected list result: %+v", listed)
	}
	if listed.Items[0].ID != newerID || listed.Items[0].ReadAt != nil {
		t.Fatalf("unexpected unread item: %+v", listed.Items[0])
	}

	// 全件一覧は created_at DESC, id DESC 順に返る前提で画面の先頭表示を守る。
	all, err := repo.List(context.Background(), domain.ListNotificationsParams{
		Page:     1,
		PageSize: 10,
	})
	if err != nil {
		t.Fatalf("List all returned error: %v", err)
	}
	if len(all.Items) != 2 || all.Items[0].ID != newerID || all.Items[1].ID != olderID {
		t.Fatalf("unexpected all notifications order: %+v", all.Items)
	}

	got, err := repo.GetByID(context.Background(), olderID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if got == nil || got.SourceID == nil || *got.SourceID != sourceID || got.FetchJobID == nil || *got.FetchJobID != fetchJobID || got.ReadAt == nil {
		t.Fatalf("unexpected notification payload: %+v", got)
	}

	updatedRead, err := repo.UpdateReadStatus(context.Background(), newerID, true)
	if err != nil {
		t.Fatalf("UpdateReadStatus(true) returned error: %v", err)
	}
	if updatedRead == nil || !updatedRead.IsRead || updatedRead.ReadAt == nil {
		t.Fatalf("expected read notification with read_at, got %+v", updatedRead)
	}

	updatedUnread, err := repo.UpdateReadStatus(context.Background(), newerID, false)
	if err != nil {
		t.Fatalf("UpdateReadStatus(false) returned error: %v", err)
	}
	if updatedUnread == nil || updatedUnread.IsRead || updatedUnread.ReadAt != nil {
		t.Fatalf("expected unread notification without read_at, got %+v", updatedUnread)
	}

	missing, err := repo.UpdateReadStatus(context.Background(), 999999, true)
	if err != nil {
		t.Fatalf("unexpected missing update error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing notification, got %+v", missing)
	}

	// page 境界でも total / total_pages と並び順が崩れず、画面のページ送りと整合することを確認する。
	secondPage, err := repo.List(context.Background(), domain.ListNotificationsParams{
		Page:     2,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("List second page returned error: %v", err)
	}
	if secondPage.Total != 2 || secondPage.TotalPages != 2 || len(secondPage.Items) != 1 || secondPage.Items[0].ID != olderID {
		t.Fatalf("unexpected second page result: %+v", secondPage)
	}
}
