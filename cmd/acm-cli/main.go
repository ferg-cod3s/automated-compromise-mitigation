// Package main is the entry point for the ACM CLI client.
//
// The ACM CLI provides a command-line interface for interacting with the ACM service.
// It connects to the local ACM service daemon via gRPC over mTLS.
//
// Key features:
//   - Detect compromised credentials
//   - Rotate credentials interactively or in batch
//   - View service health and status
//   - Manage credentials
package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	cliName    = "acm"
	cliVersion = "0.1.0-dev"
	serviceAddr = "127.0.0.1:8443"
	defaultTimeout = 30 * time.Second
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "health":
		runHealth()
	case "detect":
		runDetect()
	case "rotate":
		runRotate()
	case "list":
		runList()
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

// createClient creates a gRPC client with mTLS
func createClient() (*grpc.ClientConn, error) {
	// Get certificate directory
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}
	certDir := filepath.Join(home, ".acm", "certs")

	// Load client certificate
	clientCert, err := tls.LoadX509KeyPair(
		filepath.Join(certDir, "client-cert.pem"),
		filepath.Join(certDir, "client-key.pem"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w\nHave you started the ACM service first?", err)
	}

	// Load CA certificate
	caCert, err := os.ReadFile(filepath.Join(certDir, "ca-cert.pem"))
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add CA certificate to pool")
	}

	// Create TLS config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS13,
		ServerName:   "localhost",
	}

	// Create gRPC connection
	creds := credentials.NewTLS(tlsConfig)
	conn, err := grpc.Dial(serviceAddr, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to service at %s: %w\nIs the ACM service running?", serviceAddr, err)
	}

	return conn, nil
}

// runHealth checks the health of the ACM service
func runHealth() {
	conn, err := createClient()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := acmv1.NewHealthServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	resp, err := client.Check(ctx, &acmv1.HealthRequest{})
	if err != nil {
		log.Fatalf("Health check failed: %v", err)
	}

	fmt.Println("ACM Service Health Check")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Status: %s\n", resp.Status)
	fmt.Printf("Message: %s\n", resp.Message)
	fmt.Printf("Timestamp: %s\n", time.Unix(resp.Timestamp, 0).Format(time.RFC3339))
	fmt.Println("\nComponents:")
	for name, healthy := range resp.Components {
		status := "✓"
		if !healthy {
			status = "✗"
		}
		fmt.Printf("  %s %s\n", status, name)
	}
}

// runDetect detects compromised credentials
func runDetect() {
	conn, err := createClient()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := acmv1.NewCredentialServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	fmt.Println("Scanning for compromised credentials...")
	fmt.Println("(This may take a moment depending on vault size)")
	fmt.Println()

	resp, err := client.DetectCompromised(ctx, &acmv1.DetectRequest{})
	if err != nil {
		log.Fatalf("Detect failed: %v", err)
	}

	if resp.Status.Code != acmv1.StatusCode_STATUS_CODE_SUCCESS {
		log.Fatalf("Detection failed: %s", resp.Status.Message)
	}

	fmt.Printf("Detection Results: %s\n", resp.Status.Message)
	fmt.Println(strings.Repeat("=", 70))

	if len(resp.Credentials) == 0 {
		fmt.Println("✓ No compromised credentials found!")
		return
	}

	fmt.Printf("Found %d compromised credential(s):\n\n", resp.TotalCount)

	for i, cred := range resp.Credentials {
		fmt.Printf("%d. Site: %s\n", i+1, cred.Site)
		fmt.Printf("   Username: %s\n", cred.Username)
		fmt.Printf("   Breach: %s\n", cred.BreachName)
		fmt.Printf("   Date: %s\n", time.Unix(cred.BreachDate, 0).Format("2006-01-02"))
		fmt.Printf("   Severity: %s\n", cred.Severity)
		fmt.Printf("   ID Hash: %s\n", cred.IdHash)
		fmt.Println()
	}

	fmt.Println("To rotate a credential, use: acm rotate <id-hash>")
}

// runRotate rotates a specific credential
func runRotate() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s rotate <credential-id-hash>\n", cliName)
		os.Exit(1)
	}

	credentialID := os.Args[2]

	conn, err := createClient()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := acmv1.NewCredentialServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	// Default password policy
	policy := &acmv1.PasswordPolicy{
		Length:           16,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumbers:   true,
		RequireSymbols:   true,
	}

	fmt.Printf("Rotating credential: %s\n", credentialID)
	fmt.Println("Using password policy:")
	fmt.Printf("  Length: %d\n", policy.Length)
	fmt.Printf("  Uppercase: %v\n", policy.RequireUppercase)
	fmt.Printf("  Lowercase: %v\n", policy.RequireLowercase)
	fmt.Printf("  Numbers: %v\n", policy.RequireNumbers)
	fmt.Printf("  Symbols: %v\n", policy.RequireSymbols)
	fmt.Println()

	resp, err := client.RotateCredential(ctx, &acmv1.RotateRequest{
		CredentialIdHash: credentialID,
		Policy:           policy,
	})
	if err != nil {
		log.Fatalf("Rotation failed: %v", err)
	}

	if resp.Status.Code == acmv1.StatusCode_STATUS_CODE_HIM_REQUIRED {
		fmt.Println("⚠ Human intervention required!")
		fmt.Printf("Reason: %s\n", resp.Status.Message)
		fmt.Println("\nPlease:")
		fmt.Println("  1. Unlock your password manager vault")
		fmt.Println("  2. Try the rotation command again")
		os.Exit(1)
	}

	if resp.Status.Code != acmv1.StatusCode_STATUS_CODE_SUCCESS {
		log.Fatalf("Rotation failed: %s", resp.Status.Message)
	}

	fmt.Println("✓ Credential rotated successfully!")
	fmt.Printf("Status: %s\n", resp.Status.Message)
	if resp.NewPassword != "" {
		fmt.Println("\nNew password has been updated in your vault.")
		fmt.Println("⚠ IMPORTANT: The password manager will sync this change.")
	}
}

// runList lists all credentials
func runList() {
	conn, err := createClient()
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	client := acmv1.NewCredentialServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	resp, err := client.ListCredentials(ctx, &acmv1.ListRequest{})
	if err != nil {
		log.Fatalf("List failed: %v", err)
	}

	if resp.Status.Code != acmv1.StatusCode_STATUS_CODE_SUCCESS {
		log.Fatalf("List failed: %s", resp.Status.Message)
	}

	fmt.Println("Credential List")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Status: %s\n", resp.Status.Message)

	if len(resp.Credentials) == 0 {
		fmt.Println("\nNo credentials found (list functionality not fully implemented in Phase I)")
		return
	}

	fmt.Printf("\nFound %d credential(s):\n\n", len(resp.Credentials))
	for i, cred := range resp.Credentials {
		fmt.Printf("%d. %s\n", i+1, cred.IdHash)
	}
}

// printUsage displays the CLI usage information
func printUsage() {
	fmt.Printf(`%s - Automated Compromise Mitigation CLI

Usage:
  %s <command> [arguments]

Core Commands:
  health                       Check ACM service health
  detect                       Detect compromised credentials
  rotate <id-hash>             Rotate specific credential
  list                         List all credentials (Phase I: limited)

Other Commands:
  version                      Show version information
  help                         Show this help message

Examples:
  %s health                    Check if the ACM service is running
  %s detect                    Scan for compromised credentials
  %s rotate abc123...          Rotate credential with ID hash abc123...

Prerequisites:
  - ACM service must be running (acm-service)
  - Password manager must be unlocked (Bitwarden or 1Password)
  - Certificates must be generated (~/.acm/certs)

For more information, visit: https://github.com/ferg-cod3s/automated-compromise-mitigation
`, cliName, cliName, cliName, cliName, cliName)
}
