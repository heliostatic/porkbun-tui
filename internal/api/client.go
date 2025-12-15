package api

import (
	"context"
	"fmt"
	"time"

	"github.com/bc/porkbun-tui/internal/config"
	"github.com/tuzzmaniandevil/porkbun-go"
)

type Client struct {
	pb *porkbun.Client
}

type Domain struct {
	Name         string
	Status       string
	TLD          string
	CreateDate   time.Time
	ExpireDate   time.Time
	SecurityLock bool
	WhoisPrivacy bool
	AutoRenew    bool
	NotLocal     bool
	Labels       []string
}

type DNSRecord struct {
	ID       string
	Name     string
	Type     string
	Content  string
	TTL      string
	Priority string
	Notes    string
}

type AvailabilityResult struct {
	Domain    string
	Available bool
	Price     string
	Premium   bool
}

type TLDPricing struct {
	TLD          string
	Registration string
	Renewal      string
	Transfer     string
}

func NewClient(cfg *config.Config) *Client {
	pb := porkbun.NewClient(&porkbun.Options{
		ApiKey:       cfg.APIKey,
		SecretApiKey: cfg.SecretKey,
	})
	return &Client{pb: pb}
}

func (c *Client) Ping(ctx context.Context) (string, error) {
	resp, err := c.pb.Ping(ctx)
	if err != nil {
		return "", err
	}
	return resp.YourIP, nil
}

func (c *Client) ListDomains(ctx context.Context) ([]Domain, error) {
	resp, err := c.pb.Domains.ListDomains(ctx, nil)
	if err != nil {
		return nil, err
	}

	domains := make([]Domain, 0, len(resp.Domains))
	for _, d := range resp.Domains {
		domain := Domain{
			Name:         d.Domain,
			Status:       d.Status,
			TLD:          d.TLD,
			SecurityLock: bool(d.SecurityLock),
			WhoisPrivacy: bool(d.WhoisPrivacy),
			AutoRenew:    bool(d.AutoRenew),
			NotLocal:     bool(d.NotLocal),
			CreateDate:   d.CreateDate,
			ExpireDate:   d.ExpireDate,
		}

		// Handle labels
		if d.Labels != nil {
			for _, l := range d.Labels {
				domain.Labels = append(domain.Labels, l.Title)
			}
		}

		domains = append(domains, domain)
	}

	return domains, nil
}

func (c *Client) GetDNSRecords(ctx context.Context, domain string) ([]DNSRecord, error) {
	resp, err := c.pb.Dns.GetRecords(ctx, domain, nil)
	if err != nil {
		return nil, err
	}

	records := make([]DNSRecord, 0, len(resp.Records))
	for _, r := range resp.Records {
		id := ""
		if r.ID != nil {
			id = fmt.Sprintf("%d", *r.ID)
		}
		records = append(records, DNSRecord{
			ID:       id,
			Name:     r.Name,
			Type:     string(r.Type),
			Content:  r.Content,
			TTL:      r.TTL,
			Priority: r.Prio,
			Notes:    r.Notes,
		})
	}

	return records, nil
}

func (c *Client) GetNameservers(ctx context.Context, domain string) ([]string, error) {
	resp, err := c.pb.Domains.GetNameServers(ctx, domain)
	if err != nil {
		return nil, err
	}
	return resp.NS, nil
}

func (c *Client) UpdateNameservers(ctx context.Context, domain string, nameservers []string) error {
	ns := porkbun.NameServers(nameservers)
	_, err := c.pb.Domains.UpdateNameServers(ctx, domain, &ns)
	return err
}

func (c *Client) CheckAvailability(ctx context.Context, domain string) (*AvailabilityResult, error) {
	// The porkbun-go SDK doesn't seem to have domain availability check
	// We'll need to make a direct API call or use the pricing endpoint
	// For now, return a placeholder that indicates this feature needs implementation
	return &AvailabilityResult{
		Domain:    domain,
		Available: false,
		Price:     "Feature not yet implemented in SDK",
		Premium:   false,
	}, nil
}

func (c *Client) GetPricing(ctx context.Context) (map[string]TLDPricing, error) {
	resp, err := c.pb.Pricing.ListPricing(ctx)
	if err != nil {
		return nil, err
	}

	pricing := make(map[string]TLDPricing, len(resp.Pricing))
	for tld, p := range resp.Pricing {
		pricing[tld] = TLDPricing{
			TLD:          tld,
			Registration: p.Registration,
			Renewal:      p.Renewal,
			Transfer:     p.Transfer,
		}
	}

	return pricing, nil
}
