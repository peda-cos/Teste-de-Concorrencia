package account

import (
	"path/filepath"
	"sync"
	"testing"

	"concorrencia/internal/db"
)

func TestService_Credit_increases_balance(t *testing.T) {
	dir := t.TempDir()
	conn, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	svc := NewService(conn, nil)

	tests := []struct {
		calls int
		want  int
	}{
		{calls: 1, want: 1},
		{calls: 5, want: 6},
		{calls: 10, want: 16},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			for i := 0; i < tt.calls; i++ {
				if _, err := svc.Credit(); err != nil {
					t.Fatalf("credit: %v", err)
				}
			}
			got, err := svc.Balance()
			if err != nil {
				t.Fatalf("balance: %v", err)
			}
			if got != tt.want {
				t.Fatalf("balance = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestService_Debit_decreases_balance(t *testing.T) {
	dir := t.TempDir()
	conn, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	svc := NewService(conn, nil)
	if _, err := svc.Credit(); err != nil {
		t.Fatalf("seed credit: %v", err)
	}

	tests := []struct {
		calls int
		want  int
	}{
		{calls: 1, want: 0},
		{calls: 3, want: -3},
		{calls: 7, want: -10},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			for i := 0; i < tt.calls; i++ {
				if _, err := svc.Debit(); err != nil {
					t.Fatalf("debit: %v", err)
				}
			}
			got, err := svc.Balance()
			if err != nil {
				t.Fatalf("balance: %v", err)
			}
			if got != tt.want {
				t.Fatalf("balance = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestService_concurrent_mixed_operations_are_atomic(t *testing.T) {
	dir := t.TempDir()
	conn, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	svc := NewService(conn, nil)

	const credits = 100
	const debits = 100

	var wg sync.WaitGroup
	wg.Add(credits + debits)

	for i := 0; i < credits; i++ {
		go func() {
			defer wg.Done()
			if _, err := svc.Credit(); err != nil {
				t.Errorf("credit: %v", err)
			}
		}()
	}
	for i := 0; i < debits; i++ {
		go func() {
			defer wg.Done()
			if _, err := svc.Debit(); err != nil {
				t.Errorf("debit: %v", err)
			}
		}()
	}

	wg.Wait()

	got, err := svc.Balance()
	if err != nil {
		t.Fatalf("balance: %v", err)
	}
	if got != credits-debits {
		t.Fatalf("balance = %d, want %d", got, credits-debits)
	}
}
