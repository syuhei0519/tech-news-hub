package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"tech-feed-hub/collector-service/internal/collector"
	"tech-feed-hub/collector-service/internal/events"
	"tech-feed-hub/collector-service/internal/httpapi"
)

func Run() error {
	port := getenv("PORT", "8082")
	articleServiceURL := os.Getenv("ARTICLE_SERVICE_URL")
	if articleServiceURL == "" {
		return fmt.Errorf("ARTICLE_SERVICE_URL is required")
	}

	rawSources := os.Getenv("COLLECTOR_SOURCES_JSON")
	if rawSources == "" {
		return fmt.Errorf("COLLECTOR_SOURCES_JSON is required")
	}

	var sources []collector.SourceConfig
	if err := json.Unmarshal([]byte(rawSources), &sources); err != nil {
		return fmt.Errorf("parse COLLECTOR_SOURCES_JSON: %w", err)
	}

	service := collector.NewService(articleServiceURL, sources)
	if amqpURL := os.Getenv("AMQP_URL"); amqpURL != "" {
		publisher, err := events.NewRabbitMQPublisher(amqpURL, getenv("AMQP_EXCHANGE", "tech-feed.events"))
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := publisher.Close(); closeErr != nil {
				log.Printf("close collector-service publisher: %v", closeErr)
			}
		}()
		service.SetEventPublisher(publisher)
	}

	router := gin.Default()
	httpapi.RegisterRoutes(router, service)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}
	return server.ListenAndServe()
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
