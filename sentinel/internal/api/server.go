package api

import (
	"encoding/json"
	"net/http"

	"sentinel/internal/ratelimit"
	"sentinel/internal/stats"
)

// Server wires together the rate limiter and the stats counter behind
// HTTP handlers. This is the ONLY package that imports net/http —
// keeping HTTP concerns separate from the core algorithms means we can
// swap transports later (Day 10: gRPC) without touching ratelimit or stats.
type Server struct {
	registry *ratelimit.Registry
	counter  *stats.ExactCounter
}

func NewServer(registry *ratelimit.Registry, counter *stats.ExactCounter) *Server {
	return &Server{
		registry: registry,
		counter:  counter,
	}
}

// Routes returns an http.Handler with all routes registered.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/check", s.handleCheck)
	mux.HandleFunc("/stats", s.handleStats)
	return mux
}

// handleCheck answers "is this request allowed?" for a given key
// (e.g. ?key=user-123) and records the hit in the exact counter.
func (s *Server) handleCheck(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "missing 'key' query parameter", http.StatusBadRequest)
		return
	}

	clientIP := r.RemoteAddr
	s.counter.Increment("endpoint:/check:"+key, clientIP)

	allowed := s.registry.Allow(key)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"allowed": allowed})
}

// handleStats returns the exact hit counts collected so far.
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(s.counter.Snapshot())
}