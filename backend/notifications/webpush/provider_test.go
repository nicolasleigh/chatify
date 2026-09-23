package webpush

import (
	"net/http"
	"testing"
)

func validConfig() Config {
	return Config{
		VAPIDPublicKey:  "public-key",
		VAPIDPrivateKey: "private-key",
		Subject:         "mailto:notifications@example.com",
	}
}

func TestNewRequiresVAPIDSettings(t *testing.T) {
	if _, err := New(Config{}); err == nil {
		t.Fatal("New() error = nil, want VAPID validation error")
	}

	client, err := New(validConfig())
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if client.ttl != defaultTTL {
		t.Fatalf("ttl = %d, want %d", client.ttl, defaultTTL)
	}
}

func TestPermanentStatusClassification(t *testing.T) {
	for _, statusCode := range []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusNotFound, http.StatusGone} {
		if !isPermanentStatus(statusCode) {
			t.Errorf("isPermanentStatus(%d) = false, want true", statusCode)
		}
	}
	for _, statusCode := range []int{http.StatusRequestTimeout, http.StatusTooManyRequests, http.StatusBadGateway, http.StatusServiceUnavailable} {
		if isPermanentStatus(statusCode) {
			t.Errorf("isPermanentStatus(%d) = true, want false", statusCode)
		}
	}
}
