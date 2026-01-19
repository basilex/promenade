package printer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/require"

	"github.com/basilex/promenade/internal/contexts/fiscal/receipt/aggregate"
	"github.com/basilex/promenade/pkg/fiscal/checkbox"
	"github.com/basilex/promenade/pkg/jsonstore"
	"github.com/basilex/promenade/pkg/uuidv7"
)

func setCheckboxClientTransport(t *testing.T, client *checkbox.Client, baseURL string, httpClient *http.Client) {
	t.Helper()

	value := reflect.ValueOf(client).Elem()
	baseURLField := value.FieldByName("baseURL")
	httpClientField := value.FieldByName("httpClient")

	reflect.NewAt(baseURLField.Type(), unsafe.Pointer(baseURLField.UnsafeAddr())).Elem().SetString(baseURL)
	reflect.NewAt(httpClientField.Type(), unsafe.Pointer(httpClientField.UnsafeAddr())).Elem().Set(reflect.ValueOf(httpClient))
}

func TestCheckboxPrinter_Print_NoClient(t *testing.T) {
	printer := NewCheckboxPrinter(nil)

	rec, err := aggregate.NewReceipt(
		uuidv7.New(),
		uuidv7.New(),
		aggregate.PaymentTypeCash,
		aggregate.ReceiptTypeSale,
		"UAH",
		[]aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}},
		uuidv7.New(),
	)
	require.NoError(t, err)

	result, err := printer.Print(context.Background(), rec)
	require.Error(t, err)
	require.Nil(t, result)
}

func TestCheckboxPrinter_Print_InvalidPaymentType(t *testing.T) {
	printer := NewCheckboxPrinter(&checkbox.Client{})

	lines := jsonstore.Field[[]aggregate.ReceiptLine]{}
	lines.Set([]aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}})
	rec := &aggregate.Receipt{
		PaymentType: aggregate.PaymentType("invalid"),
		TotalAmount: 1000,
		Lines:       lines,
	}

	result, err := printer.Print(context.Background(), rec)
	require.Error(t, err)
	require.Nil(t, result)
}

func TestCheckboxPrinter_Print_ClientError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"failure","code":500}`))
	}))
	defer server.Close()

	client := checkbox.NewClient(&checkbox.Config{APIKey: "test", Sandbox: true})
	setCheckboxClientTransport(t, client, server.URL, server.Client())

	printer := NewCheckboxPrinter(client)
	lines := jsonstore.Field[[]aggregate.ReceiptLine]{}
	lines.Set([]aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}})
	rec := &aggregate.Receipt{
		PaymentType: aggregate.PaymentTypeCash,
		TotalAmount: 1000,
		Lines:       lines,
	}

	result, err := printer.Print(context.Background(), rec)
	require.Error(t, err)
	require.Nil(t, result)
}

func TestCheckboxPrinter_Print_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"1","type":"sell","status":"done","fiscal_code":"FC-1","fiscal_url":"https://example.com/receipt","qrcode_url":"https://example.com/qr","created_at":"2026-01-01T00:00:00Z"}`))
	}))
	defer server.Close()

	client := checkbox.NewClient(&checkbox.Config{APIKey: "test", Sandbox: true})
	setCheckboxClientTransport(t, client, server.URL, server.Client())

	printer := NewCheckboxPrinter(client)
	lines := jsonstore.Field[[]aggregate.ReceiptLine]{}
	lines.Set([]aggregate.ReceiptLine{{Name: "Item", Quantity: 1, PriceCents: 1000, TaxRate: 20}})
	rec := &aggregate.Receipt{
		PaymentType: aggregate.PaymentTypeCash,
		TotalAmount: 1000,
		Lines:       lines,
	}

	result, err := printer.Print(context.Background(), rec)
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "1", result.ProviderReceiptID)
	require.Equal(t, "FC-1", result.FiscalNumber)
	require.Equal(t, "https://example.com/receipt", result.FiscalURL)
	require.Equal(t, "https://example.com/qr", result.QRCode)
}

func TestCheckboxPrinter_Cancel_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if r.URL.Path != "/receipts/rcpt-1/cancel" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":"rcpt-1","status":"cancelled"}`))
	}))
	defer server.Close()

	client := checkbox.NewClient(&checkbox.Config{APIKey: "test", Sandbox: true})
	setCheckboxClientTransport(t, client, server.URL, server.Client())

	printer := NewCheckboxPrinter(client)
	rec := &aggregate.Receipt{ProviderReceiptID: "rcpt-1"}

	err := printer.Cancel(context.Background(), rec, "customer request")
	require.NoError(t, err)
}
