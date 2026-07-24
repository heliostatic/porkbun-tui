package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestClient(serverURL string) *Client {
	return &Client{
		apiKey:     "pk1_test",
		secretKey:  "sk1_test",
		baseURL:    serverURL,
		httpClient: &http.Client{},
	}
}

func TestCheckAvailabilityAvailable(t *testing.T) {
	var gotPath, gotMethod, gotContentType string
	var gotBody map[string]string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotContentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding request body: %v", err)
		}
		w.Write([]byte(`{"status":"SUCCESS","response":{"avail":"yes","type":"registration","price":"11.06","premium":"no"}}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	result, err := c.CheckAvailability(context.Background(), "example.com")
	if err != nil {
		t.Fatalf("CheckAvailability returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/domain/checkDomain/example.com" {
		t.Errorf("path = %q, want /domain/checkDomain/example.com", gotPath)
	}
	if gotBody["apikey"] != "pk1_test" || gotBody["secretapikey"] != "sk1_test" {
		t.Errorf("request body missing credentials: %v", gotBody)
	}
	if gotContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", gotContentType)
	}
	if !result.Available {
		t.Error("result.Available = false, want true")
	}
	if result.Domain != "example.com" {
		t.Errorf("result.Domain = %q, want example.com", result.Domain)
	}
	if result.Price != "11.06" {
		t.Errorf("result.Price = %q, want 11.06", result.Price)
	}
	if result.Premium {
		t.Error("result.Premium = true, want false")
	}
}

func TestCheckAvailabilityTakenAndPremium(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"SUCCESS","response":{"avail":"no","price":"1200.00","premium":"yes"}}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	result, err := c.CheckAvailability(context.Background(), "taken.com")
	if err != nil {
		t.Fatalf("CheckAvailability returned error: %v", err)
	}
	if result.Available {
		t.Error("result.Available = true, want false")
	}
	if !result.Premium {
		t.Error("result.Premium = false, want true")
	}
}

func TestCheckAvailabilityAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"status":"ERROR","message":"Invalid API key. (002)"}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	_, err := c.CheckAvailability(context.Background(), "example.com")
	if err == nil {
		t.Fatal("CheckAvailability returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("error %q does not contain API message", err)
	}
}

func TestCheckAvailabilityMissingAvailFieldIsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"SUCCESS"}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	result, err := c.CheckAvailability(context.Background(), "example.com")
	if err == nil {
		t.Fatalf("CheckAvailability returned nil error for SUCCESS body without avail field, got result %+v", result)
	}
}

func TestCheckAvailabilityNonJSONErrorIncludesHTTPStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`<html><body>Bad Gateway</body></html>`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	_, err := c.CheckAvailability(context.Background(), "example.com")
	if err == nil {
		t.Fatal("CheckAvailability returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "502") {
		t.Errorf("error %q does not mention HTTP status 502", err)
	}
}

func TestCheckAvailabilityEscapesDomainInPath(t *testing.T) {
	var gotEscapedPath, gotRawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEscapedPath = r.URL.EscapedPath()
		gotRawQuery = r.URL.RawQuery
		w.Write([]byte(`{"status":"SUCCESS","response":{"avail":"no","price":"","premium":"no"}}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	if _, err := c.CheckAvailability(context.Background(), "a/b?x=1"); err != nil {
		t.Fatalf("CheckAvailability returned error: %v", err)
	}

	if gotEscapedPath != "/domain/checkDomain/a%2Fb%3Fx=1" {
		t.Errorf("escaped path = %q, want /domain/checkDomain/a%%2Fb%%3Fx=1", gotEscapedPath)
	}
	if gotRawQuery != "" {
		t.Errorf("query = %q, want empty (input must not become a query string)", gotRawQuery)
	}
}

func TestDollarsToCents(t *testing.T) {
	valid := map[string]int{
		"2.04":    204,
		"11.06":   1106,
		"1200.00": 120000,
		"7":       700,
		"7.5":     750,
		"0.99":    99,
		" 3.10 ":  310,
	}
	for in, want := range valid {
		got, err := DollarsToCents(in)
		if err != nil {
			t.Errorf("DollarsToCents(%q) error: %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("DollarsToCents(%q) = %d, want %d", in, got, want)
		}
	}

	invalid := []string{
		"", "abc", "1.2.3", "-1.00", "-0.50", "1.234", "$2.04", ".",
		// A price of zero must never arm a "Buy for $0.00" purchase.
		"0", "0.0", "0.00",
		// Sign characters and huge values must error, not wrap or mislead:
		// Atoi accepts "+5", and unbounded parsing integer-wraps
		// "184467440737095516.16" to 0 and similar values to negatives.
		"+5", "1.+5",
		"184467440737095516.16", "92233720368547759.00", "12345678901234567890.00",
	}
	for _, in := range invalid {
		if got, err := DollarsToCents(in); err == nil {
			t.Errorf("DollarsToCents(%q) = %d, want error", in, got)
		}
	}
}

func TestRegisterDomainSuccess(t *testing.T) {
	var gotPath, gotMethod string
	var gotBody map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Errorf("decoding request body: %v", err)
		}
		w.Write([]byte(`{"status":"SUCCESS","domain":"newdomain.com","cost":868,"orderId":123456,"balance":5000}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	result, err := c.RegisterDomain(context.Background(), "newdomain.com", 868)
	if err != nil {
		t.Fatalf("RegisterDomain returned error: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Errorf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/domain/create/newdomain.com" {
		t.Errorf("path = %q, want /domain/create/newdomain.com", gotPath)
	}
	if gotBody["apikey"] != "pk1_test" || gotBody["secretapikey"] != "sk1_test" {
		t.Errorf("request body missing credentials: %v", gotBody)
	}
	// JSON numbers decode as float64; the cost must be sent as a number, not a string.
	if cost, ok := gotBody["cost"].(float64); !ok || cost != 868 {
		t.Errorf("request cost = %v (%T), want number 868", gotBody["cost"], gotBody["cost"])
	}
	if gotBody["agreeToTerms"] != "yes" {
		t.Errorf("agreeToTerms = %v, want \"yes\"", gotBody["agreeToTerms"])
	}

	if result.Domain != "newdomain.com" {
		t.Errorf("result.Domain = %q, want newdomain.com", result.Domain)
	}
	if result.OrderID != 123456 {
		t.Errorf("result.OrderID = %d, want 123456", result.OrderID)
	}
	if result.CostCents != 868 {
		t.Errorf("result.CostCents = %d, want 868", result.CostCents)
	}
	if result.BalanceCents != 5000 {
		t.Errorf("result.BalanceCents = %d, want 5000", result.BalanceCents)
	}
}

func TestRegisterDomainAcceptsInCentsFieldNames(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"SUCCESS","domain":"newdomain.com","costInCents":868,"orderId":9,"balanceInCents":5000}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	result, err := c.RegisterDomain(context.Background(), "newdomain.com", 868)
	if err != nil {
		t.Fatalf("RegisterDomain returned error: %v", err)
	}
	if result.CostCents != 868 || result.BalanceCents != 5000 {
		t.Errorf("cost/balance = %d/%d, want 868/5000 from *InCents fields", result.CostCents, result.BalanceCents)
	}
}

func TestRegisterDomainEscapesDomainInPath(t *testing.T) {
	var gotEscapedPath, gotRawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotEscapedPath = r.URL.EscapedPath()
		gotRawQuery = r.URL.RawQuery
		w.Write([]byte(`{"status":"SUCCESS","domain":"x","cost":1,"orderId":1,"balance":1}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	if _, err := c.RegisterDomain(context.Background(), "a/b?x=1", 100); err != nil {
		t.Fatalf("RegisterDomain returned error: %v", err)
	}
	if gotEscapedPath != "/domain/create/a%2Fb%3Fx=1" {
		t.Errorf("escaped path = %q, want /domain/create/a%%2Fb%%3Fx=1", gotEscapedPath)
	}
	if gotRawQuery != "" {
		t.Errorf("query = %q, want empty", gotRawQuery)
	}
}

func TestRegisterDomainAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"status":"ERROR","message":"Domain is not available."}`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	_, err := c.RegisterDomain(context.Background(), "taken.com", 868)
	if err == nil {
		t.Fatal("RegisterDomain returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "Domain is not available") {
		t.Errorf("error %q does not contain API message", err)
	}
}

func TestRegisterDomainTransportErrorWarnsPurchaseMayHaveCompleted(t *testing.T) {
	// A timeout or connection failure does NOT mean the charge didn't go
	// through server-side; the error must warn before the user retries.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close() // connection refused

	c := newTestClient(server.URL)
	_, err := c.RegisterDomain(context.Background(), "example.com", 868)
	if err == nil {
		t.Fatal("RegisterDomain returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "may still have completed") {
		t.Errorf("transport error %q does not warn that the purchase may have completed", err)
	}
}

func TestRegisterDomainNonJSONErrorIncludesHTTPStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`<html>down</html>`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	_, err := c.RegisterDomain(context.Background(), "newdomain.com", 868)
	if err == nil {
		t.Fatal("RegisterDomain returned nil error, want error")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("error %q does not mention HTTP status 503", err)
	}
}

func TestCheckAvailabilityMalformedResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	c := newTestClient(server.URL)
	_, err := c.CheckAvailability(context.Background(), "example.com")
	if err == nil {
		t.Fatal("CheckAvailability returned nil error, want error")
	}
}
