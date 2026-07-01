package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type user struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type order struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Item   string `json:"item"`
	Cents  int    `json:"amount_cents"`
}

// In-memory fixtures. The gateway is responsible for ensuring callers only ever
// reach their own records; this service just serves whatever id it is asked for.
var users = map[string]user{
	"1": {ID: "1", Name: "Alice", Email: "alice@example.com"},
	"2": {ID: "2", Name: "Bob", Email: "bob@example.com"},
}

var orders = map[string][]order{
	"1": {{ID: "1001", UserID: "1", Item: "Keyboard", Cents: 7999}},
	"2": {{ID: "2001", UserID: "2", Item: "Monitor", Cents: 24999}},
}

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("api ok"))
	})

	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {
		u, ok := users[r.PathValue("id")]
		if !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, u)
	})

	mux.HandleFunc("GET /users/{id}/orders", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if _, ok := users[id]; !ok {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		writeJSON(w, orders[id])
	})

	log.Println("demo api listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "encode error", http.StatusInternalServerError)
	}
}
