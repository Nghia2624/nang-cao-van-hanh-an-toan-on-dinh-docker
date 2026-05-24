package middleware

import (
	"net/http"
	"time"

	"github.com/nghia/dockerai/backend/internal/pkg/logger"
)

// RequestLogger logs basic request info.
func RequestLogger(log *logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			log.Infow("http request",
				"method", r.Method,
				"path", r.URL.Path,
				"remote", r.RemoteAddr,
				"duration_ms", time.Since(start).Milliseconds(),
			)
		})
	}
}
