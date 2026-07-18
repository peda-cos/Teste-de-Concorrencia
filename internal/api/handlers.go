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
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	balance, err := h.svc.Credit()
	writeBalance(w, balance, err)
}

// Debit handles POST /debit.
func (h *Handler) Debit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	balance, err := h.svc.Debit()
	writeBalance(w, balance, err)
}

// Balance handles GET /balance.
func (h *Handler) Balance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}
	balance, err := h.svc.Balance()
	writeBalance(w, balance, err)
}

func writeBalance(w http.ResponseWriter, balance int, err error) {
	if err != nil {
		log.Printf("account handler error: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]int{"balance": balance})
}

func writeMethodNotAllowed(w http.ResponseWriter, method string) {
	w.Header().Set("allow", method)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

// LoadTestStart handles POST /load-test/start.
func (h *Handler) LoadTestStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}
	if h.orchestrator == nil {
		writeJSONError(w, http.StatusServiceUnavailable, "load test unavailable")
		return
	}

	var groups []loadtest.Group
	if err := json.NewDecoder(r.Body).Decode(&groups); err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid json")
		return
	}

	report, err := h.orchestrator.Run(groups)
	if err != nil {
		log.Printf("load test error: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	w.Header().Set("content-type", "application/json")
	_ = json.NewEncoder(w).Encode(report)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
