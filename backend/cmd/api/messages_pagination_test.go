package main

import "testing"

func TestParseMessageHistoryLimit(t *testing.T) {
	tests := []struct {
		name      string
		rawLimit  string
		want      int32
		wantError bool
	}{
		{name: "default", want: defaultMessageHistoryLimit},
		{name: "custom", rawLimit: "100", want: 100},
		{name: "zero", rawLimit: "0", wantError: true},
		{name: "too large", rawLimit: "101", wantError: true},
		{name: "invalid", rawLimit: "abc", wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseMessageHistoryLimit(tt.rawLimit)
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, want error: %t", err, tt.wantError)
			}
			if err == nil && got != tt.want {
				t.Fatalf("limit = %d, want %d", got, tt.want)
			}
		})
	}
}
