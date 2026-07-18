package main

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"
	"os"

	"concorrencia/internal/account"
	"concorrencia/internal/api"
	"concorrencia/internal/db"
	"concorrencia/internal/loadtest"
	"concorrencia/internal/ws"
)

//go:embed frontend
var frontend embed.FS

func run() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "concorrencia.db"
	}

	conn, err := db.Open(dbPath)
	if err != nil {
		return err
	}
	defer conn.Close()

	hub := ws.NewHub()
	svc := account.NewService(conn, hub)
	h := api.New(svc)
	h.SetOrchestrator(loadtest.New("http://localhost:" + port))

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "application/json")
		_ = json.NewEncoder(w).Encode(struct {
			Status string `json:"status"`
		}{Status: "ok"})
	})
	mux.HandleFunc("/credit", h.Credit)
	mux.HandleFunc("/debit", h.Debit)
	mux.HandleFunc("/balance", h.Balance)
	mux.Handle("/ws", hub)
	mux.HandleFunc("/load-test/start", h.LoadTestStart)

	frontendFS, err := fs.Sub(frontend, "frontend")
	if err != nil {
		return err
	}
	mux.Handle("/", http.FileServer(http.FS(frontendFS)))

	addr := ":" + port
	log.Printf("server listening on http://localhost%s", addr)
	log.Printf("tela 1 (load test)   → http://localhost:%s/", port)
	log.Printf("tela 2 (saldo real)  → http://localhost:%s/saldo.html", port)
	return http.ListenAndServe(addr, mux)
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
