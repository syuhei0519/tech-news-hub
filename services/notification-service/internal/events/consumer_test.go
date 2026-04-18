package events

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

type stubHandler struct {
	handleArticleIngestedFunc      func(context.Context, Envelope[ArticleIngestedPayload]) error
	handleCollectorFetchFailedFunc func(context.Context, Envelope[CollectorFetchFailedPayload]) error
}

func (h *stubHandler) HandleArticleIngested(ctx context.Context, event Envelope[ArticleIngestedPayload]) error {
	if h.handleArticleIngestedFunc != nil {
		return h.handleArticleIngestedFunc(ctx, event)
	}
	return nil
}

func (h *stubHandler) HandleCollectorFetchFailed(ctx context.Context, event Envelope[CollectorFetchFailedPayload]) error {
	if h.handleCollectorFetchFailedFunc != nil {
		return h.handleCollectorFetchFailedFunc(ctx, event)
	}
	return nil
}

func TestHandleDeliveryDispatchesArticleIngestedFixture(t *testing.T) {
	t.Parallel()

	// producer 側 fixture をそのまま読めることを確認し、event 契約の片側変更を consumer テストで拾う。
	var handled bool
	consumer := &Consumer{
		handler: &stubHandler{
			handleArticleIngestedFunc: func(_ context.Context, event Envelope[ArticleIngestedPayload]) error {
				handled = true
				if event.EventType != EventTypeArticleIngested || event.Payload.JobID != 42 || event.Payload.InsertedCount != 3 {
					t.Fatalf("unexpected event: %+v", event)
				}
				return nil
			},
		},
	}

	if err := consumer.handleDelivery(context.Background(), amqp.Delivery{
		RoutingKey: EventTypeArticleIngested,
		Body:       mustReadFixture(t, "article.ingested.json"),
	}); err != nil {
		t.Fatalf("handleDelivery returned error: %v", err)
	}
	if !handled {
		t.Fatal("expected article handler to be called")
	}
}

func TestHandleDeliveryDispatchesCollectorFetchFailedFixture(t *testing.T) {
	t.Parallel()

	// collector.fetch.failed も同じ fixture 基準で decode し、routing key ごとの dispatch を固定する。
	var handled bool
	consumer := &Consumer{
		handler: &stubHandler{
			handleCollectorFetchFailedFunc: func(_ context.Context, event Envelope[CollectorFetchFailedPayload]) error {
				handled = true
				if event.EventType != EventTypeCollectorFetchError || event.Payload.SourceID != 7 || event.Payload.ErrorMessage != "fetch rss status=502" {
					t.Fatalf("unexpected event: %+v", event)
				}
				return nil
			},
		},
	}

	if err := consumer.handleDelivery(context.Background(), amqp.Delivery{
		RoutingKey: EventTypeCollectorFetchError,
		Body:       mustReadFixture(t, "collector.fetch.failed.json"),
	}); err != nil {
		t.Fatalf("handleDelivery returned error: %v", err)
	}
	if !handled {
		t.Fatal("expected collector handler to be called")
	}
}

func TestHandleDeliveryIgnoresUnknownEventType(t *testing.T) {
	t.Parallel()

	// 未対応 event を poison message にしないよう、unknown type は ignore する現在方針を固定する。
	consumer := &Consumer{handler: &stubHandler{}}
	body := []byte(`{"event_type":"unknown.event","payload":{"value":1}}`)
	if err := consumer.handleDelivery(context.Background(), amqp.Delivery{RoutingKey: "unknown.event", Body: body}); err != nil {
		t.Fatalf("expected unknown event to be ignored, got %v", err)
	}
}

func TestHandleDeliveryReturnsDecodeErrorForMalformedJSON(t *testing.T) {
	t.Parallel()

	// 壊れた JSON は retry 可能な失敗として返し、呼び出し側の Nack(requeue) に流せるようにする。
	consumer := &Consumer{handler: &stubHandler{}}
	err := consumer.handleDelivery(context.Background(), amqp.Delivery{RoutingKey: EventTypeArticleIngested, Body: []byte(`{"event_type":`)})
	if err == nil || !strings.Contains(err.Error(), "decode event header") {
		t.Fatalf("expected header decode error, got %v", err)
	}
}

func TestHandleDeliveryReturnsDecodeErrorForInvalidPayloadShape(t *testing.T) {
	t.Parallel()

	// header だけ読めても payload shape が崩れていれば処理しないことを契約テストで押さえる。
	consumer := &Consumer{handler: &stubHandler{}}
	body := mustReadFixture(t, "article.ingested.json")

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	payload["payload"] = "invalid"
	invalid, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal invalid payload: %v", err)
	}

	err = consumer.handleDelivery(context.Background(), amqp.Delivery{RoutingKey: EventTypeArticleIngested, Body: invalid})
	if err == nil || !strings.Contains(err.Error(), "decode article event") {
		t.Fatalf("expected payload decode error, got %v", err)
	}
}

func TestHandleDeliveryReturnsHandlerError(t *testing.T) {
	t.Parallel()

	// decode 後の保存失敗も呼び出し元に返し、Ack して取りこぼさないことを確認する。
	consumer := &Consumer{
		handler: &stubHandler{
			handleCollectorFetchFailedFunc: func(context.Context, Envelope[CollectorFetchFailedPayload]) error {
				return errors.New("repository down")
			},
		},
	}

	err := consumer.handleDelivery(context.Background(), amqp.Delivery{
		RoutingKey: EventTypeCollectorFetchError,
		Body:       mustReadFixture(t, "collector.fetch.failed.json"),
	})
	if err == nil || !strings.Contains(err.Error(), "repository down") {
		t.Fatalf("expected handler error, got %v", err)
	}
}

func mustReadFixture(t *testing.T, name string) []byte {
	t.Helper()

	body, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "docs", "openapi", name))
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return body
}
