package loadtest

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestOrchestrator_invalid_group_rejected(t *testing.T) {
	orch := New("http://localhost:1")
	_, err := orch.Run([]Group{{Users: 0, Requests: 1, Type: "credit"}})
	if err == nil {
		t.Fatal("expected error for zero users")
	}

	_, err = orch.Run([]Group{{Users: 1, Requests: 0, Type: "credit"}})
	if err == nil {
		t.Fatal("expected error for zero requests")
	}

	_, err = orch.Run([]Group{{Users: 1, Requests: 1, Type: "transfer"}})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestOrchestrator_client_replacement(t *testing.T) {
	var calls atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.Header().Set("content-type", "application/json")
		if r.URL.Path == "/balance" {
			_ = json.NewEncoder(w).Encode(map[string]int{"balance": 0})
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	orch := New(srv.URL).WithClient(srv.Client())
	report, err := orch.Run([]Group{{Users: 2, Requests: 5, Type: "credit"}})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if n := int(calls.Load()); n != 11 {
		t.Fatalf("server calls = %d, want 11", n)
	}
	if report.TotalRequests != 10 {
		t.Fatalf("totalRequests = %d, want 10", report.TotalRequests)
	}
}
