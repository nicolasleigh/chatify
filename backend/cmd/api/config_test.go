package main

import (
	"reflect"
	"testing"
)

func TestParseTrustedOrigins(t *testing.T) {
	want := []string{"http://localhost:3000", "https://chat.linze.pro"}
	got := parseTrustedOrigins(" http://localhost:3000, ,https://chat.linze.pro ")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("origins = %#v, want %#v", got, want)
	}
}
