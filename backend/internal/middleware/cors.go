package middleware

import (
	"net/http"
	"strings"
)

// CORS sets Access-Control headers. If allowedOrigins is empty, it allows "*" (dev mode).
// In production, pass explicit origins like "https://myapp.example.com".
func CORSWithOrigins(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := "*"
			if len(allowedOrigins) > 0 && origin != "" {
				for _, o := range allowedOrigins {
					if strings.EqualFold(o, origin) || o == "*" {
						allowed = origin
						break
					}
				}
				// If no match and we have specific origins, don't set the header
				if allowed == "*" && allowedOrigins[0] != "*" {
					allowed = ""
				}
			}
			if allowed != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowed)
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")
			w.Header().Set("Access-Control-Max-Age", "3600")
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// CORS is a backward-compatible permissive CORS handler (for dev use).
func CORS(next http.Handler) http.Handler {
	return CORSWithOrigins([]string{"*"})(next)
}
