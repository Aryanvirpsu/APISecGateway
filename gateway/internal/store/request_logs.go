package store

import (
	"context"

	"github.com/google/uuid"
)

// InsertRequestLog persists a single proxied request.
func (db *DB) InsertRequestLog(ctx context.Context, in LogInput) error {
	_, err := db.Conn.ExecContext(ctx, `
		INSERT INTO request_logs
			(id, request_id, source_ip, method, path, status, latency_ms, user_agent, auth_subject, token_id)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		uuid.New(),
		in.RequestID,
		in.SourceIP,
		in.Method,
		in.Path,
		in.Status,
		in.LatencyMS,
		in.UserAgent,
		in.AuthSub,
		in.TokenID,
	)
	return err
}
