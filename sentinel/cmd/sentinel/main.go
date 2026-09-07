package main

import (
	"log"
	"net/http"

	"sentinel/internal/api"
	"sentinel/internal/ratelimit"
	"sentinel/internal/stats"
)

func main() {
	// defaultCapacity=10, defaultRefillRate=2 tokens/sec — tune later.
	registry := ratelimit.NewRegistry(10, 2)
	counter := stats.NewExactCounter()

	server := api.NewServer(registry, counter)

	log.Println("sentinel listening on :8080")
	if err := http.ListenAndServe(":8080", server.Routes()); err != nil {
		log.Fatal(err)
	}
}