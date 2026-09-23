package main

import "testing"

func TestMemoryPresenceKeepsUserOnlineUntilLastConnectionCloses(t *testing.T) {
	presence := newMemoryPresence()

	presence.MarkOnline(42)
	presence.MarkOnline(42)
	if !presence.IsOnline(42) {
		t.Fatal("user should be online after registering connections")
	}

	presence.MarkOffline(42)
	if !presence.IsOnline(42) {
		t.Fatal("user should remain online while one connection is active")
	}

	presence.MarkOffline(42)
	if presence.IsOnline(42) {
		t.Fatal("user should be offline after the last connection closes")
	}
}

func TestMemoryPresenceIgnoresInvalidUserIDs(t *testing.T) {
	presence := newMemoryPresence()
	presence.MarkOnline(0)
	presence.MarkOffline(0)

	if presence.IsOnline(0) {
		t.Fatal("invalid user ID should never be online")
	}
}
