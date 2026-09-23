package main

import "testing"

func TestIsTrustedOrigin(t *testing.T) {
	trustedOrigins := []string{"http://localhost:3000", "https://chat.example.com"}

	tests := []struct {
		name   string
		origin string
		want   bool
	}{
		{name: "configured origin", origin: "https://chat.example.com", want: true},
		{name: "unknown origin", origin: "https://evil.example.com", want: false},
		{name: "missing origin", origin: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTrustedOrigin(tt.origin, trustedOrigins); got != tt.want {
				t.Fatalf("isTrustedOrigin(%q) = %v, want %v", tt.origin, got, tt.want)
			}
		})
	}
}
