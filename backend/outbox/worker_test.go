package outbox

import "testing"

func TestRetryDelaySecondsUsesCappedExponentialBackoff(t *testing.T) {
	tests := []struct {
		attempts int32
		want     int32
	}{
		{attempts: 1, want: 1},
		{attempts: 2, want: 2},
		{attempts: 3, want: 4},
		{attempts: 4, want: 8},
		{attempts: 10, want: 300},
	}

	for _, test := range tests {
		if got := retryDelaySeconds(test.attempts); got != test.want {
			t.Errorf("retryDelaySeconds(%d) = %d, want %d", test.attempts, got, test.want)
		}
	}
}
