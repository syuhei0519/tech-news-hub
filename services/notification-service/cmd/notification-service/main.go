package main

import (
	"log"

	"tech-feed-hub/notification-service/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
