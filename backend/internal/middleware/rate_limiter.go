package middleware

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type limiterEntry struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

type limiterStore struct {
	mu       sync.Mutex
	limiters map[string]*limiterEntry
	r        rate.Limit
	b        int
}

func newLimiterStore(r rate.Limit, b int) *limiterStore {
	s := &limiterStore{
		limiters: make(map[string]*limiterEntry),
		r:        r,
		b:        b,
	}
	go s.cleanup()
	return s
}

func (s *limiterStore) get(ip string) *rate.Limiter {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e, ok := s.limiters[ip]; ok {
		e.lastUsed = time.Now()
		return e.limiter
	}
	l := rate.NewLimiter(s.r, s.b)
	s.limiters[ip] = &limiterEntry{limiter: l, lastUsed: time.Now()}
	return l
}

// cleanup removes stale IP entries every 5 minutes (idle > 10 min).
func (s *limiterStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		s.mu.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)
		for ip, e := range s.limiters {
			if e.lastUsed.Before(cutoff) {
				delete(s.limiters, ip)
			}
		}
		s.mu.Unlock()
	}
}

// RateLimit enforces per-IP rate limiting.
func RateLimit(r rate.Limit, b int) func(http.Handler) http.Handler {
	store := newLimiterStore(r, b)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			ip, _, err := net.SplitHostPort(req.RemoteAddr)
			if err != nil {
				ip = req.RemoteAddr
			}
			limiter := store.get(ip)
			if !limiter.Allow() {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"RATE_LIMIT","message":"Too many requests"}`))
				return
			}
			next.ServeHTTP(w, req)
		})
	}
}
