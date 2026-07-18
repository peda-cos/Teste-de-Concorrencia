package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"concorrencia/internal/account"
	"concorrencia/internal/api"
	"concorrencia/internal/db"
	"concorrencia/internal/loadtest"
	"concorrencia/internal/ws"

	"github.com/gorilla/websocket"
)

func TestWebSocket_broadcasts_balance_after_credit(t *testing.T) {
	dir := t.TempDir()
	conn, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	hub := ws.NewHub()
	svc := account.NewService(conn, hub)
	h := api.New(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/credit", h.Credit)
	mux.Handle("/ws", hub)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	wsConn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial ws: %v", err)
	}
	defer wsConn.Close()

	resp, err := http.Post(srv.URL+"/credit", "application/json", nil)
	if err != nil {
		t.Fatalf("post credit: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("credit status = %d, want 200", resp.StatusCode)
	}

	var httpBody struct {
		Balance int `json:"balance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&httpBody); err != nil {
		t.Fatalf("decode credit response: %v", err)
	}

	if err := wsConn.SetReadDeadline(time.Now().Add(2 * time.Second)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	_, msg, err := wsConn.ReadMessage()
	if err != nil {
		t.Fatalf("read ws message: %v", err)
	}

	var wsBody struct {
		Balance int `json:"balance"`
	}
	if err := json.Unmarshal(msg, &wsBody); err != nil {
		t.Fatalf("unmarshal ws message: %v", err)
	}

	if httpBody.Balance != 1 {
		t.Fatalf("http balance = %d, want 1", httpBody.Balance)
	}
	if wsBody.Balance != httpBody.Balance {
		t.Fatalf("ws balance %d != http balance %d", wsBody.Balance, httpBody.Balance)
	}
}

func TestLoadTest_two_group_report(t *testing.T) {
	dir := t.TempDir()
	conn, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer conn.Close()

	hub := ws.NewHub()
	svc := account.NewService(conn, hub)
	h := api.New(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/credit", h.Credit)
	mux.HandleFunc("/debit", h.Debit)
	mux.HandleFunc("/balance", h.Balance)
	mux.HandleFunc("/load-test/start", h.LoadTestStart)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	h.SetOrchestrator(loadtest.New(srv.URL))

	body, err := json.Marshal([]loadtest.Group{
		{Users: 3, Requests: 20, Type: "credit"},
		{Users: 2, Requests: 50, Type: "debit"},
	})
	if err != nil {
		t.Fatalf("marshal groups: %v", err)
	}
	resp, err := http.Post(srv.URL+"/load-test/start", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("start load test: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var report loadtest.Report
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		t.Fatalf("decode report: %v", err)
	}

	if report.TotalRequests != 160 {
		t.Fatalf("totalRequests = %d, want 160", report.TotalRequests)
	}
	if report.Successes != 160 {
		t.Fatalf("successes = %d, want 160", report.Successes)
	}
	if report.Failures != 0 {
		t.Fatalf("failures = %d, want 0", report.Failures)
	}
	if report.FinalBalance != -40 {
		t.Fatalf("finalBalance = %d, want -40", report.FinalBalance)
	}
}
