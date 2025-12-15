package demo

import (
	"time"

	"github.com/bc/porkbun-tui/internal/api"
)

// Domains returns sample domain data for demo mode
func Domains() []api.Domain {
	now := time.Now()
	return []api.Domain{
		{Name: "acmecorp.com", TLD: "com", CreateDate: now.AddDate(-5, 0, 0), ExpireDate: now.AddDate(0, 3, 0), AutoRenew: true, Status: "ACTIVE", SecurityLock: true, WhoisPrivacy: true},
		{Name: "startupkit.io", TLD: "io", CreateDate: now.AddDate(-2, 0, 0), ExpireDate: now.AddDate(0, 6, 0), AutoRenew: true, Status: "ACTIVE", SecurityLock: true, WhoisPrivacy: true},
		{Name: "devtools.dev", TLD: "dev", CreateDate: now.AddDate(-3, 0, 0), ExpireDate: now.AddDate(0, 1, 5), AutoRenew: false, Status: "ACTIVE", SecurityLock: false, WhoisPrivacy: true},
		{Name: "cloudnative.app", TLD: "app", CreateDate: now.AddDate(-1, -3, 0), ExpireDate: now.AddDate(0, 9, 0), AutoRenew: true, Status: "ACTIVE", SecurityLock: true, WhoisPrivacy: true},
		{Name: "myblog.org", TLD: "org", CreateDate: now.AddDate(-6, 0, 0), ExpireDate: now.AddDate(0, 11, 0), AutoRenew: true, Status: "ACTIVE", SecurityLock: true, WhoisPrivacy: false},
		{Name: "shopfront.store", TLD: "store", CreateDate: now.AddDate(0, -10, 0), ExpireDate: now.AddDate(1, 2, 0), AutoRenew: true, Status: "ACTIVE", SecurityLock: true, WhoisPrivacy: true},
		{Name: "portfolio.design", TLD: "design", CreateDate: now.AddDate(-1, -8, 0), ExpireDate: now.AddDate(0, 4, 7), AutoRenew: false, Status: "ACTIVE", SecurityLock: false, WhoisPrivacy: true},
		{Name: "techbytes.net", TLD: "net", CreateDate: now.AddDate(-4, 0, 0), ExpireDate: now.AddDate(0, 8, 15), AutoRenew: true, Status: "ACTIVE", SecurityLock: true, WhoisPrivacy: true},
		{Name: "gameserver.gg", TLD: "gg", CreateDate: now.AddDate(0, -11, 0), ExpireDate: now.AddDate(1, 0, 15), AutoRenew: true, Status: "ACTIVE", SecurityLock: true, WhoisPrivacy: true},
		{Name: "cryptotrader.xyz", TLD: "xyz", CreateDate: now.AddDate(-3, 0, 0), ExpireDate: now.AddDate(0, 0, 15), AutoRenew: false, Status: "ACTIVE", SecurityLock: false, WhoisPrivacy: true},
		{Name: "mailservice.email", TLD: "email", CreateDate: now.AddDate(-1, -5, 0), ExpireDate: now.AddDate(0, 7, 0), AutoRenew: true, Status: "ACTIVE", SecurityLock: true, WhoisPrivacy: true},
		{Name: "aistartup.ai", TLD: "ai", CreateDate: now.AddDate(0, -9, 0), ExpireDate: now.AddDate(0, 2, 15), AutoRenew: true, Status: "ACTIVE", SecurityLock: true, WhoisPrivacy: true},
	}
}

// Pricing returns sample TLD pricing data for demo mode
func Pricing() map[string]api.TLDPricing {
	return map[string]api.TLDPricing{
		"com":    {TLD: "com", Registration: "9.73", Renewal: "10.37", Transfer: "10.37"},
		"io":     {TLD: "io", Registration: "32.98", Renewal: "32.98", Transfer: "32.98"},
		"dev":    {TLD: "dev", Registration: "14.00", Renewal: "14.00", Transfer: "14.00"},
		"app":    {TLD: "app", Registration: "14.00", Renewal: "14.00", Transfer: "14.00"},
		"org":    {TLD: "org", Registration: "10.87", Renewal: "10.87", Transfer: "10.87"},
		"store":  {TLD: "store", Registration: "5.00", Renewal: "25.00", Transfer: "25.00"},
		"design": {TLD: "design", Registration: "25.00", Renewal: "25.00", Transfer: "25.00"},
		"net":    {TLD: "net", Registration: "11.52", Renewal: "11.52", Transfer: "11.52"},
		"gg":     {TLD: "gg", Registration: "75.00", Renewal: "75.00", Transfer: "75.00"},
		"xyz":    {TLD: "xyz", Registration: "2.00", Renewal: "12.00", Transfer: "12.00"},
		"email":  {TLD: "email", Registration: "20.00", Renewal: "20.00", Transfer: "20.00"},
		"ai":     {TLD: "ai", Registration: "50.00", Renewal: "50.00", Transfer: "50.00"},
	}
}
