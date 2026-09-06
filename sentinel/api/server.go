package api

import (
	"encoding/json"
	"net/http"

	"DEEPSYSTEM/sentinal/internal/ratelimit"
)

type checkResponse struct {
	Key     string `json:"key"`
	Allowed bool   `json:"allowed"`
}

// NewMux builds the HTTP handler for Sentinel. Keeping this separate
// from main.go means new endpoints (e.g. /stats on Day 2, /metrics on
// Day 13) get added here without touching the entrypoint.
func NewMux(reg *ratelimit.Registry) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/check", func(w http.ResponseWriter, req *http.Request) {
		key := req.URL.Query().Get("key")
		if key == "" {
			http.Error(w, "missing ?key=", http.StatusBadRequest)
			return
		}

		allowed := reg.Allow(key)

		w.Header().Set("Content-Type", "application/json")
		if !allowed {
			w.WriteHeader(http.StatusTooManyRequests)
		}
		json.NewEncoder(w).Encode(checkResponse{Key: key, Allowed: allowed})
	})

	return mux
}