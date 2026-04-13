package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"tech-feed-hub/collector-service/internal/collector"
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

	router := gin.Default()
	httpapi.RegisterRoutes(router, collector.NewService(articleServiceURL, sources))

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
