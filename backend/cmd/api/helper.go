package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"
)

func writeJSON(w http.ResponseWriter, status int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	return json.NewEncoder(w).Encode(data)
}

func readJSON(w http.ResponseWriter, r *http.Request, ptr any) error {
	maxBytes := 1_048_576 // 1024*1024 -> 1MB
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(ptr)
}

func getClerkUser(ctx context.Context) (*clerk.User, error) {
	claims, ok := clerk.SessionClaimsFromContext(ctx)
	if !ok {
		return nil, errors.New("unauthorized")
	}
	usr, err := user.Get(ctx, claims.Subject)
	if err != nil {
		return nil, err
	}
	if usr == nil {
		return nil, errors.New("User does not exist")
	}

	return usr, nil
}
