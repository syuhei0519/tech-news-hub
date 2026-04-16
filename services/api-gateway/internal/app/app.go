package app

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"

	"tech-feed-hub/api-gateway/internal/httpapi"
)

func Run() error {
	port := getenv("PORT", "8080")
	articleServiceURL := os.Getenv("ARTICLE_SERVICE_URL")
	if articleServiceURL == "" {
		return fmt.Errorf("ARTICLE_SERVICE_URL is required")
	}
	notificationServiceURL := os.Getenv("NOTIFICATION_SERVICE_URL")
	if notificationServiceURL == "" {
		return fmt.Errorf("NOTIFICATION_SERVICE_URL is required")
	}

	router := gin.Default()
	router.Use(httpapi.RequestLogger())
	httpapi.RegisterRoutes(router, articleServiceURL, notificationServiceURL)

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
