package events

import (
	"context"
	"time"
)

const (
	EventTypeArticleIngested     = "article.ingested"
	EventTypeCollectorFetchError = "collector.fetch.failed"
)

type Envelope[T any] struct {
	EventID    string      `json:"event_id"`
	EventType  string      `json:"event_type"`
	OccurredAt time.Time   `json:"occurred_at"`
	Source     EventSource `json:"source"`
	Payload    T           `json:"payload"`
}

type EventSource struct {
	Service string `json:"service"`
}

type ArticleIngestedPayload struct {
	JobID               int64   `json:"job_id"`
	SourceID            int64   `json:"source_id"`
	SourceName          string  `json:"source_name"`
	InsertedCount       int     `json:"inserted_count"`
	RepresentativeTitle *string `json:"representative_title,omitempty"`
}

type CollectorFetchFailedPayload struct {
	JobID        int64  `json:"job_id"`
	SourceID     int64  `json:"source_id"`
	SourceName   string `json:"source_name"`
	ErrorMessage string `json:"error_message"`
}

type Handler interface {
	HandleArticleIngested(ctx context.Context, event Envelope[ArticleIngestedPayload]) error
	HandleCollectorFetchFailed(ctx context.Context, event Envelope[CollectorFetchFailedPayload]) error
}
