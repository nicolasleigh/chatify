package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBadRequestResponseDoesNotExposeInternalError(t *testing.T) {
	recorder := httptest.NewRecorder()

	badRequestResponse(recorder, assertError("database constraint details"))

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	var message string
	if err := json.Unmarshal(recorder.Body.Bytes(), &message); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if message != "invalid request" {
		t.Fatalf("message = %q, want %q", message, "invalid request")
	}
}

type assertError string

func (e assertError) Error() string { return string(e) }
