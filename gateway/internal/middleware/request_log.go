package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type LogInput struct {
	RequestID uuid.UUID
	SourceIP  string
	Method    string
	Path      string
	Status    int
	LatencyMS int
	UserAgent string
	AuthSub   *string
	TokenID   *string
}

type LogStore interface {
	InsertRequestLog(ctx context.Context, in LogInput) error
}

func RequestLogger(store LogStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		sw := &statusWriter{ResponseWriter: w, status: 200}

		reqID := uuid.New()

		// best-effort IP (later we’ll handle X-Forwarded-For properly)
		sourceIP := r.RemoteAddr

		next.ServeHTTP(sw, r)

		latency := time.Since(start).Milliseconds()

		_ = store.InsertRequestLog(r.Context(), LogInput{
			RequestID: reqID,
			SourceIP:  sourceIP,
			Method:    r.Method,
			Path:      r.URL.Path,
			Status:    sw.status,
			LatencyMS: int(latency),
			UserAgent: r.UserAgent(),
			AuthSub:   nil,
			TokenID:   nil,
		})
	})
}
