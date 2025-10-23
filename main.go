package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	nodeID   = flag.String("node", "node1", "Node ID for this instance")
	httpPort = flag.Int("http", 8080, "HTTP API port")
	raftPort = flag.Int("raft", 9000, "Raft consensus port")
	join     = flag.String("join", "", "Address of existing node to join")
	dataDir  = flag.String("data", "./data", "Data directory")
)

func main() {
	flag.Parse()

	log.Printf("Starting Chrono-DB node: %s\n", *nodeID)
	log.Printf("HTTP API: http://localhost:%d\n", *httpPort)
	log.Printf("Raft Port: %d\n", *raftPort)
	log.Printf("Data Directory: %s\n", *dataDir)

	// Initialize database engine
	db, err := NewDBEngine(*dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Initialize CRDT store
	crdtStore := NewCRDTStore()
	log.Println("CRDT store initialized for multi-master replication")

	// Initialize Raft consensus
	raftNode, err := NewRaftNode(*nodeID, *raftPort, *dataDir, db, crdtStore)
	if err != nil {
		log.Fatalf("Failed to initialize Raft: %v", err)
	}
	defer raftNode.Shutdown()

	// Join existing cluster if specified
	if *join != "" {
		log.Printf("Joining cluster at: %s\n", *join)
		if err := raftNode.Join(*nodeID, fmt.Sprintf("localhost:%d", *raftPort), *join); err != nil {
			log.Printf("Warning: Failed to join cluster: %v", err)
		}
	}

	// Start HTTP API server
	apiServer := NewAPIServer(*httpPort, db, raftNode, crdtStore)
	go func() {
		if err := apiServer.Start(); err != nil {
			log.Fatalf("API server failed: %v", err)
		}
	}()

	log.Printf("Chrono-DB is running. API available at http://localhost:%d\n", *httpPort)
	log.Println("Press Ctrl+C to shutdown...")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	log.Println("\nShutting down Chrono-DB...")
}
