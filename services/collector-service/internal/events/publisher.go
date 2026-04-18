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

const EventTypeCollectorFetchFailed = "collector.fetch.failed"

type FetchFailedPayload struct {
	JobID        int64  `json:"job_id"`
	SourceID     int64  `json:"source_id"`
	SourceName   string `json:"source_name"`
	ErrorMessage string `json:"error_message"`
}

type EventPublisher interface {
	PublishFetchFailed(ctx context.Context, payload FetchFailedPayload) error
}

type NopPublisher struct{}

func (NopPublisher) PublishFetchFailed(context.Context, FetchFailedPayload) error {
	return nil
}

type RabbitMQPublisher struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	exchange   string
}

type eventEnvelope struct {
	EventID    string            `json:"event_id"`
	EventType  string            `json:"event_type"`
	OccurredAt time.Time         `json:"occurred_at"`
	Source     map[string]string `json:"source"`
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

func (p *RabbitMQPublisher) PublishFetchFailed(ctx context.Context, payload FetchFailedPayload) error {
	body, err := marshalFetchFailedEvent(newEventID(), time.Now().UTC(), payload)
	if err != nil {
		return fmt.Errorf("marshal collector event: %w", err)
	}

	return p.channel.PublishWithContext(ctx, p.exchange, EventTypeCollectorFetchFailed, false, false, amqp.Publishing{
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

func marshalFetchFailedEvent(eventID string, occurredAt time.Time, payload FetchFailedPayload) ([]byte, error) {
	return json.Marshal(struct {
		eventEnvelope
		Payload FetchFailedPayload `json:"payload"`
	}{
		eventEnvelope: eventEnvelope{
			EventID:    eventID,
			EventType:  EventTypeCollectorFetchFailed,
			OccurredAt: occurredAt.UTC(),
			Source:     map[string]string{"service": "collector-service"},
		},
		Payload: payload,
	})
}
