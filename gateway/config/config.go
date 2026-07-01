package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Token represents an API credential recognised by the gateway. Each token is
// tied to a subject (who the caller is) and the user id whose resources the
// caller is allowed to touch. In a real deployment these would come from an
// identity provider; for the demo we load a small static set.
type Token struct {
	Subject string
	UserID  string
}

// Config holds all runtime settings, populated from the environment with sane
// defaults so the gateway runs out of the box against the bundled compose file.
type Config struct {
	ListenAddr string
	Upstream   string

	DBHost string
	DBPort int
	DBUser string
	DBPass string
	DBName string

	// Rate limiting, applied per client IP.
	RatePerSecond float64
	RateBurst     int
	BlockAfter    int           // rejections before an IP is hard-blocked
	CleanupEvery  time.Duration // how often idle limiter entries are reaped

	// Tokens maps a token value to its metadata.
	Tokens map[string]Token
}

// Load builds a Config from environment variables, falling back to defaults
// that match docker-compose.yml.
func Load() Config {
	return Config{
		ListenAddr:    getenv("GATEWAY_ADDR", ":8080"),
		Upstream:      getenv("UPSTREAM_URL", "http://localhost:8081"),
		DBHost:        getenv("DB_HOST", "localhost"),
		DBPort:        getenvInt("DB_PORT", 5432),
		DBUser:        getenv("DB_USER", "secuser"),
		DBPass:        getenv("DB_PASSWORD", "secpass"),
		DBName:        getenv("DB_NAME", "apisec"),
		RatePerSecond: getenvFloat("RATE_PER_SECOND", 5),
		RateBurst:     getenvInt("RATE_BURST", 10),
		BlockAfter:    getenvInt("BLOCK_AFTER", 20),
		CleanupEvery:  5 * time.Minute,
		Tokens:        loadTokens(),
	}
}

// loadTokens reads API_TOKENS in the form
//
//	token:subject:userid,token2:subject2:userid2
//
// and falls back to a pair of demo credentials when unset.
func loadTokens() map[string]Token {
	raw := os.Getenv("API_TOKENS")
	if raw == "" {
		return map[string]Token{
			"alice-token": {Subject: "alice", UserID: "1"},
			"bob-token":   {Subject: "bob", UserID: "2"},
		}
	}

	tokens := make(map[string]Token)
	for _, entry := range strings.Split(raw, ",") {
		parts := strings.Split(strings.TrimSpace(entry), ":")
		if len(parts) != 3 {
			continue
		}
		tokens[parts[0]] = Token{Subject: parts[1], UserID: parts[2]}
	}
	return tokens
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func getenvFloat(key string, def float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return def
}
