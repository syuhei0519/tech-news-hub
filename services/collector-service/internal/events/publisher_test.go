package events

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestMarshalFetchFailedEventMatchesContractFixture(t *testing.T) {
	t.Parallel()

	body, err := marshalFetchFailedEvent(
		"evt-fetch-failed-1",
		time.Date(2026, 4, 18, 3, 4, 5, 0, time.UTC),
		FetchFailedPayload{
			JobID:        42,
			SourceID:     7,
			SourceName:   "Example Feed",
			ErrorMessage: "fetch rss status=502",
		},
	)
	if err != nil {
		t.Fatalf("marshalFetchFailedEvent returned error: %v", err)
	}

	assertJSONFixtureEqual(t, fixturePath("collector.fetch.failed.json"), body)
}

func fixturePath(name string) string {
	return filepath.Join("..", "..", "..", "..", "docs", "openapi", name)
}

func assertJSONFixtureEqual(t *testing.T, path string, actual []byte) {
	t.Helper()

	expected, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}

	var want any
	if err := json.Unmarshal(expected, &want); err != nil {
		t.Fatalf("unmarshal expected fixture: %v", err)
	}
	var got any
	if err := json.Unmarshal(actual, &got); err != nil {
		t.Fatalf("unmarshal actual json: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("json mismatch\nactual=%s\nexpected=%s", actual, expected)
	}
}
