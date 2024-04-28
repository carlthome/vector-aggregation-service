package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"

	"github.com/go-redis/redis/v8"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

type Vectors struct {
	Vectors [][]float64 `json:"vectors"`
}

func connectToRedis(host string, port string) *redis.Client {
	// Create a new Redis client.
	client := redis.NewClient(&redis.Options{
		Addr: host + ":" + port,
		DB:   0,
	})

	// Ping the Redis server to check the connection.
	ctx := client.Context()
	val, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis:", val)

	return client
}

func getState(redisClient *redis.Client) ([]float64, []float64) {

	// Retrieve state from Redis.
	var val string
	ctx := redisClient.Context()
	val, err := redisClient.Get(ctx, "state").Result()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Retrieved value from Redis:", val)

	// Parse JSON to floats.
	var floats []float64
	err = json.Unmarshal([]byte(val), &floats)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Parsed JSON value:", floats)

	// Split the floats into total and count.
	total := floats[:len(floats)/2]
	count := floats[len(floats)/2:]

	return total, count
}

func setState(total []float64, count []float64, redisClient *redis.Client) {

	// Pack total and count into a single slice.
	floats := append(total, count...)

	// Convert floats to JSON.
	bytes, err := json.Marshal(floats)
	if err != nil {
		log.Fatal(err)
	}

	// Store the value in Redis.
	ctx := redisClient.Context()
	redisClient.Set(ctx, "state", string(bytes), 0)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("Status check by %v\n", r.RemoteAddr)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func centroidHandler(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	log.Printf("Request from %v\n", r.RemoteAddr)

	// Decode JSON request.
	var vectors Vectors
	err := json.NewDecoder(r.Body).Decode(&vectors)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Fatalf("Error decoding JSON: %v\n", err)
	}
	log.Printf("Decoded JSON: %v\n", vectors)

	centroid := aggregate(vectors, redisClient)

	// Encode and write JSON response.
	result, _ := json.Marshal(centroid)
	w.Write(result)
}

func aggregate(vectors Vectors, redisClient *redis.Client) []float64 {
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

	// Accumulate new values in Redis.
	total, count := getState(redisClient)
	for i := 0; i < numRows; i++ {
		row := matrix.RowView(i)
		for j := 0; j < numColumns; j++ {
			total[j] += row.AtVec(j)
			count[j]++
		}
	}
	setState(total, count, redisClient)

	// Calculate mean/stdev per column.
	col := make([]float64, numRows)
	mean := make([]float64, numColumns)
	stdev := make([]float64, numColumns)
	for j := 0; j < numColumns; j++ {
		log.Println("Calculating mean and stdev for column", j)
		mat.Col(col, j, matrix)
		mean[j], stdev[j] = stat.MeanStdDev(col, nil)
	}
	log.Printf("Mean per column: %v\n", mean)
	log.Printf("Stdev per column: %v\n", stdev)

	// Compute centroid.
	centroid := mat.NewVecDense(numColumns, nil)
	for i := 0; i < numRows; i++ {
		row := matrix.RowView(i)
		centroid.AddVec(centroid, row)
	}
	centroid.ScaleVec(1/float64(numRows), centroid)
	centroidData := centroid.RawVector().Data
	return centroidData
}

func main() {
	// Parse command line arguments.
	dim := flag.Int("dim", 3, "Dimensionality of input vectors")
	host := flag.String("host", "0.0.0.0", "HTTP server host")
	port := flag.String("port", "8080", "HTTP server port")
	redisHost := flag.String("redis-host", "redis", "Redis server host")
	redisPort := flag.String("redis-port", "6379", "Redis server port")
	flag.Parse()

	// Connect to Redis.
	redisClient := connectToRedis(*redisHost, *redisPort)
	defer redisClient.Close()

	// Initialize state in Redis if it doesn't exist yet.
	_, err := redisClient.Get(redisClient.Context(), "state").Result()
	if err != nil {
		log.Println("Initializing state in Redis")
		total := make([]float64, *dim)
		count := make([]float64, *dim)
		setState(total, count, redisClient)
	}

	// Register routes.
	http.HandleFunc("/status", statusHandler)
	http.HandleFunc("/centroid", func(w http.ResponseWriter, r *http.Request) { centroidHandler(w, r, redisClient) })

	// Start HTTP server on configured port.
	addr := *host + ":" + *port
	log.Printf("Server listening on %v\n", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
