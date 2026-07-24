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
