package account

import (
	"database/sql"
	"fmt"
	"sync"
)

// Broadcaster receives the new balance after every committed mutation.
type Broadcaster interface {
	Broadcast(balance int)
}

// Service protects the account read-modify-write cycle with a sync.Mutex and
// persists every change to SQLite before responding.
type Service struct {
	mu          sync.Mutex
	db          *sql.DB
	broadcaster Broadcaster
}

// NewService returns an account service backed by db. If broadcaster is non-nil,
// it is called with the new balance immediately after each committed transaction.
func NewService(db *sql.DB, broadcaster Broadcaster) *Service {
	return &Service{db: db, broadcaster: broadcaster}
}

// Credit adds one unit to the shared account and returns the new balance.
func (s *Service) Credit() (int, error) {
	return s.apply(1)
}

// Debit subtracts one unit from the shared account and returns the new balance.
func (s *Service) Debit() (int, error) {
	return s.apply(-1)
}

// Balance returns the current persisted account balance.
// Does not acquire the mutex: SQLite WAL mode provides consistent reads
// via MVCC, so a concurrent write does not corrupt the read result.
func (s *Service) Balance() (int, error) {
	var balance int
	if err := s.db.QueryRow(`SELECT balance FROM accounts WHERE id = 1`).Scan(&balance); err != nil {
		return 0, fmt.Errorf("read balance: %w", err)
	}
	return balance, nil
}

func (s *Service) apply(delta int) (int, error) {
	s.mu.Lock()

	tx, err := s.db.Begin()
	if err != nil {
		s.mu.Unlock()
		return 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE accounts SET balance = balance + ? WHERE id = 1`, delta); err != nil {
		s.mu.Unlock()
		return 0, fmt.Errorf("update balance: %w", err)
	}

	var balance int
	if err := tx.QueryRow(`SELECT balance FROM accounts WHERE id = 1`).Scan(&balance); err != nil {
		s.mu.Unlock()
		return 0, fmt.Errorf("select balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		s.mu.Unlock()
		return 0, fmt.Errorf("commit transaction: %w", err)
	}

	// Release the mutex before broadcasting — Broadcast can block on
	// slow WebSocket clients and must not serialize subsequent operations.
	s.mu.Unlock()

	if s.broadcaster != nil {
		s.broadcaster.Broadcast(balance)
	}

	return balance, nil
}
