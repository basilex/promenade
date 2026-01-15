package checkbox

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient_Defaults(t *testing.T) {
	client := NewClient(&Config{APIKey: "test-key"})

	if client.baseURL != DefaultBaseURL {
		t.Fatalf("expected baseURL %q, got %q", DefaultBaseURL, client.baseURL)
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Fatalf("expected timeout %s, got %s", (30 * time.Second), client.httpClient.Timeout)
	}
}

func TestNewClient_Sandbox(t *testing.T) {
	client := NewClient(&Config{APIKey: "test-key", Sandbox: true})

	if client.baseURL != SandboxBaseURL {
		t.Fatalf("expected baseURL %q, got %q", SandboxBaseURL, client.baseURL)
	}
}

func TestClient_CreateReceipt_Success(t *testing.T) {
	createdAt := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/receipts/sell" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		var req Receipt
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if len(req.Goods) != 1 || req.Payment.Value != 10000 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		resp := ReceiptResponse{
			ID:         "rcpt-1",
			Type:       "sell",
			Status:     "done",
			FiscalCode: "123456",
			FiscalURL:  "https://tax.gov.ua/receipt/123",
			QRCodeURL:  "https://tax.gov.ua/qr/123",
			CreatedAt:  createdAt,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	client := newTestClient(t, server)

	receipt := &Receipt{
		Goods: []ReceiptGood{{
			Good: Good{
				Name:  "Coffee",
				Price: 10000,
				Tax:   []int{20},
			},
			Quantity: 1,
		}},
		Payment: Payment{Type: "CASH", Value: 10000},
	}

	result, err := client.CreateReceipt(context.Background(), receipt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "rcpt-1" || result.FiscalCode != "123456" {
		t.Fatalf("unexpected receipt response: %#v", result)
	}

	if !result.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected created_at %s, got %s", createdAt, result.CreatedAt)
	}
}

func TestClient_GetReceipt_Success(t *testing.T) {
	createdAt := time.Date(2026, 1, 15, 11, 0, 0, 0, time.UTC)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path != "/receipts/rcpt-42" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		resp := ReceiptResponse{
			ID:        "rcpt-42",
			Type:      "sell",
			Status:    "done",
			CreatedAt: createdAt,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}))
	defer server.Close()

	client := newTestClient(t, server)

	result, err := client.GetReceipt(context.Background(), "rcpt-42")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ID != "rcpt-42" {
		t.Fatalf("unexpected receipt ID: %s", result.ID)
	}

	if !result.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected created_at %s, got %s", createdAt, result.CreatedAt)
	}
}

func TestClient_ValidateCredentials_Invalid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/cashier/me" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Message: "unauthorized"})
	}))
	defer server.Close()

	client := newTestClient(t, server)

	err := client.ValidateCredentials(context.Background())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	if err.Error() == "unauthorized" {
		t.Fatalf("expected wrapped error, got %v", err)
	}

	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected ErrorResponse, got %T", err)
	}

	if apiErr.Code != http.StatusUnauthorized {
		t.Fatalf("expected status code %d, got %d", http.StatusUnauthorized, apiErr.Code)
	}
}

func TestClient_CreateReceipt_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(ErrorResponse{Message: "invalid"})
	}))
	defer server.Close()

	client := newTestClient(t, server)

	_, err := client.CreateReceipt(context.Background(), &Receipt{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}

	var apiErr *ErrorResponse
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected ErrorResponse, got %T", err)
	}

	if apiErr.Code != http.StatusBadRequest {
		t.Fatalf("expected status code %d, got %d", http.StatusBadRequest, apiErr.Code)
	}
}

func newTestClient(t *testing.T, server *httptest.Server) *Client {
	t.Helper()

	client := NewClient(&Config{
		APIKey:  "test-key",
		Sandbox: true,
		Timeout: 2 * time.Second,
	})

	client.baseURL = server.URL
	client.httpClient = server.Client()
	client.httpClient.Timeout = 2 * time.Second

	return client
}