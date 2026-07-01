package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ctxKey int

const stateKey ctxKey = iota

// State carries per-request information that is filled in as the request moves
// through the middleware chain and read back when the request is logged. It is
// stored in the request context by pointer so later middleware can enrich it.
type State struct {
	RequestID   uuid.UUID
	AuthSubject *string
	TokenID     *string
	UserID      string // user id the authenticated token is scoped to
}

func withState(r *http.Request, st *State) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), stateKey, st))
}

func stateFrom(r *http.Request) *State {
	st, _ := r.Context().Value(stateKey).(*State)
	return st
}
