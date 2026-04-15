package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandler(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	healthHandler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 got %d", w.Code)
	}
}

func TestCreatePayment(t *testing.T) {
	store := NewStore()
	handler := createPaymentHandler(store)

	body := `{"order_id": "1"}`
	req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201 got %d", w.Code)
	}

	var payment Payment
	json.NewDecoder(w.Body).Decode(&payment)

	if payment.OrderID != "1" {
		t.Errorf("expected order_id 1 got %s", payment.OrderID)
	}
	if payment.Status != "processed" {
		t.Errorf("expected processed got %s", payment.Status)
	}
}

func TestCreatePayment_MissingOrderID(t *testing.T) {
	store := NewStore()
	handler := createPaymentHandler(store)

	req := httptest.NewRequest(http.MethodPost, "/payments", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 got %d", w.Code)
	}
}

func TestGetPayment(t *testing.T) {
	store := NewStore()
	payment := store.Add("1")

	handler := getPaymentHandler(store)
	req := httptest.NewRequest(http.MethodGet, "/payments/"+payment.ID, nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 got %d", w.Code)
	}
}

func TestGetPayment_NotFound(t *testing.T) {
	store := NewStore()
	handler := getPaymentHandler(store)

	req := httptest.NewRequest(http.MethodGet, "/payments/999", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 got %d", w.Code)
	}
}
