package store

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// DB wraps the database/sql connection pool used by the gateway.
type DB struct {
	Conn *sql.DB
}

// NewDB opens a Postgres connection pool and verifies connectivity.
func NewDB(host string, port int, user, pass, name string) (*DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, pass, name,
	)

	conn, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	conn.SetMaxOpenConns(10)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(30 * time.Minute)

	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &DB{Conn: conn}, nil
}

// Close releases the underlying connection pool.
func (db *DB) Close() error {
	return db.Conn.Close()
}
