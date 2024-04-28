package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Register a handler for the root path "/"
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check if the request method is GET
		if r.Method == http.MethodGet {
			// Write "ok" response back to the client
			fmt.Fprintf(w, "ok")
		} else {
			// If method is not GET, return method not allowed
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start the HTTP server on port 8080
	log.Println("Server listening on :8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
