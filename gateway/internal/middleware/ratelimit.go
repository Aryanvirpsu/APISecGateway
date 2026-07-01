package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"gateway/internal/store"

	"github.com/google/uuid"
)

// AlertStore records security events.
type AlertStore interface {
	InsertAlert(ctx context.Context, in store.AlertInput) error
}

// IPBlocker persists blocked IPs.
type IPBlocker interface {
	BlockIP(ctx context.Context, ip, reason string) error
}

type bucket struct {
	tokens     float64
	last       time.Time
	violations int
}

// RateLimiter applies a per-IP token bucket. Each rejection raises a rate_limit
// alert, and once an IP exceeds the configured rejection budget it is promoted
// to a hard block.
type RateLimiter struct {
	rate    float64 // tokens added per second
	burst   float64 // bucket capacity
	blockAt int     // violations before an IP is blocked

	alerts  AlertStore
	blocker IPBlocker
	blocked *StringSet

	mu      sync.Mutex
	buckets map[string]*bucket
}

// NewRateLimiter constructs a limiter.
func NewRateLimiter(rate float64, burst, blockAfter int, alerts AlertStore, blocker IPBlocker, blocked *StringSet) *RateLimiter {
	return &RateLimiter{
		rate:    rate,
		burst:   float64(burst),
		blockAt: blockAfter,
		alerts:  alerts,
		blocker: blocker,
		blocked: blocked,
		buckets: make(map[string]*bucket),
	}
}

// allow applies the token-bucket algorithm, returning whether the request is
// permitted and the running violation count for the IP.
func (rl *RateLimiter) allow(ip string) (bool, int) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b := rl.buckets[ip]
	if b == nil {
		b = &bucket{tokens: rl.burst, last: now}
		rl.buckets[ip] = b
	}

	// Refill proportionally to elapsed time, capped at the burst size.
	b.tokens += now.Sub(b.last).Seconds() * rl.rate
	if b.tokens > rl.burst {
		b.tokens = rl.burst
	}
	b.last = now

	if b.tokens >= 1 {
		b.tokens--
		return true, b.violations
	}
	b.violations++
	return false, b.violations
}

// Handler wraps next with rate limiting.
func (rl *RateLimiter) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := clientIP(r)
		ok, violations := rl.allow(ip)
		if ok {
			next.ServeHTTP(w, r)
			return
		}

		var reqID *uuid.UUID
		var subject *string
		if st := stateFrom(r); st != nil {
			reqID = &st.RequestID
			subject = st.AuthSubject
		}

		meta, _ := json.Marshal(map[string]int{"violations": violations})
		_ = rl.alerts.InsertAlert(r.Context(), store.AlertInput{
			RequestID: reqID,
			SourceIP:  ip,
			AuthSub:   subject,
			AlertType: "rate_limit",
			Severity:  2,
			Reason:    "request rate exceeded",
			Metadata:  meta,
		})

		// Sustained abuse becomes a hard block.
		if violations >= rl.blockAt && !rl.blocked.Has(ip) {
			rl.blocked.Add(ip)
			if err := rl.blocker.BlockIP(r.Context(), ip, "rate limit abuse"); err == nil {
				_ = rl.alerts.InsertAlert(r.Context(), store.AlertInput{
					RequestID: reqID,
					SourceIP:  ip,
					AuthSub:   subject,
					AlertType: "ip_blocked",
					Severity:  3,
					Reason:    "blocked after repeated rate limit violations",
				})
			}
		}

		w.Header().Set("Retry-After", "1")
		http.Error(w, "too many requests", http.StatusTooManyRequests)
	})
}

// StartCleanup periodically reaps idle buckets to bound memory use.
func (rl *RateLimiter) StartCleanup(every time.Duration) {
	go func() {
		ticker := time.NewTicker(every)
		defer ticker.Stop()
		for range ticker.C {
			cutoff := time.Now().Add(-every)
			rl.mu.Lock()
			for ip, b := range rl.buckets {
				if b.violations == 0 && b.last.Before(cutoff) {
					delete(rl.buckets, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
}
