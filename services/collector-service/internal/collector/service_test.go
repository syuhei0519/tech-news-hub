package collector

import (
	"testing"
	"time"
)

func TestParsePubDate(t *testing.T) {
	t.Parallel()

	got, err := parsePubDate("Mon, 02 Jan 2006 15:04:05 +0900")
	if err != nil {
		t.Fatalf("parsePubDate returned error: %v", err)
	}

	want := time.Date(2006, 1, 2, 6, 4, 5, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("unexpected parsed time: got=%s want=%s", got, want)
	}
}

func TestParsePubDateReturnsErrorForUnsupportedLayout(t *testing.T) {
	t.Parallel()

	if _, err := parsePubDate("2026/04/14 12:00:00"); err == nil {
		t.Fatal("expected error for unsupported layout")
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
