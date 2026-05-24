package middleware

import "net/http"

// APIKeyAuth guards endpoints with a static API key header.
func APIKeyAuth(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}
			// EventSource cannot set headers reliably across browsers; allow query param for SSE.
			provided := r.Header.Get("X-API-Key")
			if provided == "" {
				provided = r.URL.Query().Get("apiKey")
			}
			if provided == "" || provided != key {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"UNAUTHORIZED","message":"Invalid or missing API key"}`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
