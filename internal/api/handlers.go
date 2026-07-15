package api

import (
	"encoding/json"
	"log"
	"net/http"

	"concorrencia/internal/account"
	"concorrencia/internal/loadtest"
)

// Handler exposes the account operations over HTTP.
type Handler struct {
	svc          *account.Service
	orchestrator *loadtest.Orchestrator
}

// New returns an HTTP handler backed by svc.
func New(svc *account.Service) *Handler {
	return &Handler{svc: svc}
}

// SetOrchestrator wires the load-test orchestrator into the handler.
func (h *Handler) SetOrchestrator(o *loadtest.Orchestrator) {
	h.orchestrator = o
}

// Credit handles POST /credit.
func (h *Handler) Credit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	balance, err := h.svc.Credit()
	writeBalance(w, balance, err)
}

// Debit handles POST /debit.
func (h *Handler) Debit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	balance, err := h.svc.Debit()
	writeBalance(w, balance, err)
}

// Balance handles GET /balance.
func (h *Handler) Balance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	balance, err := h.svc.Balance()
	writeBalance(w, balance, err)
}

func writeBalance(w http.ResponseWriter, balance int, err error) {
	if err != nil {
		log.Printf("account handler error: %v", err)
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{"balance": balance})
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	w.Header().Set("allow", "POST")
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

// LoadTestStart handles POST /load-test/start.
func (h *Handler) LoadTestStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}
	if h.orchestrator == nil {
		http.Error(w, `{"error":"load test unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	var groups []loadtest.Group
	if err := json.NewDecoder(r.Body).Decode(&groups); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}

	report, err := h.orchestrator.Run(groups)
	if err != nil {
		log.Printf("load test error: %v", err)
		http.Error(w, fmtError(err), http.StatusBadRequest)
		return
	}

	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}

func fmtError(err error) string {
	b, _ := json.Marshal(map[string]string{"error": err.Error()})
	return string(b)
}
