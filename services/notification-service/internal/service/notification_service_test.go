package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"tech-feed-hub/notification-service/internal/domain"
	"tech-feed-hub/notification-service/internal/events"
)

type stubNotificationRepo struct {
	createFunc           func(context.Context, domain.CreateNotificationInput) (bool, error)
	listFunc             func(context.Context, domain.ListNotificationsParams) (domain.ListNotificationsResult, error)
	updateReadStatusFunc func(context.Context, int64, bool) (*domain.Notification, error)
	getByIDFunc          func(context.Context, int64) (*domain.Notification, error)
}

func (r *stubNotificationRepo) Create(ctx context.Context, input domain.CreateNotificationInput) (bool, error) {
	return r.createFunc(ctx, input)
}

func (r *stubNotificationRepo) GetByID(ctx context.Context, id int64) (*domain.Notification, error) {
	if r.getByIDFunc != nil {
		return r.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (r *stubNotificationRepo) List(ctx context.Context, params domain.ListNotificationsParams) (domain.ListNotificationsResult, error) {
	if r.listFunc != nil {
		return r.listFunc(ctx, params)
	}
	return domain.ListNotificationsResult{}, nil
}

func (r *stubNotificationRepo) UpdateReadStatus(ctx context.Context, id int64, isRead bool) (*domain.Notification, error) {
	if r.updateReadStatusFunc != nil {
		return r.updateReadStatusFunc(ctx, id, isRead)
	}
	return nil, nil
}

func TestHandleArticleIngestedCreatesNotification(t *testing.T) {
	t.Parallel()

	service := &NotificationService{
		repo: &stubNotificationRepo{
			createFunc: func(_ context.Context, input domain.CreateNotificationInput) (bool, error) {
				if input.EventID != "evt-1" || input.EventType != events.EventTypeArticleIngested {
					t.Fatalf("unexpected event metadata: %+v", input)
				}
				if input.Level != "info" || input.Title != "Kubernetes Blog から新着 3 件" {
					t.Fatalf("unexpected title: %+v", input)
				}
				if input.Body != "最新記事: New release" {
					t.Fatalf("unexpected body: %s", input.Body)
				}
				if input.SourceID == nil || *input.SourceID != 7 || input.FetchJobID == nil || *input.FetchJobID != 11 {
					t.Fatalf("unexpected ids: %+v", input)
				}
				return true, nil
			},
		},
	}

	title := "New release"
	err := service.HandleArticleIngested(context.Background(), events.Envelope[events.ArticleIngestedPayload]{
		EventID:    "evt-1",
		EventType:  events.EventTypeArticleIngested,
		OccurredAt: time.Date(2026, 4, 16, 10, 0, 0, 0, time.UTC),
		Payload: events.ArticleIngestedPayload{
			JobID:               11,
			SourceID:            7,
			SourceName:          "Kubernetes Blog",
			InsertedCount:       3,
			RepresentativeTitle: &title,
		},
	})
	if err != nil {
		t.Fatalf("HandleArticleIngested returned error: %v", err)
	}
}

func TestHandleCollectorFetchFailedCreatesErrorNotification(t *testing.T) {
	t.Parallel()

	service := &NotificationService{
		repo: &stubNotificationRepo{
			createFunc: func(_ context.Context, input domain.CreateNotificationInput) (bool, error) {
				if input.Level != "error" {
					t.Fatalf("unexpected level: %s", input.Level)
				}
				if input.Title != "GitHub Changelog の取得に失敗" {
					t.Fatalf("unexpected title: %s", input.Title)
				}
				if input.Body != "fetch rss status=502" {
					t.Fatalf("unexpected body: %s", input.Body)
				}
				return true, nil
			},
		},
	}

	err := service.HandleCollectorFetchFailed(context.Background(), events.Envelope[events.CollectorFetchFailedPayload]{
		EventID:    "evt-2",
		EventType:  events.EventTypeCollectorFetchError,
		OccurredAt: time.Now().UTC(),
		Payload: events.CollectorFetchFailedPayload{
			JobID:        19,
			SourceID:     5,
			SourceName:   "GitHub Changelog",
			ErrorMessage: "fetch rss status=502",
		},
	})
	if err != nil {
		t.Fatalf("HandleCollectorFetchFailed returned error: %v", err)
	}
}

func TestListNotificationsRejectsLargePageSize(t *testing.T) {
	t.Parallel()

	service := &NotificationService{repo: &stubNotificationRepo{}}
	_, err := service.ListNotifications(context.Background(), domain.ListNotificationsParams{Page: 1, PageSize: 101})
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestUpdateReadStatusReturnsNotFound(t *testing.T) {
	t.Parallel()

	service := &NotificationService{
		repo: &stubNotificationRepo{
			updateReadStatusFunc: func(context.Context, int64, bool) (*domain.Notification, error) {
				return nil, nil
			},
		},
	}

	_, err := service.UpdateReadStatus(context.Background(), 9, UpdateReadStatusInput{IsRead: true})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}
