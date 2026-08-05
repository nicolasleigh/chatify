package main

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func readJSON(w http.ResponseWriter, r *http.Request, ptr any) error {
	maxBytes := 1_048_576 // 1024*1024 -> 1MB
	body := r.Body
	http.MaxBytesReader(w, body, int64(maxBytes))

	dec := json.NewDecoder(body)
	dec.DisallowUnknownFields()
	return dec.Decode(ptr)
}
