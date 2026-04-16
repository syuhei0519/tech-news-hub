package app

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"tech-feed-hub/article-service/internal/events"
	"tech-feed-hub/article-service/internal/httpapi"
	"tech-feed-hub/article-service/internal/repository"
	"tech-feed-hub/article-service/internal/service"
)

func Run() error {
	port := getenv("PORT", "8081")
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		return fmt.Errorf("MYSQL_DSN is required")
	}

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}

	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	articleRepo := repository.NewArticleRepository(db)
	sourceRepo := repository.NewSourceRepository(db)
	jobRepo := repository.NewFetchJobRepository(db)
	articleService := service.NewArticleService(articleRepo, sourceRepo, jobRepo)

	if amqpURL := os.Getenv("AMQP_URL"); amqpURL != "" {
		publisher, err := events.NewRabbitMQPublisher(amqpURL, getenv("AMQP_EXCHANGE", "tech-feed.events"))
		if err != nil {
			return err
		}
		defer func() {
			if closeErr := publisher.Close(); closeErr != nil {
				log.Printf("close article-service publisher: %v", closeErr)
			}
		}()
		articleService.SetNotificationPublisher(publisher)
	}

	router := httpapi.NewRouter(db, articleService)

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("article-service listening on :%s", port)
	return server.ListenAndServe()
}

func getenv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
