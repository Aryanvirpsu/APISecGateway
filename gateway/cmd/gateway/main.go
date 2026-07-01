package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"gateway/config"
	"gateway/internal/middleware"
	"gateway/internal/store"
)

func main() {
	cfg := config.Load()

	db, err := store.NewDB(cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer db.Close()
	log.Println("connected to postgres")

	// Seed the in-memory caches so we don't hit the database on every request.
	ctx := context.Background()
	blockedSeed, err := db.LoadBlockedIPs(ctx)
	if err != nil {
		log.Fatalf("load blocked ips: %v", err)
	}
	revokedSeed, err := db.LoadRevokedTokens(ctx)
	if err != nil {
		log.Fatalf("load revoked tokens: %v", err)
	}
	blocked := middleware.NewStringSet(blockedSeed)
	revoked := middleware.NewStringSet(revokedSeed)

	limiter := middleware.NewRateLimiter(cfg.RatePerSecond, cfg.RateBurst, cfg.BlockAfter, db, db, blocked)
	limiter.StartCleanup(cfg.CleanupEvery)
	auth := middleware.NewAuthenticator(cfg.Tokens, revoked)

	target, err := url.Parse(cfg.Upstream)
	if err != nil {
		log.Fatalf("parse upstream url: %v", err)
	}
	proxy := newProxy(target)

	// Inner handler: strip the /api prefix before forwarding upstream.
	proxied := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		proxy.ServeHTTP(w, r)
	})

	// Security chain, built inside-out. Execution order is:
	//   recover -> log -> block guard -> rate limit -> auth -> IDOR -> proxy
	secured := middleware.IDORGuard(db, proxied)
	secured = auth.Handler(secured)
	secured = limiter.Handler(secured)
	secured = middleware.BlockGuard(blocked, secured)
	secured = middleware.RequestLogger(db, secured)
	secured = middleware.Recoverer(secured)

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.Handle("/api/", secured)

	srv := &http.Server{
		Addr:              cfg.ListenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	log.Printf("gateway listening on %s -> %s", cfg.ListenAddr, cfg.Upstream)
	log.Fatal(srv.ListenAndServe())
}

// newProxy returns a reverse proxy that fails cleanly when the upstream is down.
func newProxy(target *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(target)
	proxy.ErrorHandler = func(w http.ResponseWriter, _ *http.Request, err error) {
		log.Printf("proxy error: %v", err)
		http.Error(w, "upstream unavailable", http.StatusBadGateway)
	}
	return proxy
}
