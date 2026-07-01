package store

import (
	"context"

	"github.com/google/uuid"
)

// InsertAlert records a security event raised by the gateway.
func (db *DB) InsertAlert(ctx context.Context, in AlertInput) error {
	var meta interface{}
	if in.Metadata != nil {
		meta = string(in.Metadata)
	}

	_, err := db.Conn.ExecContext(ctx, `
		INSERT INTO alerts
			(id, request_id, source_ip, auth_subject, alert_type, severity, reason, metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8::jsonb)`,
		uuid.New(),
		in.RequestID,
		in.SourceIP,
		in.AuthSub,
		in.AlertType,
		in.Severity,
		in.Reason,
		meta,
	)
	return err
}
