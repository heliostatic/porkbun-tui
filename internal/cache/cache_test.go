package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/bc/porkbun-tui/internal/api"
)

func newTestCache(t *testing.T) *Cache {
	t.Helper()
	dir := t.TempDir()
	return &Cache{dir: dir}
}

func TestCache_SaveAndLoadDomains(t *testing.T) {
	c := newTestCache(t)

	domains := []api.Domain{
		{
			Name:       "example.com",
			TLD:        "com",
			ExpireDate: time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC),
			AutoRenew:  true,
		},
		{
			Name:       "test.io",
			TLD:        "io",
			ExpireDate: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			AutoRenew:  false,
		},
	}

	// Save
	if err := c.SaveDomains(domains); err != nil {
		t.Fatalf("SaveDomains failed: %v", err)
	}

	// Load
	loaded, updatedAt, err := c.LoadDomains()
	if err != nil {
		t.Fatalf("LoadDomains failed: %v", err)
	}

	// Check updatedAt is recent
	if time.Since(updatedAt) > time.Minute {
		t.Errorf("updatedAt too old: %v", updatedAt)
	}

	// Check data
	if len(loaded) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(loaded))
	}

	if loaded[0].Name != "example.com" {
		t.Errorf("expected first domain 'example.com', got '%s'", loaded[0].Name)
	}
	if loaded[1].Name != "test.io" {
		t.Errorf("expected second domain 'test.io', got '%s'", loaded[1].Name)
	}
	if !loaded[0].AutoRenew {
		t.Error("expected first domain AutoRenew to be true")
	}
}

func TestCache_LoadDomains_Empty(t *testing.T) {
	c := newTestCache(t)

	// Load from empty cache should return nil without error
	domains, updatedAt, err := c.LoadDomains()
	if err != nil {
		t.Fatalf("LoadDomains failed: %v", err)
	}

	if domains != nil {
		t.Errorf("expected nil domains, got %v", domains)
	}
	if !updatedAt.IsZero() {
		t.Errorf("expected zero time, got %v", updatedAt)
	}
}

func TestCache_SaveAndLoadPricing(t *testing.T) {
	c := newTestCache(t)

	pricing := map[string]api.TLDPricing{
		"com": {
			TLD:          "com",
			Registration: "9.73",
			Renewal:      "10.37",
			Transfer:     "10.37",
		},
		"io": {
			TLD:          "io",
			Registration: "32.98",
			Renewal:      "32.98",
			Transfer:     "32.98",
		},
	}

	// Save
	if err := c.SavePricing(pricing); err != nil {
		t.Fatalf("SavePricing failed: %v", err)
	}

	// Load
	loaded, updatedAt, err := c.LoadPricing()
	if err != nil {
		t.Fatalf("LoadPricing failed: %v", err)
	}

	// Check updatedAt is recent
	if time.Since(updatedAt) > time.Minute {
		t.Errorf("updatedAt too old: %v", updatedAt)
	}

	// Check data
	if len(loaded) != 2 {
		t.Fatalf("expected 2 TLDs, got %d", len(loaded))
	}

	com, ok := loaded["com"]
	if !ok {
		t.Fatal("missing 'com' pricing")
	}
	if com.Renewal != "10.37" {
		t.Errorf("expected com renewal '10.37', got '%s'", com.Renewal)
	}

	io, ok := loaded["io"]
	if !ok {
		t.Fatal("missing 'io' pricing")
	}
	if io.Renewal != "32.98" {
		t.Errorf("expected io renewal '32.98', got '%s'", io.Renewal)
	}
}

func TestCache_LoadPricing_Empty(t *testing.T) {
	c := newTestCache(t)

	// Load from empty cache should return nil without error
	pricing, updatedAt, err := c.LoadPricing()
	if err != nil {
		t.Fatalf("LoadPricing failed: %v", err)
	}

	if pricing != nil {
		t.Errorf("expected nil pricing, got %v", pricing)
	}
	if !updatedAt.IsZero() {
		t.Errorf("expected zero time, got %v", updatedAt)
	}
}

func TestCache_Clear(t *testing.T) {
	c := newTestCache(t)

	// Save some data
	domains := []api.Domain{{Name: "example.com"}}
	pricing := map[string]api.TLDPricing{"com": {Renewal: "10.37"}}

	if err := c.SaveDomains(domains); err != nil {
		t.Fatalf("SaveDomains failed: %v", err)
	}
	if err := c.SavePricing(pricing); err != nil {
		t.Fatalf("SavePricing failed: %v", err)
	}

	// Verify files exist
	if _, err := os.Stat(filepath.Join(c.dir, domainsFile)); os.IsNotExist(err) {
		t.Fatal("domains file should exist")
	}
	if _, err := os.Stat(filepath.Join(c.dir, pricingFile)); os.IsNotExist(err) {
		t.Fatal("pricing file should exist")
	}

	// Clear
	if err := c.Clear(); err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Verify files are gone
	if _, err := os.Stat(filepath.Join(c.dir, domainsFile)); !os.IsNotExist(err) {
		t.Error("domains file should be deleted")
	}
	if _, err := os.Stat(filepath.Join(c.dir, pricingFile)); !os.IsNotExist(err) {
		t.Error("pricing file should be deleted")
	}
}

func TestCache_LoadDomains_InvalidJSON(t *testing.T) {
	c := newTestCache(t)

	// Write invalid JSON
	path := filepath.Join(c.dir, domainsFile)
	if err := os.WriteFile(path, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Load should fail
	_, _, err := c.LoadDomains()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestCache_LoadPricing_InvalidJSON(t *testing.T) {
	c := newTestCache(t)

	// Write invalid JSON
	path := filepath.Join(c.dir, pricingFile)
	if err := os.WriteFile(path, []byte("not valid json"), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	// Load should fail
	_, _, err := c.LoadPricing()
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestNew(t *testing.T) {
	// This tests the real New() function
	c, err := New()
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if c.dir == "" {
		t.Error("cache dir should not be empty")
	}

	// Dir should exist
	if _, err := os.Stat(c.dir); os.IsNotExist(err) {
		t.Errorf("cache dir should exist: %s", c.dir)
	}
}
