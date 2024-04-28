package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"
)

func TestStatusEndpoint(t *testing.T) {
	req, err := http.NewRequest("GET", "/status", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(statusHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := `{"status":"ok"}` + "\n"
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestCentroidEndpoint(t *testing.T) {
	// Load example data from JSON file
	file, err := os.Open("example.json")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	var vectors Vectors
	if err := json.NewDecoder(file).Decode(&vectors); err != nil {
		t.Fatal(err)
	}

	body, err := json.Marshal(vectors)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/centroid", bytes.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(centroidHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var centroid []float64
	if err := json.Unmarshal(rr.Body.Bytes(), &centroid); err != nil {
		t.Fatal(err)
	}

	expectedCentroid := []float64{27.75, 55.5, 83.25}
	if !reflect.DeepEqual(centroid, expectedCentroid) {
		t.Errorf("handler returned unexpected centroid: got %v want %v",
			centroid, expectedCentroid)
	}
}
