package notifications

import (
	"strings"
	"testing"

	"github.com/nicolasleigh/chat-app/store"
)

func TestDeliveryRetryDelayUsesCappedExponentialBackoff(t *testing.T) {
	tests := []struct {
		attempts int32
		want     int32
	}{
		{attempts: 1, want: 1},
		{attempts: 2, want: 2},
		{attempts: 3, want: 4},
		{attempts: 10, want: 300},
	}

	for _, test := range tests {
		if got := deliveryRetryDelaySeconds(test.attempts); got != test.want {
			t.Errorf("deliveryRetryDelaySeconds(%d) = %d, want %d", test.attempts, got, test.want)
		}
	}
}

func TestBuildPayloadUsesConversationAndTruncatesBody(t *testing.T) {
	conversationName := "Project Chat"
	content := strings.Repeat("message ", 40)
	payload := buildPayload(42, store.GetNotificationMessageRow{
		ConversationID:   7,
		Content:          &content,
		SenderUsername:   "Alice",
		ConversationName: &conversationName,
	})

	if payload.Title != conversationName {
		t.Fatalf("title = %q, want %q", payload.Title, conversationName)
	}
	if payload.MessageID != 42 {
		t.Fatalf("message ID = %d, want 42", payload.MessageID)
	}
	if payload.ConversationID != 7 {
		t.Fatalf("conversation ID = %d, want 7", payload.ConversationID)
	}
	if len([]rune(payload.Body)) > maxNotificationBodyRunes {
		t.Fatalf("body has %d runes, want at most %d", len([]rune(payload.Body)), maxNotificationBodyRunes)
	}
	if !containsRune(payload.Body, '…') {
		t.Fatalf("body = %q, want truncation marker", payload.Body)
	}
}

func containsRune(value string, target rune) bool {
	for _, current := range value {
		if current == target {
			return true
		}
	}
	return false
}
