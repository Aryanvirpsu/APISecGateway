package middleware

import (
	"net/http"
	"strings"

	"gateway/config"
)

// Authenticator validates the bearer token, rejects revoked tokens and records
// the caller identity on the request State. Requests without a valid token are
// answered with 401.
type Authenticator struct {
	tokens  map[string]config.Token
	revoked *StringSet
}

// NewAuthenticator builds an Authenticator from the configured token set and a
// cache of revoked token ids.
func NewAuthenticator(tokens map[string]config.Token, revoked *StringSet) *Authenticator {
	return &Authenticator{tokens: tokens, revoked: revoked}
}

// Handler wraps next with bearer-token authentication.
func (a *Authenticator) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw := bearerToken(r)
		if raw == "" {
			http.Error(w, "missing bearer token", http.StatusUnauthorized)
			return
		}

		tok, ok := a.tokens[raw]
		if !ok || a.revoked.Has(raw) {
			http.Error(w, "invalid or revoked token", http.StatusUnauthorized)
			return
		}

		if st := stateFrom(r); st != nil {
			subject, tokenID := tok.Subject, raw
			st.AuthSubject = &subject
			st.TokenID = &tokenID
			st.UserID = tok.UserID
		}
		next.ServeHTTP(w, r)
	})
}

func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if len(h) > len(prefix) && strings.EqualFold(h[:len(prefix)], prefix) {
		return strings.TrimSpace(h[len(prefix):])
	}
	return ""
}
