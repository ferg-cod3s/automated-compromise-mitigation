// Package auth provides mTLS certificate management for ACM.
//
// For Phase I, we use a simplified approach with self-signed certificates.
// Phase II will add full PKI with certificate rotation.
package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"
)

// CertManager handles certificate generation and loading for mTLS.
type CertManager struct {
	certDir string
}

// NewCertManager creates a new certificate manager.
func NewCertManager(certDir string) *CertManager {
	if certDir == "" {
		// Default to ~/.acm/certs
		home, _ := os.UserHomeDir()
		certDir = filepath.Join(home, ".acm", "certs")
	}
	return &CertManager{
		certDir: certDir,
	}
}

// EnsureCertificates generates certificates if they don't exist.
func (cm *CertManager) EnsureCertificates() error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(cm.certDir, 0700); err != nil {
		return fmt.Errorf("failed to create cert directory: %w", err)
	}

	serverCertPath := filepath.Join(cm.certDir, "server-cert.pem")
	serverKeyPath := filepath.Join(cm.certDir, "server-key.pem")
	clientCertPath := filepath.Join(cm.certDir, "client-cert.pem")
	clientKeyPath := filepath.Join(cm.certDir, "client-key.pem")
	caPath := filepath.Join(cm.certDir, "ca-cert.pem")

	// Check if certificates exist
	if fileExists(serverCertPath) && fileExists(serverKeyPath) &&
		fileExists(clientCertPath) && fileExists(clientKeyPath) && fileExists(caPath) {
		return nil // Certificates already exist
	}

	// Generate CA certificate
	caCert, caKey, err := generateCA()
	if err != nil {
		return fmt.Errorf("failed to generate CA: %w", err)
	}

	// Save CA certificate
	if err := saveCertificate(caPath, caCert); err != nil {
		return err
	}

	// Generate server certificate
	serverCert, serverKey, err := generateCertificate(caCert, caKey, "localhost", true)
	if err != nil {
		return fmt.Errorf("failed to generate server certificate: %w", err)
	}

	if err := saveCertificate(serverCertPath, serverCert); err != nil {
		return err
	}
	if err := savePrivateKey(serverKeyPath, serverKey); err != nil {
		return err
	}

	// Generate client certificate
	clientCert, clientKey, err := generateCertificate(caCert, caKey, "acm-client", false)
	if err != nil {
		return fmt.Errorf("failed to generate client certificate: %w", err)
	}

	if err := saveCertificate(clientCertPath, clientCert); err != nil {
		return err
	}
	if err := savePrivateKey(clientKeyPath, clientKey); err != nil {
		return err
	}

	return nil
}

// GetServerTLSConfig returns the TLS config for the gRPC server.
func (cm *CertManager) GetServerTLSConfig() (*tls.Config, error) {
	serverCert, err := tls.LoadX509KeyPair(
		filepath.Join(cm.certDir, "server-cert.pem"),
		filepath.Join(cm.certDir, "server-key.pem"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate: %w", err)
	}

	caCert, err := os.ReadFile(filepath.Join(cm.certDir, "ca-cert.pem"))
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add CA certificate to pool")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS13,
	}, nil
}

// GetClientTLSConfig returns the TLS config for gRPC clients.
func (cm *CertManager) GetClientTLSConfig() (*tls.Config, error) {
	clientCert, err := tls.LoadX509KeyPair(
		filepath.Join(cm.certDir, "client-cert.pem"),
		filepath.Join(cm.certDir, "client-key.pem"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	caCert, err := os.ReadFile(filepath.Join(cm.certDir, "ca-cert.pem"))
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to add CA certificate to pool")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS13,
		ServerName:   "localhost",
	}, nil
}

// Helper functions

func generateCA() (*x509.Certificate, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"ACM"},
			CommonName:   "ACM Root CA",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
		MaxPathLen:            1,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, nil, err
	}

	return cert, privateKey, nil
}

func generateCertificate(caCert *x509.Certificate, caKey *rsa.PrivateKey, commonName string, isServer bool) (*x509.Certificate, *rsa.PrivateKey, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	serialNumber, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return nil, nil, err
	}

	template := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"ACM"},
			CommonName:   commonName,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().AddDate(1, 0, 0), // 1 year
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	if isServer {
		template.DNSNames = []string{"localhost"}
		template.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, template, caCert, &privateKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	cert, err := x509.ParseCertificate(certBytes)
	if err != nil {
		return nil, nil, err
	}

	return cert, privateKey, nil
}

func saveCertificate(path string, cert *x509.Certificate) error {
	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	})
	return os.WriteFile(path, certPEM, 0600)
}

func savePrivateKey(path string, key *rsa.PrivateKey) error {
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	return os.WriteFile(path, keyPEM, 0600)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
