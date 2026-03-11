// Command https-vpn is a lightweight VPN using HTTP/2 CONNECT over TLS.
package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nativemind/https-vpn/core"
	"github.com/nativemind/https-vpn/infra/conf"
)

const version = "0.1.0-dev"

func main() {
	// Define commands
	runCmd := flag.NewFlagSet("run", flag.ExitOnError)
	runConfig := runCmd.String("c", "config.json", "Path to config file")

	initCmd := flag.NewFlagSet("init", flag.ExitOnError)
	initCrypto := initCmd.String("crypto", "us", "Crypto provider (us, ru, cn)")

	versionCmd := flag.NewFlagSet("version", flag.ExitOnError)

	// Parse top-level command
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		runCmd.Parse(os.Args[2:])
		runServer(*runConfig)
	case "init":
		initCmd.Parse(os.Args[2:])
		initConfig(*initCrypto)
	case "version":
		versionCmd.Parse(os.Args[2:])
		fmt.Printf("https-vpn %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`https-vpn - Lightweight VPN over HTTP/2

Usage:
  https-vpn <command> [options]

Commands:
  run       Run HTTPS VPN server
  init      Initialize configuration file
  version   Show version
  help      Show this help message

Examples:
  https-vpn init -crypto us
  https-vpn run -c config.json
  https-vpn version`)
}

func runServer(configPath string) {
	// Load config
	cfg, err := conf.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Create instance
	instance, err := core.New(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create instance: %v\n", err)
		os.Exit(1)
	}

	// Start instance
	if err := instance.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("HTTPS VPN server started\n")

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
	if err := instance.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Shutdown error: %v\n", err)
	}
}

func initConfig(cryptoProvider string) {
	cfg := conf.DefaultConfig()
	cfg.Inbounds[0].StreamSettings.TLSSettings.CryptoProvider = cryptoProvider

	// Check if config already exists
	if _, err := os.Stat("config.json"); err == nil {
		fmt.Fprintf(os.Stderr, "config.json already exists\n")
		os.Exit(1)
	}

	if err := conf.SaveConfig("config.json", cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Created config.json")
	fmt.Println("Edit the config file with your certificate paths, then run:")
	fmt.Println("  https-vpn run -c config.json")
}
