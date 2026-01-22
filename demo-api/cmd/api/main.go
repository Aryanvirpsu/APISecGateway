package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("api ok"))
	})

	log.Println("demo api listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
