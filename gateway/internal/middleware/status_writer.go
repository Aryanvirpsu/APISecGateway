package middleware

import "net/http"

// statusWriter lets us capture the final HTTP status code the handler writes.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
