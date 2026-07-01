package middleware

import (
	"net/http"
	"sync"
)

// StringSet is a concurrency-safe set of strings. It backs both the blocked-IP
// cache and the revoked-token cache, seeded from the database at start-up and
// updated as the gateway blocks abusive clients.
type StringSet struct {
	mu sync.RWMutex
	m  map[string]struct{}
}

// NewStringSet builds a set pre-populated with seed.
func NewStringSet(seed []string) *StringSet {
	s := &StringSet{m: make(map[string]struct{}, len(seed))}
	for _, v := range seed {
		s.m[v] = struct{}{}
	}
	return s
}

// Add inserts a value.
func (s *StringSet) Add(v string) {
	s.mu.Lock()
	s.m[v] = struct{}{}
	s.mu.Unlock()
}

// Has reports whether a value is present.
func (s *StringSet) Has(v string) bool {
	s.mu.RLock()
	_, ok := s.m[v]
	s.mu.RUnlock()
	return ok
}

// BlockGuard rejects requests originating from a blocked IP.
func BlockGuard(blocked *StringSet, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if blocked.Has(clientIP(r)) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
