package main

import (
	"fmt"
	"log"
	"net/http"

	"gonum.org/v1/gonum/mat"
)

func main() {
	// Initialize with the length of the vector, followed by a slice of floats containing the data.
	u := mat.NewVecDense(3, []float64{1, 2, 3})
	v := mat.NewVecDense(3, []float64{4, 5, 6})

	// Add the vectors u and v and save results in u.
	u.AddVec(u, v)

	// Display the sum of the vectors.
	fmt.Printf("The sum of the vectors is: %v\n", u)

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
