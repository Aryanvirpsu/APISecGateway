package middleware

import (
	"context"
	"log"
	"net/http"
	"time"

	"gateway/internal/store"

	"github.com/google/uuid"
)

// LogStore persists request logs.
type LogStore interface {
	InsertRequestLog(ctx context.Context, in store.LogInput) error
}

// RequestLogger assigns a request id, runs the rest of the chain and then
// records the outcome. It must sit near the top of the chain so the shared
// State exists for downstream middleware to enrich.
func RequestLogger(s LogStore, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		st := &State{RequestID: uuid.New()}
		r = withState(r, st)
		w.Header().Set("X-Request-Id", st.RequestID.String())

		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		ip := clientIP(r)

		// Capture the original request line up front: downstream handlers (the
		// proxy) rewrite r.URL.Path before this runs to completion.
		method, path := r.Method, r.URL.Path

		next.ServeHTTP(sw, r)

		err := s.InsertRequestLog(r.Context(), store.LogInput{
			RequestID: st.RequestID,
			SourceIP:  ip,
			Method:    method,
			Path:      path,
			Status:    sw.status,
			LatencyMS: int(time.Since(start).Milliseconds()),
			UserAgent: r.UserAgent(),
			AuthSub:   st.AuthSubject,
			TokenID:   st.TokenID,
		})
		if err != nil {
			log.Printf("request log insert failed: %v", err)
		}
	})
}
