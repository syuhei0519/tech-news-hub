package events

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const EventTypeArticleIngested = "article.ingested"

type ArticleIngestedPayload struct {
	JobID               int64   `json:"job_id"`
	SourceID            int64   `json:"source_id"`
	SourceName          string  `json:"source_name"`
	InsertedCount       int     `json:"inserted_count"`
	RepresentativeTitle *string `json:"representative_title,omitempty"`
}

type NotificationPublisher interface {
	PublishArticleIngested(ctx context.Context, payload ArticleIngestedPayload) error
}

type NopPublisher struct{}

func (NopPublisher) PublishArticleIngested(context.Context, ArticleIngestedPayload) error {
	return nil
}

type RabbitMQPublisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	exchange   string
}

func NewRabbitMQPublisher(amqpURL string, exchange string) (*RabbitMQPublisher, error) {
	connection, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("dial rabbitmq: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("open rabbitmq channel: %w", err)
	}

	if err := channel.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return nil, fmt.Errorf("declare exchange: %w", err)
	}

	return &RabbitMQPublisher{
		connection: connection,
		channel:    channel,
		exchange:   exchange,
	}, nil
}

func (p *RabbitMQPublisher) Close() error {
	if p.channel != nil {
		_ = p.channel.Close()
	}
	if p.connection != nil {
		return p.connection.Close()
	}
	return nil
}

func (p *RabbitMQPublisher) PublishArticleIngested(ctx context.Context, payload ArticleIngestedPayload) error {
	body, err := json.Marshal(struct {
		EventID    string                 `json:"event_id"`
		EventType  string                 `json:"event_type"`
		OccurredAt time.Time              `json:"occurred_at"`
		Source     map[string]string      `json:"source"`
		Payload    ArticleIngestedPayload `json:"payload"`
	}{
		EventID:    newEventID(),
		EventType:  EventTypeArticleIngested,
		OccurredAt: time.Now().UTC(),
		Source:     map[string]string{"service": "article-service"},
		Payload:    payload,
	})
	if err != nil {
		return fmt.Errorf("marshal article event: %w", err)
	}

	return p.channel.PublishWithContext(ctx, p.exchange, EventTypeArticleIngested, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now().UTC(),
		Body:         body,
	})
}

func newEventID() string {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return fmt.Sprintf("evt-%d", time.Now().UTC().UnixNano())
	}
	return hex.EncodeToString(value)
}
