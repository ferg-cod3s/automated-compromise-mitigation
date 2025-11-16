// Package main is the entry point for the ACM CLI/TUI client.
//
// The ACM CLI provides both a command-line interface and a terminal user interface (TUI)
// for interacting with the ACM service. It connects to the local ACM service daemon via
// gRPC over mTLS.
//
// Key features:
//   - Detect compromised credentials
//   - Rotate credentials interactively or in batch
//   - View audit logs
//   - Manage ACVS compliance settings
//   - Beautiful TUI using Bubbletea framework
package main

import (
	"fmt"
	"log"
	"os"
)

const (
	cliName    = "acm"
	cliVersion = "0.1.0-dev"
)

func main() {
	// TODO: Parse command-line arguments and subcommands
	// TODO: Implement commands:
	//   - acm status           (show service status)
	//   - acm detect           (detect compromised credentials)
	//   - acm rotate <id>      (rotate specific credential)
	//   - acm rotate --all     (rotate all compromised)
	//   - acm audit --since 7d (view audit log)
	//   - acm compliance ...   (ACVS commands)
	//   - acm config ...       (configuration commands)
	//   - acm cert renew       (renew client certificate)

	// TODO: Load client configuration from ~/.acm/config/client.yaml
	// TODO: Initialize gRPC client with mTLS
	// TODO: Load client certificate from OS keychain

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "status":
		// TODO: Implement status command
		log.Println("Status command - not yet implemented")
	case "detect":
		// TODO: Implement detect command with TUI
		log.Println("Detect command - not yet implemented")
	case "rotate":
		// TODO: Implement rotate command
		log.Println("Rotate command - not yet implemented")
	case "audit":
		// TODO: Implement audit log viewer
		log.Println("Audit command - not yet implemented")
	case "compliance":
		// TODO: Implement ACVS compliance commands
		log.Println("Compliance command - not yet implemented")
	case "config":
		// TODO: Implement configuration commands
		log.Println("Config command - not yet implemented")
	case "version":
		fmt.Printf("%s version %s\n", cliName, cliVersion)
	case "help", "--help", "-h":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

// printUsage displays the CLI usage information
func printUsage() {
	fmt.Printf(`%s - Automated Compromise Mitigation CLI

Usage:
  %s <command> [arguments]

Core Commands:
  status                       Show service status and configuration
  detect                       Detect compromised credentials
  rotate <item-id>             Rotate specific credential
  rotate --all                 Rotate all compromised credentials
  audit --since <duration>     View audit log (e.g., 7d, 24h)

ACVS Commands (requires opt-in):
  compliance enable            Enable ACVS (accepts EULA)
  compliance analyze <url>     Analyze ToS and generate CRC
  compliance status            Show ACVS configuration

Configuration Commands:
  config show                  Display current configuration
  config set <key> <value>     Update configuration
  cert renew                   Renew client certificate

Other Commands:
  version                      Show version information
  help                         Show this help message

For more information, visit: https://github.com/ferg-cod3s/automated-compromise-mitigation
`, cliName, cliName)
}
