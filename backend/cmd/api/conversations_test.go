package main

import "testing"

func TestValidateGroupInput(t *testing.T) {
	tests := []struct {
		name      string
		groupName string
		memberIDs []int64
		wantError bool
	}{
		{name: "valid", groupName: "Friends", memberIDs: []int64{1, 2}},
		{name: "missing name", memberIDs: []int64{1}, wantError: true},
		{name: "missing members", groupName: "Friends", wantError: true},
		{name: "zero member", groupName: "Friends", memberIDs: []int64{0}, wantError: true},
		{name: "duplicate member", groupName: "Friends", memberIDs: []int64{1, 1}, wantError: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGroupInput(tt.groupName, tt.memberIDs)
			if (err != nil) != tt.wantError {
				t.Fatalf("error = %v, want error: %t", err, tt.wantError)
			}
		})
	}
}
