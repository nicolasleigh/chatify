package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const maxJSONBodyBytes = 1_048_576

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func readJSON(w http.ResponseWriter, r *http.Request, ptr any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(ptr); err != nil {
		return err
	}

	var trailing any
	if err := dec.Decode(&trailing); err != io.EOF {
		if err == nil {
			return errors.New("request body must contain a single JSON value")
		}
		return err
	}

	return nil
}
