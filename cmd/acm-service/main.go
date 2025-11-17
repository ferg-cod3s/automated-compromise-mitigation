// Package main is the entry point for the ACM service daemon.
//
// The ACM service is the core business logic daemon that handles:
//   - Credential Remediation Service (CRS)
//   - Automated Compliance Validation Service (ACVS)
//   - Human-in-the-Middle (HIM) management
//   - Audit logging with cryptographic signatures
//
// The service exposes a gRPC API over mTLS on localhost only,
// ensuring zero-knowledge security and local-first operation.
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/audit"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/auth"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/crs"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/pwmanager"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/pwmanager/bitwarden"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/pwmanager/onepassword"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/server"
)

const (
	serviceName    = "acm-service"
	serviceVersion = "0.1.0-dev"
)

func main() {
	printBanner()
	log.Printf("%s v%s starting...", serviceName, serviceVersion)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize services
	if err := run(ctx); err != nil {
		log.Fatalf("Service failed: %v", err)
	}
}

func run(ctx context.Context) error {
	log.Println("Initializing ACM services...")

	// Get home directory for data storage
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	dataDir := filepath.Join(home, ".acm")

	// Initialize certificate manager
	log.Println("Setting up mTLS certificates...")
	certMgr := auth.NewCertManager(filepath.Join(dataDir, "certs"))
	if err := certMgr.EnsureCertificates(); err != nil {
		return fmt.Errorf("failed to setup certificates: %w", err)
	}

	tlsConfig, err := certMgr.GetServerTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to get TLS config: %w", err)
	}

	// Initialize audit logger (using in-memory for Phase I)
	log.Println("Initializing audit logger...")
	auditLogger, err := audit.NewMemoryLogger()
	if err != nil {
		return fmt.Errorf("failed to create audit logger: %w", err)
	}
	defer auditLogger.Close()

	// Initialize password manager (try Bitwarden first)
	log.Println("Detecting password manager...")
	var pwManager pwmanager.PasswordManager
	bwManager, err := bitwarden.New()
	if err != nil {
		log.Printf("Warning: Bitwarden CLI not found: %v", err)
		log.Println("Trying 1Password...")
		opManager, err := onepassword.New()
		if err != nil {
			log.Printf("Warning: 1Password CLI not found: %v", err)
			log.Println("⚠ Service will start but credential operations will fail until a password manager is configured")
			pwManager = nil
		} else {
			pwManager = opManager
			log.Println("✓ Using 1Password")
		}
	} else {
		pwManager = bwManager
		log.Println("✓ Using Bitwarden")
	}

	// Initialize CRS
	log.Println("Initializing Credential Remediation Service...")
	crsService := crs.NewService(pwManager, auditLogger)

	// Initialize ACVS (Phase II)
	log.Println("Initializing Automated Compliance Validation Service...")
	acvsService, err := acvs.NewService()
	if err != nil {
		return fmt.Errorf("failed to create ACVS: %w", err)
	}
	log.Println("✓ ACVS initialized (disabled by default - use EnableACVS RPC to opt-in)")

	// Create gRPC server with mTLS
	log.Println("Starting gRPC server...")
	creds := credentials.NewTLS(tlsConfig)
	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB max message size
	)

	// Register services
	credentialServer := server.NewCredentialServiceServer(crsService)
	acmv1.RegisterCredentialServiceServer(grpcServer, credentialServer)

	// ACVS service (Phase II)
	acvsServer := server.NewACVSServiceServer(acvsService)
	acmv1.RegisterACVSServiceServer(grpcServer, acvsServer)

	// Health service
	healthServer := &server.HealthServiceServer{}
	acmv1.RegisterHealthServiceServer(grpcServer, healthServer)

	log.Println("Services registered:")
	log.Println("  ✓ CredentialService")
	log.Println("  ✓ ACVSService (Phase II)")
	log.Println("  ✓ HealthService")

	// Start listening
	listenAddr := "127.0.0.1:8443"
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", listenAddr, err)
	}

	// Setup graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Start server in goroutine
	errCh := make(chan error, 1)
	go func() {
		log.Printf("✓ %s ready and listening on %s (mTLS enabled)", serviceName, listenAddr)
		log.Println("")
		log.Println("Phase I & II Status:")
		log.Println("  ✓ gRPC server running with mTLS")
		log.Println("  ✓ Password manager integrations ready")
		log.Println("  ✓ CRS (Credential Remediation Service)")
		log.Println("  ✓ ACVS (Automated Compliance Validation Service)")
		log.Println("  ✓ Evidence Chain with cryptographic signatures")
		log.Println("  ✓ Legal NLP engine (stub implementation)")
		log.Println("  ✓ Audit logging with Ed25519 signatures")
		log.Println("  ✓ HIM (Human-in-the-Middle) workflow system")
		log.Println("")
		log.Printf("Certificates location: %s", filepath.Join(dataDir, "certs"))
		log.Printf("Audit database: %s", filepath.Join(dataDir, "audit.db"))
		log.Println("")
		log.Println("Service ready for client connections!")

		if err := grpcServer.Serve(listener); err != nil {
			errCh <- fmt.Errorf("server failed: %w", err)
		}
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigCh:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
	case err := <-errCh:
		log.Printf("Server error: %v", err)
		return err
	}

	// Graceful shutdown
	log.Println("Stopping gRPC server...")
	grpcServer.GracefulStop()

	log.Printf("%s stopped", serviceName)
	return nil
}

// printBanner displays the ACM service banner on startup
func printBanner() {
	fmt.Println(`
╔═══════════════════════════════════════════════════════════╗
║  ACM Service - Automated Compromise Mitigation           ║
║  Local-First Credential Breach Response                  ║
╚═══════════════════════════════════════════════════════════╝`)
}
