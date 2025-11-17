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
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/audit"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/auth"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/crs"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/logging"
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

	// Initialize structured logging first
	config := logging.DefaultConfig()
	// Use pretty format for development, JSON for production
	if os.Getenv("ACM_ENV") == "production" {
		config.Format = logging.FormatJSON
	} else {
		config.Format = logging.FormatPretty
	}

	if err := logging.Initialize(config); err != nil {
		log.Fatalf("Failed to initialize logging: %v", err)
	}

	logger := logging.NewLogger("main")
	logger.Info("ACM service starting",
		"service", serviceName,
		"version", serviceVersion,
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize services
	if err := run(ctx, logger); err != nil {
		logger.Error("Service failed", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *logging.Logger) error {
	logger.Info("Initializing ACM services")

	// Get home directory for data storage
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}
	dataDir := filepath.Join(home, ".acm")

	// Initialize certificate manager
	logger.Info("Setting up mTLS certificates", "cert_dir", filepath.Join(dataDir, "certs"))
	certMgr := auth.NewCertManager(filepath.Join(dataDir, "certs"))
	if err := certMgr.EnsureCertificates(); err != nil {
		return fmt.Errorf("failed to setup certificates: %w", err)
	}

	tlsConfig, err := certMgr.GetServerTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to get TLS config: %w", err)
	}

	// Initialize audit logger (using in-memory for Phase I)
	logger.Info("Initializing audit logger")
	auditLogger, err := audit.NewMemoryLogger()
	if err != nil {
		return fmt.Errorf("failed to create audit logger: %w", err)
	}
	defer auditLogger.Close()

	// Initialize password manager (try Bitwarden first)
	logger.Info("Detecting password manager")
	var pwManager pwmanager.PasswordManager
	bwManager, err := bitwarden.New()
	if err != nil {
		logger.Warn("Bitwarden CLI not found", "error", err)
		logger.Info("Trying 1Password")
		opManager, err := onepassword.New()
		if err != nil {
			logger.Warn("1Password CLI not found", "error", err)
			logger.Warn("Service will start but credential operations will fail until a password manager is configured")
			pwManager = nil
		} else {
			pwManager = opManager
			logger.Info("Password manager detected", "manager", "1Password")
		}
	} else {
		pwManager = bwManager
		logger.Info("Password manager detected", "manager", "Bitwarden")
	}

	// Initialize CRS
	logger.Info("Initializing Credential Remediation Service")
	crsService := crs.NewService(pwManager, auditLogger)

	// Initialize ACVS (Phase II)
	logger.Info("Initializing Automated Compliance Validation Service")
	acvsService, err := acvs.NewService()
	if err != nil {
		return fmt.Errorf("failed to create ACVS: %w", err)
	}
	logger.Info("ACVS initialized", "enabled_by_default", false)

	// Create gRPC server with mTLS and logging middleware
	logger.Info("Starting gRPC server with middleware")
	creds := credentials.NewTLS(tlsConfig)

	// Create server logger for gRPC middleware
	grpcLogger := logging.NewLogger("grpc")

	grpcServer := grpc.NewServer(
		grpc.Creds(creds),
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB max message size
		// Add request ID and logging interceptors
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(grpcLogger),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(grpcLogger),
		),
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

	logger.Info("Services registered",
		"services", []string{"CredentialService", "ACVSService", "HealthService"},
	)

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
		logger.Info("ACM service ready",
			"address", listenAddr,
			"mtls", true,
			"cert_dir", filepath.Join(dataDir, "certs"),
			"audit_db", filepath.Join(dataDir, "audit.db"),
		)

		logger.Info("Phase I & II components active",
			"grpc_server", true,
			"mtls", true,
			"password_managers", pwManager != nil,
			"crs", true,
			"acvs", true,
			"evidence_chain", true,
			"legal_nlp", "stub",
			"audit_logging", true,
			"him_workflows", true,
		)

		logger.Info("Service ready for client connections")

		if err := grpcServer.Serve(listener); err != nil {
			errCh <- fmt.Errorf("server failed: %w", err)
		}
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigCh:
		logger.Info("Shutdown signal received", "signal", sig.String())
	case err := <-errCh:
		logger.Error("Server error", "error", err)
		return err
	}

	// Graceful shutdown
	logger.Info("Stopping gRPC server")
	grpcServer.GracefulStop()

	logger.Info("ACM service stopped", "service", serviceName)

	// Shutdown logging system (flush and close log files)
	if err := logging.Shutdown(); err != nil {
		logger.Error("Failed to shutdown logging", "error", err)
	}

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
