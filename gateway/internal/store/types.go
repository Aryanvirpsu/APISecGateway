package store

import "github.com/google/uuid"

// LogInput is a single request_logs row.
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

// AlertInput is a single alerts row. RequestID, AuthSub and Metadata are
// optional and stored as NULL when nil.
type AlertInput struct {
	RequestID *uuid.UUID
	SourceIP  string
	AuthSub   *string
	AlertType string
	Severity  int
	Reason    string
	Metadata  []byte // JSON document, may be nil
}
