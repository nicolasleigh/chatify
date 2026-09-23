package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEnableCORSAllowsConfiguredPreflightMethods(t *testing.T) {
	app := &application{config: config{cors: cors{trustedOrigins: []string{"https://chat.example.com"}}}}
	handler := app.enableCORS(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/messages/1", nil)
	req.Header.Set("Origin", "https://chat.example.com")
	req.Header.Set("Access-Control-Request-Method", http.MethodGet)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, req)

	if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != "https://chat.example.com" {
		t.Fatalf("allow origin = %q, want configured origin", got)
	}
	methods := recorder.Header().Get("Access-Control-Allow-Methods")
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodOptions} {
		if !strings.Contains(methods, method) {
			t.Fatalf("allow methods %q does not contain %s", methods, method)
		}
	}
}
