package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestReadJSONEnforcesBodyLimit(t *testing.T) {
	payload := `{"content":"` + strings.Repeat("a", maxJSONBodyBytes) + `"}`
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
	recorder := httptest.NewRecorder()

	var body struct {
		Content string `json:"content"`
	}
	err := readJSON(recorder, req, &body)
	if err == nil {
		t.Fatal("expected oversized JSON body to be rejected")
	}

	var maxBytesError *http.MaxBytesError
	if !errors.As(err, &maxBytesError) {
		t.Fatalf("expected http.MaxBytesError, got %T: %v", err, err)
	}
}
