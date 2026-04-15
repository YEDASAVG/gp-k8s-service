package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
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

func main() {
	store := NewStore()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/payments", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			createPaymentHandler(store)(w, r)
		default:
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		}
	})
	mux.HandleFunc("/payments/", getPaymentHandler(store))

	log.Println("payment-service starting on :8081")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatal(err)
	}
}
