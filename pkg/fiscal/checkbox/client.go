// Package checkbox provides Checkbox API integration for Ukrainian fiscal compliance (ПРРО)
// API Documentation: https://dev.checkbox.ua/uk/docs/api/
package checkbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	// DefaultBaseURL is the production Checkbox API endpoint
	DefaultBaseURL = "https://api.checkbox.ua/api/v1"

	// SandboxBaseURL is the test/sandbox endpoint
	SandboxBaseURL = "https://api.sandbox.checkbox.ua/api/v1"
)

// Client represents a Checkbox API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Config contains Checkbox client configuration
type Config struct {
	APIKey  string
	Sandbox bool // Use sandbox environment for testing
	Timeout time.Duration
}

// NewClient creates a new Checkbox API client
func NewClient(cfg *Config) *Client {
	baseURL := DefaultBaseURL
	if cfg.Sandbox {
		baseURL = SandboxBaseURL
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		baseURL: baseURL,
		apiKey:  cfg.APIKey,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Receipt represents a fiscal receipt request
type Receipt struct {
	Goods       []ReceiptGood `json:"goods"`
	Payment     Payment       `json:"payment"`
	Delivery    *Delivery     `json:"delivery,omitempty"`
	TaxNumber   string        `json:"tax_number,omitempty"`   // ІПН покупця (optional)
	Header      string        `json:"header,omitempty"`        // Text at top of receipt
	Footer      string        `json:"footer,omitempty"`        // Text at bottom of receipt
}

// ReceiptGood represents a line item in the receipt
type ReceiptGood struct {
	Good     Good `json:"good"`
	Quantity int  `json:"quantity"` // In units (e.g., 1000 = 1 item)
}

// Good represents a product/service
type Good struct {
	Code  string `json:"code"`            // Product code (optional)
	Name  string `json:"name"`            // Product name (required)
	Price int    `json:"price"`           // Price in kopiyky (required)
	Tax   []int  `json:"tax,omitempty"`   // Tax rates (e.g., [20] for 20% VAT)
}

// Payment represents payment information
type Payment struct {
	Type  string `json:"type"`  // "CASH", "CARD", "CASHLESS"
	Value int    `json:"value"` // Total amount in kopiyky
}

// Delivery represents delivery information (optional)
type Delivery struct {
	Email string `json:"email,omitempty"`
}

// ReceiptResponse represents the response from creating a receipt
type ReceiptResponse struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Status     string    `json:"status"`
	FiscalCode string    `json:"fiscal_code"` // Фіскальний номер
	FiscalURL  string    `json:"fiscal_url"`  // URL to verify receipt
	QRCodeURL  string    `json:"qrcode_url"`  // QR code image URL
	CreatedAt  time.Time `json:"created_at"`
}

// ErrorResponse represents an error response from Checkbox API
type ErrorResponse struct {
	Message string                 `json:"message"`
	Errors  map[string]interface{} `json:"errors,omitempty"`
	Code    int                    `json:"code"`
}

func (e *ErrorResponse) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("checkbox API error: %s (code: %d)", e.Message, e.Code)
	}
	return fmt.Sprintf("checkbox API error: code %d", e.Code)
}

// CreateReceipt creates a new fiscal receipt
func (c *Client) CreateReceipt(ctx context.Context, receipt *Receipt) (*ReceiptResponse, error) {
	var result ReceiptResponse
	if err := c.post(ctx, "/receipts/sell", receipt, &result); err != nil {
		return nil, fmt.Errorf("failed to create receipt: %w", err)
	}
	return &result, nil
}

// GetReceipt retrieves receipt by ID
func (c *Client) GetReceipt(ctx context.Context, receiptID string) (*ReceiptResponse, error) {
	var result ReceiptResponse
	if err := c.get(ctx, fmt.Sprintf("/receipts/%s", receiptID), &result); err != nil {
		return nil, fmt.Errorf("failed to get receipt: %w", err)
	}
	return &result, nil
}

// ValidateCredentials validates API credentials by making a test request
func (c *Client) ValidateCredentials(ctx context.Context) error {
	// Try to get cashier info (lightweight endpoint)
	if err := c.get(ctx, "/cashier/me", nil); err != nil {
		return fmt.Errorf("invalid API credentials: %w", err)
	}
	return nil
}

// post makes a POST request to the Checkbox API
func (c *Client) post(ctx context.Context, path string, body, result interface{}) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	return c.doRequest(req, result)
}

// get makes a GET request to the Checkbox API
func (c *Client) get(ctx context.Context, path string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	return c.doRequest(req, result)
}

// doRequest executes HTTP request and handles response
func (c *Client) doRequest(req *http.Request, result interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log error but don't fail the request
			_ = closeErr
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(body, &errResp); err != nil {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
		}
		errResp.Code = resp.StatusCode
		return &errResp
	}

	// Parse result if provided
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}
