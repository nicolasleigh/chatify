package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithRequestIDPreservesValidHeader(t *testing.T) {
	handler := withRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := requestIDFromContext(r.Context()); got != "request-123" {
			t.Fatalf("request ID in context = %q, want %q", got, "request-123")
		}
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(requestIDHeader, "request-123")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if got := recorder.Header().Get(requestIDHeader); got != "request-123" {
		t.Fatalf("response request ID = %q, want %q", got, "request-123")
	}
}

func TestWithRequestIDGeneratesMissingHeader(t *testing.T) {
	handler := withRequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !validRequestID(requestIDFromContext(r.Context())) {
			t.Fatal("expected a valid request ID in context")
		}
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if !validRequestID(recorder.Header().Get(requestIDHeader)) {
		t.Fatal("expected a valid request ID in response")
	}
}
