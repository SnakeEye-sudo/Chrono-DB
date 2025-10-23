package main

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// RaftNode represents a Raft consensus node
type RaftNode struct {
	mu          sync.RWMutex
	nodeID      string
	raftPort    int
	peers       map[string]string // nodeID -> address
	state       RaftState
	currentTerm int64
	votedFor    string
	log         []LogEntry
	commitIndex int64
	lastApplied int64
	db          *DBEngine
	crdtStore   *CRDTStore
	dataDir     string
	shutdownCh  chan struct{}
}

// RaftState represents the state of a Raft node
type RaftState int

const (
	Follower RaftState = iota
	Candidate
	Leader
)

// LogEntry represents a log entry in Raft
type LogEntry struct {
	Term    int64       `json:"term"`
	Index   int64       `json:"index"`
	Command interface{} `json:"command"`
}

// NewRaftNode creates a new Raft node
func NewRaftNode(nodeID string, raftPort int, dataDir string, db *DBEngine, crdtStore *CRDTStore) (*RaftNode, error) {
	node := &RaftNode{
		nodeID:      nodeID,
		raftPort:    raftPort,
		peers:       make(map[string]string),
		state:       Follower,
		currentTerm: 0,
		log:         []LogEntry{},
		commitIndex: 0,
		lastApplied: 0,
		db:          db,
		crdtStore:   crdtStore,
		dataDir:     dataDir,
		shutdownCh:  make(chan struct{}),
	}

	// Start background consensus process
	go node.runConsensus()

	log.Printf("Raft node initialized: %s (port: %d)\n", nodeID, raftPort)
	return node, nil
}

// Join adds this node to an existing cluster
func (r *RaftNode) Join(nodeID, addr, leaderAddr string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// In a real implementation, this would send a join request to the leader
	// For now, we'll just add the peer
	r.peers[nodeID] = addr
	log.Printf("Node %s joined cluster. Peers: %v\n", nodeID, r.peers)
	return nil
}

// runConsensus runs the Raft consensus algorithm
func (r *RaftNode) runConsensus() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-r.shutdownCh:
			return
		case <-ticker.C:
			r.mu.RLock()
			state := r.state
			r.mu.RUnlock()

			switch state {
			case Follower:
				// Follower logic - wait for heartbeats
			case Candidate:
				// Start election
				r.startElection()
			case Leader:
				// Send heartbeats
				r.sendHeartbeats()
			}
		}
	}
}

// startElection initiates a new election
func (r *RaftNode) startElection() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.currentTerm++
	r.votedFor = r.nodeID
	log.Printf("Node %s starting election for term %d\n", r.nodeID, r.currentTerm)

	// In a single-node setup, become leader immediately
	if len(r.peers) == 0 {
		r.state = Leader
		log.Printf("Node %s became leader for term %d\n", r.nodeID, r.currentTerm)
	}
}

// sendHeartbeats sends heartbeat messages to all peers
func (r *RaftNode) sendHeartbeats() {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.peers) > 0 {
		log.Printf("Leader %s sending heartbeats to %d peers\n", r.nodeID, len(r.peers))
	}
}

// Apply applies a command to the state machine
func (r *RaftNode) Apply(command interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// In a real implementation, this would replicate the command via Raft
	// For now, apply it directly
	entry := LogEntry{
		Term:    r.currentTerm,
		Index:   int64(len(r.log) + 1),
		Command: command,
	}

	r.log = append(r.log, entry)
	r.commitIndex = entry.Index
	r.lastApplied = entry.Index

	log.Printf("Raft applied command at index %d\n", entry.Index)
	return nil
}

// GetState returns the current state of the Raft node
func (r *RaftNode) GetState() (string, int64) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stateStr := "follower"
	switch r.state {
	case Leader:
		stateStr = "leader"
	case Candidate:
		stateStr = "candidate"
	}

	return stateStr, r.currentTerm
}

// Shutdown stops the Raft node
func (r *RaftNode) Shutdown() error {
	log.Printf("Shutting down Raft node %s\n", r.nodeID)
	close(r.shutdownCh)
	return nil
}
