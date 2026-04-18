package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	"tech-feed-hub/article-service/internal/domain"
	"tech-feed-hub/article-service/internal/testutil"
)

func TestSourceRepositoryCRUDAndConstraints(t *testing.T) {
	db := testutil.OpenMySQLForTest(t)
	testutil.ResetMySQLTables(t, db)

	repo := NewSourceRepository(db)

	created, err := repo.Create(context.Background(), domain.Source{
		Name:            "Kubernetes Blog",
		Type:            "rss",
		FetchURL:        "https://example.com/k8s.xml",
		FetchMethod:     "rss",
		IntervalMinutes: 60,
		DefaultCategory: "k8s",
		IsEnabled:       true,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if created == nil || created.ID < 1 || created.Name != "Kubernetes Blog" {
		t.Fatalf("unexpected created source: %+v", created)
	}

	listed, err := repo.List(context.Background())
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(listed) != 1 || listed[0].ID != created.ID {
		t.Fatalf("unexpected listed sources: %+v", listed)
	}

	updated, err := repo.Update(context.Background(), domain.Source{
		ID:              created.ID,
		Name:            "Kubernetes Official Blog",
		Type:            "rss",
		FetchURL:        "https://example.com/official.xml",
		FetchMethod:     "rss",
		IntervalMinutes: 30,
		DefaultCategory: "platform",
		IsEnabled:       false,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if updated == nil || updated.Name != "Kubernetes Official Blog" || updated.IsEnabled {
		t.Fatalf("unexpected updated source: %+v", updated)
	}

	lastError := "timeout"
	if err := repo.UpdateFetchStatus(context.Background(), created.ID, "failed", &lastError); err != nil {
		t.Fatalf("UpdateFetchStatus returned error: %v", err)
	}
	got, err := repo.GetByID(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if got == nil || got.LastFetchStatus == nil || *got.LastFetchStatus != "failed" || got.LastErrorMsg == nil || *got.LastErrorMsg != "timeout" || got.LastFetchedAt == nil {
		t.Fatalf("unexpected fetch status payload: %+v", got)
	}

	ensuredID, err := repo.EnsureSource(context.Background(), domain.Source{
		Name:            got.Name,
		Type:            "rss",
		FetchURL:        "https://example.com/reconciled.xml",
		FetchMethod:     "rss",
		IntervalMinutes: 45,
		DefaultCategory: "infra",
	})
	if err != nil {
		t.Fatalf("EnsureSource returned error: %v", err)
	}
	if ensuredID != created.ID {
		t.Fatalf("expected existing source id %d, got %d", created.ID, ensuredID)
	}

	duplicate, err := repo.Create(context.Background(), domain.Source{
		Name:            got.Name,
		Type:            "rss",
		FetchURL:        "https://example.com/dup.xml",
		FetchMethod:     "rss",
		IntervalMinutes: 10,
		DefaultCategory: "dup",
		IsEnabled:       true,
	})
	if err == nil || duplicate != nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("expected duplicate error, got source=%+v err=%v", duplicate, err)
	}

	deleted, err := repo.Delete(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if !deleted {
		t.Fatal("expected source to be deleted")
	}
}

func TestSourceRepositoryDeleteReturnsReferencedError(t *testing.T) {
	db := testutil.OpenMySQLForTest(t)
	testutil.ResetMySQLTables(t, db)

	repo := NewSourceRepository(db)
	sourceID := testutil.InsertSource(t, db, domain.Source{
		Name:            "Protected Source",
		Type:            "rss",
		FetchURL:        "https://example.com/protected.xml",
		FetchMethod:     "rss",
		IntervalMinutes: 60,
		DefaultCategory: "ops",
		IsEnabled:       true,
	})

	published := time.Date(2026, 4, 15, 10, 0, 0, 0, time.UTC)
	testutil.InsertArticle(t, db, domain.Article{
		Title:       "Dependent article",
		URL:         "https://example.com/dependent",
		SourceID:    sourceID,
		PublishedAt: &published,
		FetchedAt:   published,
		Excerpt:     "dependent",
		Category:    "ops",
		DedupeKey:   "protected-1",
	})

	deleted, err := repo.Delete(context.Background(), sourceID)
	if err == nil || deleted || !strings.Contains(err.Error(), "referenced") {
		t.Fatalf("expected referenced delete error, got deleted=%t err=%v", deleted, err)
	}
}
