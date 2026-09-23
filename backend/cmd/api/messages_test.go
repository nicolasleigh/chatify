package main

import "testing"

func TestValidateCreateMessageInput(t *testing.T) {
	type testCase struct {
		name         string
		conversation int64
		messageType  *string
		content      *string
		wantError    bool
	}

	validType := "text"
	validContent := "hello"
	empty := ""
	longContent := make([]byte, maxMessageContentBytes+1)

	tests := []testCase{
		{name: "valid", conversation: 1, messageType: &validType, content: &validContent},
		{name: "invalid conversation", conversation: 0, messageType: &validType, content: &validContent, wantError: true},
		{name: "missing type", conversation: 1, content: &validContent, wantError: true},
		{name: "empty content", conversation: 1, messageType: &validType, content: &empty, wantError: true},
		{name: "oversized content", conversation: 1, messageType: &validType, content: stringPointer(string(longContent)), wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCreateMessageInput(tt.conversation, tt.messageType, tt.content)
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, want error: %t", err, tt.wantError)
			}
		})
	}
}

func stringPointer(value string) *string { return &value }
