package main

import (
	"fmt"
	"net/http"

	"DEEPSYSTEM/sentinal/api"
	"DEEPSYSTEM/sentinal/internal/ratelimit"
)

func main() {
	// capacity=5, refillRate=1 token/sec -> burst of 5, then 1 req/sec sustained.
	registry := ratelimit.NewRegistry(5, 1)
	mux := api.NewMux(registry)

	fmt.Println("Sentinel listening on :8080  (try: curl 'localhost:8080/check?key=user1')")
	http.ListenAndServe(":8080", mux)
}