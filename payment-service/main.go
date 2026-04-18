package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Payment struct {
	ID      string `json:"id"`
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
}

type Store struct {
	mu       sync.RWMutex
	payments map[string]Payment
	next     int
}

func NewStore() *Store {
	return &Store{
		payments: make(map[string]Payment),
		next:     1,
	}
}

func (s *Store) Add(orderID string) Payment {
	s.mu.Lock()
	defer s.mu.Unlock()
	payment := Payment{
		ID:      fmt.Sprintf("%d", s.next),
		OrderID: orderID,
		Status:  "processed",
	}
	s.payments[payment.ID] = payment
	s.next++
	return payment
}

func (s *Store) Get(id string) (Payment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	payment, ok := s.payments[id]
	return payment, ok
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func readyHandler(ready *atomic.Bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !ready.Load() {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not ready"})
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	}
}

func createPaymentHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			OrderID string `json:"order_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		if body.OrderID == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "order_id required"})
			return
		}
		payment := store.Add(body.OrderID)
		writeJSON(w, http.StatusCreated, payment)
	}
}

func getPaymentHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/payments/")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing id"})
			return
		}
		payment, ok := store.Get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "payment not found"})
			return
		}
		writeJSON(w, http.StatusOK, payment)
	}
}

func configHandler(cfg map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, cfg)
	}
}

func main() {
	store := NewStore()
	port := getEnv("PORT", "8081")
	logLevel := getEnv("LOG_LEVEL", "info")
	dbURL := getEnv("DATABASE_URL", "")
	apiKey := getEnv("API_KEY", "")

	log.Printf("config: LOG_LEVEL=%s DATABASE_URL_SET=%v API_KEY_SET=%v",
		logLevel,
		dbURL != "",
		apiKey != "",
	)

	var ready atomic.Bool

	cfg := map[string]string{
		"service":     "payment-service",
		"port":        port,
		"log_level":   logLevel,
		"has_db_url":  fmt.Sprintf("%t", dbURL != ""),
		"has_api_key": fmt.Sprintf("%t", apiKey != ""),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ready", readyHandler(&ready))
	mux.HandleFunc("/config", configHandler(cfg))
	mux.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createPaymentHandler(store)(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})
	mux.HandleFunc("/payments/", getPaymentHandler(store))

	go func() {
		time.Sleep(2 * time.Second)
		ready.Store(true)
		log.Printf("payment-service ready")
	}()

	log.Printf("payment-service starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
