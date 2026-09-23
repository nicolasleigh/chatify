package main

import "sync"

// PresenceStore abstracts whether a user currently has at least one active
// WebSocket connection. The first notification release uses process memory;
// a future multi-instance deployment can replace it with Redis without
// changing notification decision logic.
type PresenceStore interface {
	MarkOnline(userID int64)
	MarkOffline(userID int64)
	IsOnline(userID int64) bool
}

// memoryPresence counts connections instead of storing a boolean. A user may
// have multiple browser tabs or devices, so closing one connection must not
// make the user look offline while another connection is still active.
type memoryPresence struct {
	mu          sync.RWMutex
	connections map[int64]int
}

func newMemoryPresence() *memoryPresence {
	return &memoryPresence{connections: make(map[int64]int)}
}

func (p *memoryPresence) MarkOnline(userID int64) {
	if userID <= 0 {
		return
	}
	p.mu.Lock()
	p.connections[userID]++
	p.mu.Unlock()
}

func (p *memoryPresence) MarkOffline(userID int64) {
	if userID <= 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	count := p.connections[userID]
	if count <= 1 {
		delete(p.connections, userID)
		return
	}
	p.connections[userID] = count - 1
}

func (p *memoryPresence) IsOnline(userID int64) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connections[userID] > 0
}
