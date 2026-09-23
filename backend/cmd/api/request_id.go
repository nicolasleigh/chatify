package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
)

const requestIDHeader = "X-Request-ID"

type requestIDContextKey struct{}

func withRequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := strings.TrimSpace(r.Header.Get(requestIDHeader))
		if !validRequestID(requestID) {
			var err error
			requestID, err = newRequestID()
			if err != nil {
				http.Error(w, "failed to create request id", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set(requestIDHeader, requestID)
		ctx := context.WithValue(r.Context(), requestIDContextKey{}, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func requestIDFromContext(ctx context.Context) string {
	requestID, _ := ctx.Value(requestIDContextKey{}).(string)
	return requestID
}

func validRequestID(requestID string) bool {
	if requestID == "" || len(requestID) > 128 {
		return false
	}

	for _, r := range requestID {
		if r <= ' ' || r == 0x7f {
			return false
		}
	}

	return true
}

func newRequestID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}
