package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"

	"tech-feed-hub/notification-service/internal/events"
	"tech-feed-hub/notification-service/internal/httpapi"
	"tech-feed-hub/notification-service/internal/repository"
	"tech-feed-hub/notification-service/internal/service"
)

func Run() error {
	port := getenv("PORT", "8083")
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		return fmt.Errorf("MYSQL_DSN is required")
	}

	amqpURL := os.Getenv("AMQP_URL")
	if amqpURL == "" {
		return fmt.Errorf("AMQP_URL is required")
	}

	exchange := getenv("AMQP_EXCHANGE", "tech-feed.events")
	queueName := getenv("AMQP_QUEUE_NAME", "notification-service")

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

	notificationRepo := repository.NewNotificationRepository(db)
	notificationService := service.NewNotificationService(notificationRepo)
	router := httpapi.NewRouter(db, notificationService)

	consumer, err := events.NewConsumer(amqpURL, exchange, queueName, notificationService)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := consumer.Close(); closeErr != nil {
			log.Printf("close notification consumer: %v", closeErr)
		}
	}()

	go func() {
		if err := consumer.Consume(context.Background()); err != nil {
			log.Printf("notification consumer stopped: %v", err)
		}
	}()

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("notification-service listening on :%s", port)
	return server.ListenAndServe()
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
