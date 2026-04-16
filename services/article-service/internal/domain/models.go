package domain

import "time"

type Article struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	URL         string     `json:"url"`
	SourceID    int64      `json:"source_id"`
	SourceName  string     `json:"source_name,omitempty"`
	PublishedAt *time.Time `json:"published_at"`
	FetchedAt   time.Time  `json:"fetched_at"`
	Excerpt     string     `json:"excerpt"`
	Category    string     `json:"category"`
	Tags        []string   `json:"tags"`
	IsRead      bool       `json:"is_read"`
	IsFavorite  bool       `json:"is_favorite"`
	DedupeKey   string     `json:"dedupe_key,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Source struct {
	ID              int64      `json:"id"`
	Name            string     `json:"name"`
	Type            string     `json:"type"`
	FetchURL        string     `json:"fetch_url"`
	FetchMethod     string     `json:"fetch_method"`
	IntervalMinutes int        `json:"interval_minutes"`
	DefaultCategory string     `json:"default_category"`
	IsEnabled       bool       `json:"is_enabled"`
	LastFetchedAt   *time.Time `json:"last_fetched_at"`
	LastFetchStatus *string    `json:"last_fetch_status"`
	LastErrorMsg    *string    `json:"last_error_message"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type FetchJob struct {
	ID              int64      `json:"id"`
	SourceID        int64      `json:"source_id"`
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at"`
	Status          string     `json:"status"`
	FetchedCount    int        `json:"fetched_count"`
	InsertedCount   int        `json:"inserted_count"`
	DuplicatedCount int        `json:"duplicated_count"`
	ErrorMessage    *string    `json:"error_message"`
}

type ListFetchJobsParams struct {
	SourceID int64
	Status   string
	Page     int
	PageSize int
}

type ListFetchJobsResult struct {
	Items      []FetchJob `json:"items"`
	Total      int        `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

type ListArticlesParams struct {
	Query    string
	Category string
	SourceID int64
	Page     int
	PageSize int
	Sort     string
	Order    string
}

type ListArticlesResult struct {
	Items      []Article `json:"items"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}
