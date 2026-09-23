package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMessageHistoryRequiresAuthentication(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/messages/1", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusForbidden)
	}
}

func TestHealthRemainsPublic(t *testing.T) {
	handler := (&application{}).NewRouter()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}
