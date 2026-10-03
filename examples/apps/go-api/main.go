package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello from go-api")
	})

	addr := ":" + port
	log.Printf("go-api listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
