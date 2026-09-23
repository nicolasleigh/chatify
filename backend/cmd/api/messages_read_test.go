package main

import "testing"

func TestValidateMarkReadInput(t *testing.T) {
	messageID := int64(2)
	if err := validateMarkReadInput(1, &messageID); err != nil {
		t.Fatalf("valid input error = %v", err)
	}
	if err := validateMarkReadInput(0, &messageID); err == nil {
		t.Fatal("expected invalid conversation ID error")
	}
	if err := validateMarkReadInput(1, nil); err == nil {
		t.Fatal("expected missing message ID error")
	}
}
