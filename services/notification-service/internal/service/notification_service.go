package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"tech-feed-hub/notification-service/internal/domain"
	"tech-feed-hub/notification-service/internal/events"
	"tech-feed-hub/notification-service/internal/repository"
)

var (
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
)

type serviceError struct {
	kind    error
	message string
}

func (e *serviceError) Error() string {
	return e.message
}

func (e *serviceError) Unwrap() error {
	return e.kind
}

type notificationRepository interface {
	Create(ctx context.Context, input domain.CreateNotificationInput) (bool, error)
	GetByID(ctx context.Context, id int64) (*domain.Notification, error)
	List(ctx context.Context, params domain.ListNotificationsParams) (domain.ListNotificationsResult, error)
	UpdateReadStatus(ctx context.Context, id int64, isRead bool) (*domain.Notification, error)
}

type NotificationService struct {
	repo notificationRepository
}

func NewNotificationService(repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) ListNotifications(ctx context.Context, params domain.ListNotificationsParams) (domain.ListNotificationsResult, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PageSize < 1 {
		params.PageSize = 20
	}
	if params.PageSize > 100 {
		return domain.ListNotificationsResult{}, newServiceError(ErrValidation, "page_size must be less than or equal to 100")
	}
	return s.repo.List(ctx, params)
}

type UpdateReadStatusInput struct {
	IsRead bool `json:"is_read"`
}

func (s *NotificationService) UpdateReadStatus(ctx context.Context, id int64, input UpdateReadStatusInput) (*domain.Notification, error) {
	if id < 1 {
		return nil, newServiceError(ErrValidation, "notification id is required")
	}

	updated, err := s.repo.UpdateReadStatus(ctx, id, input.IsRead)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, newServiceError(ErrNotFound, "notification not found")
	}
	return updated, nil
}

func (s *NotificationService) HandleArticleIngested(ctx context.Context, event events.Envelope[events.ArticleIngestedPayload]) error {
	title := fmt.Sprintf("%s から新着 %d 件", event.Payload.SourceName, event.Payload.InsertedCount)
	body := "新着記事を確認してください。"
	if event.Payload.RepresentativeTitle != nil && strings.TrimSpace(*event.Payload.RepresentativeTitle) != "" {
		body = fmt.Sprintf("最新記事: %s", strings.TrimSpace(*event.Payload.RepresentativeTitle))
	}

	_, err := s.repo.Create(ctx, domain.CreateNotificationInput{
		EventID:    event.EventID,
		EventType:  event.EventType,
		Level:      "info",
		Title:      title,
		Body:       body,
		SourceID:   int64Ptr(event.Payload.SourceID),
		FetchJobID: int64Ptr(event.Payload.JobID),
		CreatedAt:  coalesceEventTime(event.OccurredAt),
	})
	return err
}

func (s *NotificationService) HandleCollectorFetchFailed(ctx context.Context, event events.Envelope[events.CollectorFetchFailedPayload]) error {
	_, err := s.repo.Create(ctx, domain.CreateNotificationInput{
		EventID:    event.EventID,
		EventType:  event.EventType,
		Level:      "error",
		Title:      fmt.Sprintf("%s の取得に失敗", event.Payload.SourceName),
		Body:       strings.TrimSpace(event.Payload.ErrorMessage),
		SourceID:   int64Ptr(event.Payload.SourceID),
		FetchJobID: int64Ptr(event.Payload.JobID),
		CreatedAt:  coalesceEventTime(event.OccurredAt),
	})
	return err
}

func coalesceEventTime(value time.Time) time.Time {
	if value.IsZero() {
		return time.Now().UTC()
	}
	return value.UTC()
}

func int64Ptr(value int64) *int64 {
	return &value
}

func newServiceError(kind error, message string) error {
	return &serviceError{kind: kind, message: message}
}
