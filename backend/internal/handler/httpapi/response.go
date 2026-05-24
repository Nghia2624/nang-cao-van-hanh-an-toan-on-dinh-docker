package httpapi

import (
	"encoding/json"
	"net/http"
	"time"
)

type APIError struct {
	Error     string    `json:"error"`
	Message   string    `json:"message,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeErr(w http.ResponseWriter, status int, code string, msg string) {
	writeJSON(w, status, APIError{
		Error:     code,
		Message:   msg,
		Timestamp: time.Now().UTC(),
	})
}

