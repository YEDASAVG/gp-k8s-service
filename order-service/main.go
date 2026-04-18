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

type Order struct {
	ID       string `json:"id"`
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
}

type Store struct {
	mu     sync.RWMutex
	orders map[string]Order
	next   int
}

func NewStore() *Store {
	return &Store{
		orders: make(map[string]Order),
		next:   1,
	}
}

func (s *Store) Add(item string, quantity int) Order {
	s.mu.Lock()
	defer s.mu.Unlock()
	order := Order{
		ID:       fmt.Sprintf("%d", s.next),
		Item:     item,
		Quantity: quantity,
	}
	s.orders[order.ID] = order
	s.next++
	return order
}

func (s *Store) Get(id string) (Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.orders[id]
	return order, ok
}

func (s *Store) List() []Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	list := make([]Order, 0, len(s.orders))
	for _, o := range s.orders {
		list = append(list, o)
	}
	return list
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

func listOrdersHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, store.List())
	}
}

func createOrderHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Item     string `json:"item"`
			Quantity int    `json:"quantity"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request"})
			return
		}
		if body.Item == "" || body.Quantity <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "item and quantity required"})
			return
		}
		order := store.Add(body.Item, body.Quantity)
		writeJSON(w, http.StatusCreated, order)
	}
}

func getOrderHandler(store *Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/orders/")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing id"})
			return
		}
		order, ok := store.Get(id)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]string{"error": "order not found"})
			return
		}
		writeJSON(w, http.StatusOK, order)
	}
}

func configHandler(cfg map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, cfg)
	}
}

func main() {
	store := NewStore()
	port := getEnv("PORT", "8080")
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
		"service":     "order-service",
		"port":        port,
		"log_level":   logLevel,
		"has_db_url":  fmt.Sprintf("%t", dbURL != ""),
		"has_api_key": fmt.Sprintf("%t", apiKey != ""),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/ready", readyHandler(&ready))
	mux.HandleFunc("/config", configHandler(cfg))
	mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			listOrdersHandler(store)(w, r)
		case http.MethodPost:
			createOrderHandler(store)(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})
	mux.HandleFunc("/orders/", getOrderHandler(store))

	go func() {
		time.Sleep(2 * time.Second)
		ready.Store(true)
		log.Printf("order-service ready")
	}()

	log.Printf("order-service starting on :%s", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		log.Fatal(err)
	}
}
