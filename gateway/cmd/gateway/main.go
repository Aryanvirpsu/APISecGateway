package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"gateway/internal/middleware"
	"gateway/internal/store"
)

func main() {
	// Connect Postgres (docker compose ports)
	db, err := store.NewDB("localhost", 5432, "secuser", "secpass", "apisec")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected to postgres")

	target, err := url.Parse("http://localhost:8081")
	if err != nil {
		log.Fatal(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(target)

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("ok"))
	})

	// Proxy handler
	proxyHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = strings.TrimPrefix(r.URL.Path, "/api")
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}
		proxy.ServeHTTP(w, r)
	})

	// Wrap proxy with request logger middleware
	mux.Handle("/api/", middleware.RequestLogger(db, proxyHandler))

	log.Println("gateway listening on :8080 (logging -> postgres)")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
