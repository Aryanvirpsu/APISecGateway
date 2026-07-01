package store

import "context"

// LoadRevokedTokens returns every revoked token id, used to seed the in-memory
// cache at start-up.
func (db *DB) LoadRevokedTokens(ctx context.Context) ([]string, error) {
	rows, err := db.Conn.QueryContext(ctx, `SELECT token_id FROM revoked_tokens`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
