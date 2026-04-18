package repository

import (
	"context"
	"testing"
	"time"

	"tech-feed-hub/article-service/internal/domain"
	"tech-feed-hub/article-service/internal/testutil"
)

func TestArticleRepositoryListExportAndStatusUpdates(t *testing.T) {
	// 検索条件、ページング、CSV export、状態更新が MySQL 実体で崩れないことを確認する。
	db := testutil.OpenMySQLForTest(t)
	testutil.ResetMySQLTables(t, db)

	repo := NewArticleRepository(db)

	sourceAID := testutil.InsertSource(t, db, domain.Source{
		Name:            "Source A",
		Type:            "rss",
		FetchURL:        "https://example.com/a.xml",
		FetchMethod:     "rss",
		IntervalMinutes: 60,
		DefaultCategory: "k8s",
		IsEnabled:       true,
	})
	sourceBID := testutil.InsertSource(t, db, domain.Source{
		Name:            "Source B",
		Type:            "rss",
		FetchURL:        "https://example.com/b.xml",
		FetchMethod:     "rss",
		IntervalMinutes: 30,
		DefaultCategory: "infra",
		IsEnabled:       true,
	})

	publishedA1 := time.Date(2026, 4, 10, 10, 0, 0, 0, time.UTC)
	publishedA2 := time.Date(2026, 4, 12, 10, 0, 0, 0, time.UTC)
	fetchedBase := time.Date(2026, 4, 12, 11, 0, 0, 0, time.UTC)

	articleA1ID := testutil.InsertArticle(t, db, domain.Article{
		Title:       "Kubernetes release",
		URL:         "https://example.com/articles/1",
		SourceID:    sourceAID,
		PublishedAt: &publishedA1,
		FetchedAt:   fetchedBase,
		Excerpt:     "release note",
		Category:    "k8s",
		Tags:        []string{"go", "k8s"},
		IsRead:      false,
		IsFavorite:  true,
		DedupeKey:   "dedupe-1",
	})
	testutil.InsertArticle(t, db, domain.Article{
		Title:       "Kubernetes deep dive",
		URL:         "https://example.com/articles/2",
		SourceID:    sourceAID,
		PublishedAt: &publishedA2,
		FetchedAt:   fetchedBase.Add(2 * time.Hour),
		Excerpt:     "deep dive",
		Category:    "k8s",
		Tags:        []string{"k8s"},
		IsRead:      false,
		IsFavorite:  true,
		DedupeKey:   "dedupe-2",
	})
	testutil.InsertArticle(t, db, domain.Article{
		Title:      "Terraform plan",
		URL:        "https://example.com/articles/3",
		SourceID:   sourceBID,
		FetchedAt:  fetchedBase.Add(time.Hour),
		Excerpt:    "iac",
		Category:   "infra",
		Tags:       []string{"terraform"},
		IsRead:     true,
		IsFavorite: false,
		DedupeKey:  "dedupe-3",
	})

	isRead := false
	isFavorite := true
	from := time.Date(2026, 4, 9, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 4, 13, 0, 0, 0, 0, time.UTC)

	// 一覧は filter と sort を SQL で解決するため、返却件数と先頭行の両方を見る。
	listResult, err := repo.List(context.Background(), domain.ListArticlesParams{
		ArticleFilterParams: domain.ArticleFilterParams{
			Query:         "Kubernetes",
			Category:      "k8s",
			SourceID:      sourceAID,
			IsRead:        &isRead,
			IsFavorite:    &isFavorite,
			PublishedFrom: &from,
			PublishedTo:   &to,
		},
		Page:     1,
		PageSize: 1,
	})
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if listResult.Total != 2 || listResult.TotalPages != 2 || len(listResult.Items) != 1 {
		t.Fatalf("unexpected list meta: %+v", listResult)
	}
	if got := listResult.Items[0]; got.Title != "Kubernetes deep dive" || got.SourceName != "Source A" {
		t.Fatalf("unexpected first item: %+v", got)
	}

	// CSV export は limit+1 件を読み、service 層が上限超過を判定できる前提を守る。
	exported, err := repo.Export(context.Background(), domain.ExportArticlesParams{
		ArticleFilterParams: domain.ArticleFilterParams{
			SourceID:   sourceAID,
			Sort:       "created_at",
			Order:      "asc",
			IsFavorite: &isFavorite,
		},
		Limit: 1,
	})
	if err != nil {
		t.Fatalf("Export returned error: %v", err)
	}
	if len(exported) != 2 {
		t.Fatalf("expected limit+1 rows for over-limit detection, got %d", len(exported))
	}
	if exported[0].Title != "Kubernetes release" || exported[1].Title != "Kubernetes deep dive" {
		t.Fatalf("unexpected export order: %+v", exported)
	}

	// 既読・お気に入り更新は更新後の最新行を再取得して返す契約を守る。
	updatedRead, err := repo.UpdateReadStatus(context.Background(), articleA1ID, true)
	if err != nil {
		t.Fatalf("UpdateReadStatus returned error: %v", err)
	}
	if updatedRead == nil || !updatedRead.IsRead {
		t.Fatalf("unexpected read-status update result: %+v", updatedRead)
	}

	updatedFavorite, err := repo.UpdateFavoriteStatus(context.Background(), articleA1ID, false)
	if err != nil {
		t.Fatalf("UpdateFavoriteStatus returned error: %v", err)
	}
	if updatedFavorite == nil || updatedFavorite.IsFavorite {
		t.Fatalf("unexpected favorite-status update result: %+v", updatedFavorite)
	}

	// 更新対象が存在しない場合は not found 相当として nil を返す。
	missing, err := repo.UpdateFavoriteStatus(context.Background(), 999999, true)
	if err != nil {
		t.Fatalf("unexpected missing update error: %v", err)
	}
	if missing != nil {
		t.Fatalf("expected nil for missing article, got %+v", missing)
	}
}

func TestArticleRepositoryBulkUpsertCountsInsertedAndDuplicated(t *testing.T) {
	// dedupe_key の unique 制約を使った insert / duplicate 判定が件数に反映されることを確認する。
	db := testutil.OpenMySQLForTest(t)
	testutil.ResetMySQLTables(t, db)

	repo := NewArticleRepository(db)
	sourceID := testutil.InsertSource(t, db, domain.Source{
		Name:            "Collector",
		Type:            "rss",
		FetchURL:        "https://example.com/feed.xml",
		FetchMethod:     "rss",
		IntervalMinutes: 15,
		DefaultCategory: "cloud",
		IsEnabled:       true,
	})

	published := time.Date(2026, 4, 14, 9, 0, 0, 0, time.UTC)
	fetched := time.Date(2026, 4, 14, 9, 30, 0, 0, time.UTC)
	inserted, duplicated, err := repo.BulkUpsert(context.Background(), sourceID, []domain.Article{
		{
			Title:       "First title",
			URL:         "https://example.com/first",
			PublishedAt: &published,
			FetchedAt:   fetched,
			Excerpt:     "first",
			Category:    "cloud",
			Tags:        []string{"aws"},
			DedupeKey:   "bulk-1",
		},
		{
			Title:       "Second title",
			URL:         "https://example.com/second",
			PublishedAt: &published,
			FetchedAt:   fetched.Add(time.Minute),
			Excerpt:     "second",
			Category:    "cloud",
			Tags:        []string{"gcp"},
			DedupeKey:   "bulk-2",
		},
	})
	if err != nil {
		t.Fatalf("initial BulkUpsert returned error: %v", err)
	}
	if inserted != 2 || duplicated != 0 {
		t.Fatalf("unexpected initial counts: inserted=%d duplicated=%d", inserted, duplicated)
	}

	// duplicate 側は件数だけでなく、更新列が実際に反映されることまで見る。
	inserted, duplicated, err = repo.BulkUpsert(context.Background(), sourceID, []domain.Article{
		{
			Title:       "First title updated",
			URL:         "https://example.com/first",
			PublishedAt: &published,
			FetchedAt:   fetched,
			Excerpt:     "updated",
			Category:    "cloud",
			Tags:        []string{"aws", "eks"},
			DedupeKey:   "bulk-1",
		},
		{
			Title:       "Third title",
			URL:         "https://example.com/third",
			PublishedAt: &published,
			FetchedAt:   fetched.Add(2 * time.Minute),
			Excerpt:     "third",
			Category:    "cloud",
			Tags:        []string{"azure"},
			DedupeKey:   "bulk-3",
		},
	})
	if err != nil {
		t.Fatalf("second BulkUpsert returned error: %v", err)
	}
	if inserted != 1 || duplicated != 1 {
		t.Fatalf("unexpected second counts: inserted=%d duplicated=%d", inserted, duplicated)
	}

	updatedArticle, err := repo.GetByID(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetByID returned error: %v", err)
	}
	if updatedArticle == nil || updatedArticle.Title != "First title updated" || updatedArticle.Excerpt != "updated" {
		t.Fatalf("unexpected updated article: %+v", updatedArticle)
	}
}
