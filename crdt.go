package main

import (
	"sync"
	"time"
)

// CRDTStore implements Conflict-free Replicated Data Type for multi-master replication
type CRDTStore struct {
	mu      sync.RWMutex
	gcounter map[string]GCounter
	lww     map[string]LWWRegister
}

// GCounter implements a grow-only counter CRDT
type GCounter struct {
	NodeCounts map[string]int64 `json:"node_counts"`
}

// LWWRegister implements Last-Writer-Wins register
type LWWRegister struct {
	Value     interface{} `json:"value"`
	Timestamp time.Time   `json:"timestamp"`
	NodeID    string      `json:"node_id"`
}

// NewCRDTStore creates a new CRDT store
func NewCRDTStore() *CRDTStore {
	return &CRDTStore{
		gcounter: make(map[string]GCounter),
		lww:      make(map[string]LWWRegister),
	}
}

// IncrementCounter increments a grow-only counter
func (c *CRDTStore) IncrementCounter(key, nodeID string, delta int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	gc, exists := c.gcounter[key]
	if !exists {
		gc = GCounter{NodeCounts: make(map[string]int64)}
	}

	gc.NodeCounts[nodeID] += delta
	c.gcounter[key] = gc
}

// GetCounter returns the total value of a counter
func (c *CRDTStore) GetCounter(key string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	gc, exists := c.gcounter[key]
	if !exists {
		return 0
	}

	var total int64
	for _, count := range gc.NodeCounts {
		total += count
	}
	return total
}

// SetLWW sets a value using Last-Writer-Wins semantics
func (c *CRDTStore) SetLWW(key string, value interface{}, timestamp time.Time, nodeID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	existing, exists := c.lww[key]
	if !exists || timestamp.After(existing.Timestamp) ||
		(timestamp.Equal(existing.Timestamp) && nodeID > existing.NodeID) {
		c.lww[key] = LWWRegister{
			Value:     value,
			Timestamp: timestamp,
			NodeID:    nodeID,
		}
	}
}

// GetLWW gets the current value from LWW register
func (c *CRDTStore) GetLWW(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	reg, exists := c.lww[key]
	if !exists {
		return nil, false
	}
	return reg.Value, true
}

// MergeCounter merges counter state from another node
func (c *CRDTStore) MergeCounter(key string, other GCounter) {
	c.mu.Lock()
	defer c.mu.Unlock()

	local, exists := c.gcounter[key]
	if !exists {
		local = GCounter{NodeCounts: make(map[string]int64)}
	}

	for nodeID, count := range other.NodeCounts {
		if count > local.NodeCounts[nodeID] {
			local.NodeCounts[nodeID] = count
		}
	}

	c.gcounter[key] = local
}

// MergeLWW merges LWW register from another node
func (c *CRDTStore) MergeLWW(key string, other LWWRegister) {
	c.mu.Lock()
	defer c.mu.Unlock()

	local, exists := c.lww[key]
	if !exists || other.Timestamp.After(local.Timestamp) ||
		(other.Timestamp.Equal(local.Timestamp) && other.NodeID > local.NodeID) {
		c.lww[key] = other
	}
}
