package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// APIServer provides the REST API interface
type APIServer struct {
	port      int
	db        *DBEngine
	raftNode  *RaftNode
	crdtStore *CRDTStore
}

// NewAPIServer creates a new API server
func NewAPIServer(port int, db *DBEngine, raftNode *RaftNode, crdtStore *CRDTStore) *APIServer {
	return &APIServer{
		port:      port,
		db:        db,
		raftNode:  raftNode,
		crdtStore: crdtStore,
	}
}

// Start starts the API server
func (s *APIServer) Start() error {
	http.HandleFunc("/", s.handleRoot)
	http.HandleFunc("/api/v1/insert", s.handleInsert)
	http.HandleFunc("/api/v1/query", s.handleQuery)
	http.HandleFunc("/api/v1/history", s.handleHistory)
	http.HandleFunc("/api/v1/temporal", s.handleTemporal)
	http.HandleFunc("/api/v1/status", s.handleStatus)
	http.HandleFunc("/api/v1/crdt/counter", s.handleCounter)

	addr := fmt.Sprintf(":" + "%d", s.port)
	log.Printf("API server listening on %s\n", addr)
	return http.ListenAndServe(addr, nil)
}

// handleRoot handles root endpoint
func (s *APIServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"service": "Chrono-DB",
		"version": "1.0.0",
		"description": "Distributed temporal database with CRDT and Raft consensus",
	})
}

// handleInsert handles data insertion
func (s *APIServer) handleInsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Key        string      `json:"key"`
		Value      interface{} `json:"value"`
		ValidStart string      `json:"valid_start,omitempty"`
		ValidEnd   string      `json:"valid_end,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse time or use defaults
	validStart := time.Now()
	validEnd := time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC)

	if req.ValidStart != "" {
		if t, err := time.Parse(time.RFC3339, req.ValidStart); err == nil {
			validStart = t
		}
	}
	if req.ValidEnd != "" {
		if t, err := time.Parse(time.RFC3339, req.ValidEnd); err == nil {
			validEnd = t
		}
	}

	if err := s.db.Insert(req.Key, req.Value, validStart, validEnd); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "success",
		"key":    req.Key,
	})
}

// handleQuery handles current value queries
func (s *APIServer) handleQuery(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key parameter required", http.StatusBadRequest)
		return
	}

	value, found := s.db.QueryCurrent(key)
	w.Header().Set("Content-Type", "application/json")

	if !found {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"found": false,
			"key":   key,
		})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"found": true,
		"key":   key,
		"value": value,
	})
}

// handleHistory returns historical records
func (s *APIServer) handleHistory(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key parameter required", http.StatusBadRequest)
		return
	}

	history := s.db.GetHistory(key)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key":     key,
		"history": history,
	})
}

// handleTemporal handles bitemporal queries
func (s *APIServer) handleTemporal(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	asOfStr := r.URL.Query().Get("as_of")
	validStr := r.URL.Query().Get("valid_time")

	if key == "" {
		http.Error(w, "key parameter required", http.StatusBadRequest)
		return
	}

	asOf := time.Now()
	validTime := time.Now()

	if asOfStr != "" {
		if t, err := time.Parse(time.RFC3339, asOfStr); err == nil {
			asOf = t
		}
	}
	if validStr != "" {
		if t, err := time.Parse(time.RFC3339, validStr); err == nil {
			validTime = t
		}
	}

	value, found := s.db.QueryTemporal(key, asOf, validTime)
	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"found":      found,
		"key":        key,
		"value":      value,
		"as_of":      asOf.Format(time.RFC3339),
		"valid_time": validTime.Format(time.RFC3339),
	})
}

// handleStatus returns cluster status
func (s *APIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	state, term := s.raftNode.GetState()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"node_id":     s.raftNode.nodeID,
		"raft_state":  state,
		"raft_term":   term,
		"timestamp":   time.Now().Format(time.RFC3339),
	})
}

// handleCounter handles CRDT counter operations
func (s *APIServer) handleCounter(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key parameter required", http.StatusBadRequest)
		return
	}

	if r.Method == http.MethodPost {
		// Increment counter
		s.crdtStore.IncrementCounter(key, s.raftNode.nodeID, 1)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "incremented",
			"key":    key,
		})
		return
	}

	// GET - return counter value
	count := s.crdtStore.GetCounter(key)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"key":   key,
		"value": count,
	})
}
