package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	connection *amqp.Connection
	channel    *amqp.Channel
	exchange   string
	queueName  string
	handler    Handler
}

func NewConsumer(amqpURL string, exchange string, queueName string, handler Handler) (*Consumer, error) {
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

	queue, err := channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return nil, fmt.Errorf("declare queue: %w", err)
	}

	for _, routingKey := range []string{EventTypeArticleIngested, EventTypeCollectorFetchError} {
		if err := channel.QueueBind(queue.Name, routingKey, exchange, false, nil); err != nil {
			_ = channel.Close()
			_ = connection.Close()
			return nil, fmt.Errorf("bind queue %s: %w", routingKey, err)
		}
	}

	return &Consumer{
		connection: connection,
		channel:    channel,
		exchange:   exchange,
		queueName:  queue.Name,
		handler:    handler,
	}, nil
}

func (c *Consumer) Close() error {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.connection != nil {
		return c.connection.Close()
	}
	return nil
}

func (c *Consumer) Consume(ctx context.Context) error {
	deliveries, err := c.channel.Consume(c.queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("register consumer: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case delivery, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("rabbitmq delivery channel closed")
			}
			if err := c.handleDelivery(ctx, delivery); err != nil {
				log.Printf("notification consumer failed: routing_key=%s err=%v", delivery.RoutingKey, err)
				if nackErr := delivery.Nack(false, true); nackErr != nil {
					return fmt.Errorf("nack delivery: %w", nackErr)
				}
				continue
			}
			if err := delivery.Ack(false); err != nil {
				return fmt.Errorf("ack delivery: %w", err)
			}
		}
	}
}

func (c *Consumer) handleDelivery(ctx context.Context, delivery amqp.Delivery) error {
	var header struct {
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(delivery.Body, &header); err != nil {
		return fmt.Errorf("decode event header: %w", err)
	}

	switch header.EventType {
	case EventTypeArticleIngested:
		var event Envelope[ArticleIngestedPayload]
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			return fmt.Errorf("decode article event: %w", err)
		}
		return c.handler.HandleArticleIngested(ctx, event)
	case EventTypeCollectorFetchError:
		var event Envelope[CollectorFetchFailedPayload]
		if err := json.Unmarshal(delivery.Body, &event); err != nil {
			return fmt.Errorf("decode collector event: %w", err)
		}
		return c.handler.HandleCollectorFetchFailed(ctx, event)
	default:
		log.Printf("notification consumer ignored unknown event type: %s", header.EventType)
		return nil
	}
}
