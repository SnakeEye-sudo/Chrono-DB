package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// DBEngine implements bitemporal database functionality
type DBEngine struct {
	mu       sync.RWMutex
	data     map[string][]TemporalRecord
	dataDir  string
}

// TemporalRecord represents a bitemporal data record
type TemporalRecord struct {
	Key              string                 `json:"key"`
	Value            interface{}            `json:"value"`
	ValidTimeStart   time.Time              `json:"valid_time_start"`
	ValidTimeEnd     time.Time              `json:"valid_time_end"`
	TransactionTime  time.Time              `json:"transaction_time"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// NewDBEngine creates a new database engine instance
func NewDBEngine(dataDir string) (*DBEngine, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db := &DBEngine{
		data:    make(map[string][]TemporalRecord),
		dataDir: dataDir,
	}

	// Load existing data
	if err := db.loadData(); err != nil {
		return nil, fmt.Errorf("failed to load data: %w", err)
	}

	return db, nil
}

// Insert adds a new temporal record
func (db *DBEngine) Insert(key string, value interface{}, validStart, validEnd time.Time) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	record := TemporalRecord{
		Key:             key,
		Value:           value,
		ValidTimeStart:  validStart,
		ValidTimeEnd:    validEnd,
		TransactionTime: time.Now(),
	}

	db.data[key] = append(db.data[key], record)
	return db.persistData()
}

// QueryTemporal performs bitemporal queries
func (db *DBEngine) QueryTemporal(key string, asOfTime, validTime time.Time) (interface{}, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	records, exists := db.data[key]
	if !exists {
		return nil, false
	}

	// Find the record valid at the specified times
	for i := len(records) - 1; i >= 0; i-- {
		rec := records[i]
		// Check transaction time (as-of time)
		if rec.TransactionTime.After(asOfTime) {
			continue
		}
		// Check valid time
		if validTime.After(rec.ValidTimeStart) && validTime.Before(rec.ValidTimeEnd) {
			return rec.Value, true
		}
	}

	return nil, false
}

// QueryCurrent returns the current value for a key
func (db *DBEngine) QueryCurrent(key string) (interface{}, bool) {
	return db.QueryTemporal(key, time.Now(), time.Now())
}

// GetHistory returns all historical records for a key
func (db *DBEngine) GetHistory(key string) []TemporalRecord {
	db.mu.RLock()
	defer db.mu.RUnlock()

	records, exists := db.data[key]
	if !exists {
		return []TemporalRecord{}
	}

	// Return a copy to prevent external modification
	history := make([]TemporalRecord, len(records))
	copy(history, records)
	return history
}

// persistData saves data to disk
func (db *DBEngine) persistData() error {
	dataFile := filepath.Join(db.dataDir, "chrono_db.json")
	file, err := os.Create(dataFile)
	if err != nil {
		return fmt.Errorf("failed to create data file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(db.data)
}

// loadData loads data from disk
func (db *DBEngine) loadData() error {
	dataFile := filepath.Join(db.dataDir, "chrono_db.json")
	file, err := os.Open(dataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No data file yet, start fresh
		}
		return fmt.Errorf("failed to open data file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&db.data)
}

// Close closes the database
func (db *DBEngine) Close() error {
	return db.persistData()
}
