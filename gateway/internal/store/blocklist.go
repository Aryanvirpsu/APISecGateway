package store

import "context"

// LoadBlockedIPs returns every IP currently on the block list, used to seed the
// in-memory cache at start-up.
func (db *DB) LoadBlockedIPs(ctx context.Context) ([]string, error) {
	rows, err := db.Conn.QueryContext(ctx, `SELECT ip FROM blocked_ips`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ips []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return nil, err
		}
		ips = append(ips, ip)
	}
	return ips, rows.Err()
}

// BlockIP adds an IP to the block list, ignoring duplicates.
func (db *DB) BlockIP(ctx context.Context, ip, reason string) error {
	_, err := db.Conn.ExecContext(ctx, `
		INSERT INTO blocked_ips (ip, reason)
		VALUES ($1, $2)
		ON CONFLICT (ip) DO NOTHING`,
		ip, reason,
	)
	return err
}
