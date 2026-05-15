package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-trader/internal/config"
	"github.com/go-trader/internal/exchange"
	"github.com/go-trader/internal/strategy"
)

const (
	version = "0.1.0"
)

func main() {
	// Parse command-line flags
	configFile := flag.String("config", "config.yaml", "Path to configuration file")
	showVersion := flag.Bool("version", false, "Print version and exit")
	// Default dry-run to true so I don't accidentally place real orders while experimenting
	dryRun := flag.Bool("dry-run", true, "Run in simulation mode without placing real orders")
	flag.Parse()

	if *showVersion {
		fmt.Printf("go-trader v%s\n", version)
		os.Exit(0)
	}

	log.Printf("Starting go-trader v%s", version)

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", *configFile, err)
	}

	if *dryRun {
		log.Println("Running in dry-run mode — no real orders will be placed")
		cfg.DryRun = true
	} else {
		// Extra warning so I don't forget I'm trading with real money
		log.Println("WARNING: dry-run is disabled — REAL orders will be placed!")
	}

	// Initialize exchange client
	client, err := exchange.NewClient(cfg.Exchange)
	if err != nil {
		log.Fatalf("Failed to initialize exchange client: %v", err)
	}
	defer client.Close()

	// Initialize and start strategy engine
	engine, err := strategy.NewEngine(cfg, client)
	if err != nil {
		log.Fatalf("Failed to initialize strategy engine: %v", err)
	}

	if err := engine.Start(); err != nil {
		log.Fatalf("Failed to start strategy engine: %v", err)
	}
	defer engine.Stop()

	log.Printf("go-trader running. Exchange: %s, Strategy: %s", cfg.Exchange.Name, cfg.Strategy.Name)

	// Wait for termination signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Printf("Received signal %s, shutting down...", sig)
}
