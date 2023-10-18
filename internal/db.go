package internal

import (
	"strings"
	"sync"
)

type Database struct {
	store map[string]string
	mu    sync.Mutex
}

// NewDatabase creates a new instance of Database struct
func NewDatabase() *Database {
	return &Database{
		store: make(map[string]string),
		mu:    sync.Mutex{},
	}
}

// PerformOperation is responsible for performing the operation on database.
// It also is responsible for validating the operation before applying it on the database.
func (db *Database) PerformOperation(operation string) string {
	db.mu.Lock()
	defer db.mu.Unlock()

	splits := strings.Split(operation, " ")
	if len(splits) == 0 {
		return INVALID_DATABASE_REQUEST
	}
	if splits[0] == "get" {
		return db.performGet(splits)
	} else if splits[0] == "set" {
		return db.performSet(splits)
	} else {
		return INVALID_DATABASE_REQUEST
	}
}

func (db *Database) performGet(splits []string) string {
	if len(splits) != 2 {
		return INVALID_DATABASE_REQUEST
	}
	key := strings.TrimSpace(splits[1])
	val, exists := db.store[key]
	if !exists {
		return VALUE_DOES_NOT_EXIST
	}
	return val
}

func (db *Database) performSet(splits []string) string {
	if len(splits) != 3 {
		return INVALID_DATABASE_REQUEST
	}
	key, val := strings.TrimSpace(splits[1]), strings.TrimSpace(splits[2])
	db.store[key] = val
	return UPDATE_PERFORMED_SUCCESSFULLY
}
