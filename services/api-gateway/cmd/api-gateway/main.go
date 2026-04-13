package main

import (
	"log"

	"tech-feed-hub/api-gateway/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
