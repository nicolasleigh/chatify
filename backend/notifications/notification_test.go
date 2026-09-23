package notifications

import (
	"errors"
	"testing"
)

func TestDeliveryErrorPreservesClassification(t *testing.T) {
	rootErr := errors.New("endpoint expired")
	err := NewPermanentError(rootErr)

	if !IsPermanentError(err) {
		t.Fatal("IsPermanentError() = false, want true")
	}
	if !errors.Is(err, rootErr) {
		t.Fatal("DeliveryError should unwrap the provider error")
	}
}

func TestTransientErrorIsNotPermanent(t *testing.T) {
	if IsPermanentError(NewTransientError(errors.New("temporary timeout"))) {
		t.Fatal("transient error should not be classified as permanent")
	}
}
