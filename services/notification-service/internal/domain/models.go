package domain

import "time"

type Notification struct {
	ID         int64      `json:"id"`
	EventID    string     `json:"event_id"`
	EventType  string     `json:"event_type"`
	Level      string     `json:"level"`
	Title      string     `json:"title"`
	Body       string     `json:"body"`
	SourceID   *int64     `json:"source_id"`
	FetchJobID *int64     `json:"fetch_job_id"`
	IsRead     bool       `json:"is_read"`
	CreatedAt  time.Time  `json:"created_at"`
	ReadAt     *time.Time `json:"read_at"`
}

type CreateNotificationInput struct {
	EventID    string
	EventType  string
	Level      string
	Title      string
	Body       string
	SourceID   *int64
	FetchJobID *int64
	CreatedAt  time.Time
}

type ListNotificationsParams struct {
	IsRead   *bool
	Page     int
	PageSize int
}

type ListNotificationsResult struct {
	Items      []Notification `json:"items"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}
