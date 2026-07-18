package loadtest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Group describes one virtual-user cohort.
type Group struct {
	Users    int    `json:"users"`
	Requests int    `json:"requests"`
	Type     string `json:"type"`
}

// Report contains the aggregate results of a load-test run.
type Report struct {
	TotalRequests int   `json:"totalRequests"`
	Successes     int   `json:"successes"`
	Failures      int   `json:"failures"`
	DurationMs    int64 `json:"durationMs"`
	FinalBalance  int   `json:"finalBalance"`
}

// Orchestrator drives configurable virtual-user groups against the local HTTP API.
type Orchestrator struct {
	client  *http.Client
	baseURL string
}

// New returns an orchestrator that targets baseURL (e.g. http://localhost:8080).
func New(baseURL string) *Orchestrator {
	return &Orchestrator{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// Run executes all groups concurrently and returns a report.
func (o *Orchestrator) Run(groups []Group) (*Report, error) {
	if len(groups) == 0 {
		return nil, fmt.Errorf("at least one group is required")
	}

	var total int64
	var successes atomic.Int64
	var failures atomic.Int64
	var wg sync.WaitGroup

	start := time.Now()
	for _, g := range groups {
		if err := validateGroup(g); err != nil {
			return nil, err
		}

		path := "/credit"
		if g.Type == "debit" {
			path = "/debit"
		}

		for range g.Users {
			wg.Add(1)
			total += int64(g.Requests)
			url := o.baseURL + path
			go func(url string, requests int) {
				defer wg.Done()
				for range requests {
					resp, err := o.client.Post(url, "application/json", nil)
					if err != nil {
						failures.Add(1)
						continue
					}
					resp.Body.Close()
					if resp.StatusCode == http.StatusOK {
						successes.Add(1)
					} else {
						failures.Add(1)
					}
				}
			}(url, g.Requests)
		}
	}

	wg.Wait()
	duration := time.Since(start).Milliseconds()

	balance, err := o.fetchBalance()
	if err != nil {
		return nil, fmt.Errorf("fetch final balance: %w", err)
	}

	return &Report{
		TotalRequests: int(total),
		Successes:     int(successes.Load()),
		Failures:      int(failures.Load()),
		DurationMs:    duration,
		FinalBalance:  balance,
	}, nil
}

func validateGroup(g Group) error {
	if g.Users <= 0 {
		return fmt.Errorf("users must be greater than zero")
	}
	if g.Requests <= 0 {
		return fmt.Errorf("requests must be greater than zero")
	}
	if g.Type != "credit" && g.Type != "debit" {
		return fmt.Errorf("type must be credit or debit")
	}
	return nil
}

func (o *Orchestrator) fetchBalance() (int, error) {
	resp, err := o.client.Get(o.baseURL + "/balance")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("fetch balance: http %d", resp.StatusCode)
	}

	var body struct {
		Balance int `json:"balance"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return 0, fmt.Errorf("fetch balance: decode: %w", err)
	}
	return body.Balance, nil
}

// WithClient replaces the default HTTP client. Passing nil is a no-op.
func (o *Orchestrator) WithClient(client *http.Client) *Orchestrator {
	if client != nil {
		o.client = client
	}
	return o
}
