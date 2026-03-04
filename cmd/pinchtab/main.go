package main

import (
	"fmt"
	"os"

	"github.com/pinchtab/pinchtab/internal/config"
)

var version = "dev"

func main() {
	cfg := config.Load()

	// Handle version flag
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("pinchtab %s\n", version)
		os.Exit(0)
	}

	// Handle help
	if len(os.Args) > 1 && (os.Args[1] == "help" || os.Args[1] == "--help" || os.Args[1] == "-h") {
		printHelp()
		os.Exit(0)
	}

	// Handle config command (expanded in Phase 2)
	if len(os.Args) > 1 && os.Args[1] == "config" {
		config.HandleConfigCommand(cfg)
		os.Exit(0)
	}

	// Handle connect command
	if len(os.Args) > 1 && os.Args[1] == "connect" {
		handleConnectCommand(cfg)
		os.Exit(0)
	}

	// Handle management CLI commands (health, profiles, instances, tabs)
	if len(os.Args) > 1 && isCLICommand(os.Args[1]) {
		runCLI(cfg)
		os.Exit(0)
	}

	// Check if running as bridge-only instance (spawned by orchestrator)
	if os.Getenv("BRIDGE_ONLY") == "1" {
		runBridgeServer(cfg)
		return
	}

	// Default: run dashboard mode
	// (includes 'pinchtab' with no args and unrecognized args like 'dashboard')
	runDashboard(cfg)
}
