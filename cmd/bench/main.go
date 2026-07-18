package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"concorrencia/internal/account"
	"concorrencia/internal/api"
	"concorrencia/internal/db"
	"concorrencia/internal/loadtest"
	"concorrencia/internal/ws"

	"github.com/gorilla/websocket"
)

const (
	wsClients     = 5
	creditUsers   = 20
	creditReqs    = 500
	debitUsers    = 20
	debitReqs     = 500
	wsDialTimeout = 5 * time.Second
)

func main() {
	log.SetOutput(os.Stderr)

	dir, err := os.MkdirTemp("", "concorrencia-bench")
	if err != nil {
		log.Fatalf("temp dir: %v", err)
	}
	defer os.RemoveAll(dir)

	conn, err := db.Open(dir + "/bench.db")
	if err != nil {
		log.Fatalf("db open: %v", err)
	}
	defer conn.Close()

	hub := ws.NewHub()
	svc := account.NewService(conn, hub)
	h := api.New(svc)

	mux := http.NewServeMux()
	mux.HandleFunc("/credit", h.Credit)
	mux.HandleFunc("/debit", h.Debit)
	mux.HandleFunc("/balance", h.Balance)
	mux.Handle("/ws", hub)
	mux.HandleFunc("/load-test/start", h.LoadTestStart)

	srv := httptest.NewServer(mux)
	defer srv.Close()

	h.SetOrchestrator(loadtest.New(srv.URL).WithClient(&http.Client{
		Timeout: 10 * time.Second,
		Transport: &muxTransport{mux},
	}))

	// Connect WebSocket clients to exercise the broadcast path.
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"
	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = wsDialTimeout

	var wsConns []*websocket.Conn
	for range wsClients {
		c, _, err := dialer.Dial(wsURL, nil)
		if err != nil {
			log.Fatalf("ws dial: %v", err)
		}
		defer c.Close()
		wsConns = append(wsConns, c)
	}

	// Drain WebSocket messages in background so TCP buffers don't fill.
	for _, c := range wsConns {
		go func(conn *websocket.Conn) {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					return
				}
			}
		}(c)
	}

	groups := []loadtest.Group{
		{Users: creditUsers, Requests: creditReqs, Type: "credit"},
		{Users: debitUsers, Requests: debitReqs, Type: "debit"},
	}

	body, err := json.Marshal(groups)
	if err != nil {
		log.Fatalf("marshal groups: %v", err)
	}

	start := time.Now()
	resp, err := http.Post(srv.URL+"/load-test/start", "application/json", bytes.NewReader(body))
	if err != nil {
		log.Fatalf("post load test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("load test status: %d", resp.StatusCode)
	}

	var report loadtest.Report
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		log.Fatalf("decode report: %v", err)
	}

	elapsed := time.Since(start)
	throughput := float64(report.TotalRequests) / elapsed.Seconds()

	expectedBalance := creditUsers*creditReqs - debitUsers*debitReqs
	balanceError := report.FinalBalance - expectedBalance

	// Primary metric: operations per second.
	fmt.Printf("METRIC throughput=%.2f\n", throughput)
	// Secondary: how far the final balance deviated from the expected value.
	fmt.Printf("METRIC balance_error=%d\n", balanceError)
	fmt.Printf("METRIC total_requests=%d\n", report.TotalRequests)
	fmt.Printf("METRIC successes=%d\n", report.Successes)
	fmt.Printf("METRIC failures=%d\n", report.Failures)
	fmt.Printf("METRIC duration_ms=%d\n", report.DurationMs)
	fmt.Printf("METRIC final_balance=%d\n", report.FinalBalance)
}

// muxTransport is an http.RoundTripper that bypasses TCP and serves requests
// directly to an http.Handler in-process.  This isolates pure handler / service /
// database performance from loopback TCP overhead.
type muxTransport struct {
	handler http.Handler
}

func (t *muxTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	w := httptest.NewRecorder()
	t.handler.ServeHTTP(w, req)
	return w.Result(), nil
}
