package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

type Vectors struct {
	Vectors [][]float64 `json:"vectors"`
}

func main() {

	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		fmt.Printf("Status check by %v\n", r.RemoteAddr)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	http.HandleFunc("/sum", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
		fmt.Printf("Summation request received from %v\n", r.RemoteAddr)

		// Decode request as JSON.
		var vectors Vectors
		err := json.NewDecoder(r.Body).Decode(&vectors)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			fmt.Printf("Error decoding JSON: %v\n", err)
			return
		}
		fmt.Printf("Decoded JSON: %v\n", vectors)

		// Convert data into a gonum matrix.
		numRows := len(vectors.Vectors)
		numColumns := len(vectors.Vectors[0])
		data := make([]float64, numRows*numColumns)
		for i, row := range vectors.Vectors {
			for j, val := range row {
				data[i*numColumns+j] = val
			}
		}
		matrix := mat.NewDense(numRows, numColumns, data)

		// Calculate mean/stdev per column.
		col := make([]float64, numRows)
		mean := make([]float64, numColumns)
		stdev := make([]float64, numColumns)
		for j := 0; j < numColumns; j++ {
			fmt.Println("Calculating mean and stdev for column", j)
			mat.Col(col, j, matrix)
			mean[j], stdev[j] = stat.MeanStdDev(col, nil)
		}
		fmt.Printf("Mean per column: %v\n", mean)
		fmt.Printf("Stdev per column: %v\n", stdev)

		// Compute centroid.
		u := mat.NewVecDense(numColumns, nil)
		for i := 0; i < numRows; i++ {
			row := matrix.RowView(i)
			u.AddVec(u, row)
		}
		u.ScaleVec(1/float64(numRows), u)

		// Return centroid as response.
		result, _ := json.Marshal(u.RawVector().Data)
		w.Write(result)
}

func main() {
	// Default port.
	defaultPort := ":8080"
	port := flag.String("port", defaultPort, "HTTP server port")
	flag.Parse()

	// Register routes.
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/sum", sumHandler)

	// Start HTTP server on configured port.
	log.Printf("Server listening on %v\n", *port)
	log.Fatal(http.ListenAndServe(*port, nil))
}
