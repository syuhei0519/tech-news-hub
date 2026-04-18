package collector

import (
	"strings"
	"testing"
	"time"
)

func TestParsePubDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    time.Time
		wantErr string
	}{
		{
			name: "rfc1123z with timezone",
			raw:  "Mon, 02 Jan 2006 15:04:05 +0900",
			want: time.Date(2006, 1, 2, 6, 4, 5, 0, time.UTC),
		},
		{
			name: "rfc1123 utc",
			raw:  "Mon, 02 Jan 2006 15:04:05 UTC",
			want: time.Date(2006, 1, 2, 15, 4, 5, 0, time.UTC),
		},
		{
			name: "rfc3339",
			raw:  "2006-01-02T15:04:05-07:00",
			want: time.Date(2006, 1, 2, 22, 4, 5, 0, time.UTC),
		},
		{
			name:    "empty",
			raw:     "   ",
			wantErr: "unsupported pubDate",
		},
		{
			name:    "unsupported layout",
			raw:     "2026/04/14 12:00:00",
			wantErr: "unsupported pubDate",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := parsePubDate(tt.raw)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parsePubDate returned error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Fatalf("unexpected parsed time: got=%s want=%s", got, tt.want)
			}
		})
	}
}

func TestDedupeKeyTrimsWhitespace(t *testing.T) {
	t.Parallel()

	left := dedupeKey(" https://example.com/articles/1 ")
	right := dedupeKey("https://example.com/articles/1")
	if left != right {
		t.Fatalf("dedupe key should ignore surrounding whitespace: left=%s right=%s", left, right)
	}
}

func TestStripHTMLNormalizesText(t *testing.T) {
	t.Parallel()

	got := stripHTML("<p>Hello&nbsp;World</p><br />A &amp; B")
	want := "Hello World A & B"
	if got != want {
		t.Fatalf("unexpected stripped text: got=%q want=%q", got, want)
	}
}

func TestNormalizeRSSItem(t *testing.T) {
	t.Parallel()

	fetchedAt := time.Date(2026, 4, 18, 1, 2, 3, 0, time.FixedZone("JST", 9*60*60))
	source := SourceConfig{DefaultCategory: "cloud"}

	article, ok := normalizeRSSItem(rssItem{
		Title:       "  Launch <b>News</b>  ",
		Link:        " https://example.com/articles/1 ",
		Description: "<p>Hello&nbsp;World</p><br />A &amp; B",
		PubDate:     "Mon, 02 Jan 2006 15:04:05 +0900",
	}, source, fetchedAt)
	if !ok {
		t.Fatal("expected item to be normalized")
	}
	if article.Title != "Launch <b>News</b>" {
		t.Fatalf("unexpected title: %q", article.Title)
	}
	if article.URL != "https://example.com/articles/1" {
		t.Fatalf("unexpected url: %q", article.URL)
	}
	if article.Excerpt != "Hello World A & B" {
		t.Fatalf("unexpected excerpt: %q", article.Excerpt)
	}
	if article.Category != "cloud" || len(article.Tags) != 1 || article.Tags[0] != "cloud" {
		t.Fatalf("unexpected category/tags: %+v", article)
	}
	if article.DedupeKey != dedupeKey(" https://example.com/articles/1 ") {
		t.Fatalf("unexpected dedupe key: %q", article.DedupeKey)
	}
	if !article.FetchedAt.Equal(fetchedAt.UTC()) {
		t.Fatalf("unexpected fetched_at: %s", article.FetchedAt)
	}
	if article.PublishedAt == nil || !article.PublishedAt.Equal(time.Date(2006, 1, 2, 6, 4, 5, 0, time.UTC)) {
		t.Fatalf("unexpected published_at: %+v", article.PublishedAt)
	}
}

func TestNormalizeRSSItemSkipsMissingRequiredFields(t *testing.T) {
	t.Parallel()

	for _, item := range []rssItem{
		{Title: "only title"},
		{Link: "https://example.com/articles/1"},
	} {
		if _, ok := normalizeRSSItem(item, SourceConfig{DefaultCategory: "cloud"}, time.Now().UTC()); ok {
			t.Fatalf("expected item to be skipped: %+v", item)
		}
	}
}

func TestNormalizeRSSItemAllowsMissingPubDate(t *testing.T) {
	t.Parallel()

	article, ok := normalizeRSSItem(rssItem{
		Title:       "hello",
		Link:        "https://example.com/articles/1",
		Description: "<p>line</p>",
		PubDate:     "",
	}, SourceConfig{DefaultCategory: "ops"}, time.Now().UTC())
	if !ok {
		t.Fatal("expected item to be normalized")
	}
	if article.PublishedAt != nil {
		t.Fatalf("expected published_at to be nil, got %+v", article.PublishedAt)
	}
	if article.Excerpt != "line" {
		t.Fatalf("unexpected excerpt: %q", article.Excerpt)
	}
}
