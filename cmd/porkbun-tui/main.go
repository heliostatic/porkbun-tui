package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bc/porkbun-tui/internal/api"
	"github.com/bc/porkbun-tui/internal/cache"
	"github.com/bc/porkbun-tui/internal/config"
	"github.com/bc/porkbun-tui/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

func printUsage() {
	fmt.Println("porkbun-tui - Terminal UI for managing Porkbun domains")
	fmt.Println()
	fmt.Println("Usage: porkbun-tui [options]")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  -h, --help      Show this help message")
	fmt.Println("  -v, --version   Show version")
	fmt.Println()
	fmt.Println("Configuration:")
	fmt.Println("  Set your Porkbun API credentials via environment variables:")
	fmt.Println("    export PORKBUN_API_KEY=pk1_xxx")
	fmt.Println("    export PORKBUN_SECRET_KEY=sk1_xxx")
	fmt.Println()
	fmt.Println("  Or create ~/.config/porkbun-tui/config.yaml:")
	fmt.Println("    api_key: pk1_xxx")
	fmt.Println("    secret_key: sk1_xxx")
	fmt.Println()
	fmt.Println("  Get your API keys at: https://porkbun.com/account/api")
	fmt.Println()
	fmt.Println("Keyboard shortcuts:")
	fmt.Println("  j/k, ↑/↓     Navigate lists")
	fmt.Println("  Enter        Select / expand")
	fmt.Println("  Esc          Go back")
	fmt.Println("  /            Search/filter domains")
	fmt.Println("  d            View DNS records")
	fmt.Println("  n            View/edit nameservers")
	fmt.Println("  t            TLD breakdown (costs)")
	fmt.Println("  c            Calendar view (expirations)")
	fmt.Println("  a            Check domain availability")
	fmt.Println("  r            Refresh data")
	fmt.Println("  ?            Show help")
	fmt.Println("  q            Quit")
	fmt.Println()
	fmt.Println("Cache: ~/.cache/porkbun-tui/")
}

func main() {
	// Parse flags
	showHelp := flag.Bool("help", false, "Show help")
	showVersion := flag.Bool("version", false, "Show version")
	flag.BoolVar(showHelp, "h", false, "Show help")
	flag.BoolVar(showVersion, "v", false, "Show version")
	flag.Usage = printUsage
	flag.Parse()

	if *showHelp {
		printUsage()
		os.Exit(0)
	}

	if *showVersion {
		fmt.Printf("porkbun-tui %s\n", version)
		os.Exit(0)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		fmt.Fprintln(os.Stderr, "\nSet your Porkbun API credentials:")
		fmt.Fprintln(os.Stderr, "  export PORKBUN_API_KEY=pk1_xxx")
		fmt.Fprintln(os.Stderr, "  export PORKBUN_SECRET_KEY=sk1_xxx")
		fmt.Fprintln(os.Stderr, "\nOr create ~/.config/porkbun-tui/config.yaml:")
		fmt.Fprintln(os.Stderr, "  api_key: pk1_xxx")
		fmt.Fprintln(os.Stderr, "  secret_key: sk1_xxx")
		os.Exit(1)
	}

	// Initialize cache
	appCache, err := cache.New()
	if err != nil {
		// Cache is optional, continue without it
		fmt.Fprintf(os.Stderr, "Warning: could not initialize cache: %v\n", err)
	}

	// Load cached data (errors are ignored - cache is optional)
	var cachedDomains []api.Domain
	var cachedPricing map[string]api.TLDPricing
	if appCache != nil {
		cachedDomains, _, _ = appCache.LoadDomains()
		cachedPricing, _, _ = appCache.LoadPricing()
	}

	// Create API client
	client := api.NewClient(cfg)

	// Create and run app
	app := tui.NewApp(client, appCache, cachedDomains, cachedPricing)
	p := tea.NewProgram(app, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running app: %v\n", err)
		os.Exit(1)
	}
}
