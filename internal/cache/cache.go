package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/bc/porkbun-tui/internal/api"
)

const (
	domainsFile = "domains.json"
	pricingFile = "pricing.json"
)

type Cache struct {
	dir string
}

type CachedDomains struct {
	Data      []api.Domain `json:"data"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type CachedPricing struct {
	Data      map[string]api.TLDPricing `json:"data"`
	UpdatedAt time.Time                 `json:"updated_at"`
}

// New creates a new cache instance using ~/.cache/porkbun-tui/
func New() (*Cache, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	cacheDir := filepath.Join(homeDir, ".cache", "porkbun-tui")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	return &Cache{dir: cacheDir}, nil
}

// LoadDomains loads cached domains from disk
func (c *Cache) LoadDomains() ([]api.Domain, time.Time, error) {
	path := filepath.Join(c.dir, domainsFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, time.Time{}, nil // No cache, not an error
		}
		return nil, time.Time{}, err
	}

	var cached CachedDomains
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, time.Time{}, err
	}

	return cached.Data, cached.UpdatedAt, nil
}

// SaveDomains saves domains to the cache
func (c *Cache) SaveDomains(domains []api.Domain) error {
	cached := CachedDomains{
		Data:      domains,
		UpdatedAt: time.Now(),
	}

	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(c.dir, domainsFile)
	return os.WriteFile(path, data, 0644)
}

// LoadPricing loads cached pricing from disk
func (c *Cache) LoadPricing() (map[string]api.TLDPricing, time.Time, error) {
	path := filepath.Join(c.dir, pricingFile)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, time.Time{}, nil // No cache, not an error
		}
		return nil, time.Time{}, err
	}

	var cached CachedPricing
	if err := json.Unmarshal(data, &cached); err != nil {
		return nil, time.Time{}, err
	}

	return cached.Data, cached.UpdatedAt, nil
}

// SavePricing saves pricing to the cache
func (c *Cache) SavePricing(pricing map[string]api.TLDPricing) error {
	cached := CachedPricing{
		Data:      pricing,
		UpdatedAt: time.Now(),
	}

	data, err := json.MarshalIndent(cached, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(c.dir, pricingFile)
	return os.WriteFile(path, data, 0644)
}

// Clear removes all cached data
func (c *Cache) Clear() error {
	files := []string{domainsFile, pricingFile}
	for _, f := range files {
		path := filepath.Join(c.dir, f)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}
