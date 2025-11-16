// Package auth implements authentication and authorization for the ACM service.
//
// The auth package handles:
//   - mTLS client certificate authentication
//   - JWT session token management
//   - Certificate generation and renewal
//   - OS keychain integration for certificate storage
//
// # Authentication Flow
//
// 1. Client presents X.509 client certificate during mTLS handshake
// 2. Service validates certificate against local CA
// 3. Service issues short-lived JWT token (15-30 min)
// 4. Client includes JWT in gRPC metadata for subsequent requests
// 5. Service validates JWT signature and expiration
// 6. Client proactively refreshes token before expiration
//
// # Certificate Hierarchy
//
//	ACM Local CA (Self-Signed Root)
//	├── ACM Service Certificate (server cert)
//	│   └── CN: localhost
//	├── OpenTUI Client Certificate (client cert)
//	│   └── CN: acm-tui-<device-id>
//	└── Tauri GUI Client Certificate (client cert)
//	    └── CN: acm-gui-<device-id>
//
// # TLS Configuration
//
//   - TLS 1.3 required (no fallback to older versions)
//   - Strong cipher suites only: AES-256-GCM, ChaCha20-Poly1305
//   - Client authentication required (mutual TLS)
//   - Certificate pinning for localhost
//   - 1-year certificate lifetime with renewal support
//
// # JWT Token Structure
//
//	{
//	  "alg": "EdDSA",
//	  "typ": "JWT"
//	}
//	{
//	  "iss": "acm-service",
//	  "sub": "acm-tui-<device-id>",
//	  "aud": "acm-api",
//	  "exp": 1700000000,
//	  "iat": 1699998200,
//	  "jti": "unique-token-id",
//	  "cert_fingerprint": "sha256:abcd1234..."
//	}
//
// # OS Keychain Integration
//
// Private keys are stored in platform-specific secure storage:
//
//   - macOS: Keychain Services API
//   - Windows: Windows Certificate Store
//   - Linux: Secret Service API (libsecret) or GPG keyring
//
// # Example Usage
//
//	// Server-side: Create mTLS server
//	server, err := auth.NewMTLSServer(
//	    certFile: "/home/user/.acm/certs/server.pem",
//	    keyFile:  "/home/user/.acm/certs/server-key.pem",
//	    caFile:   "/home/user/.acm/certs/ca.pem",
//	    addr:     "127.0.0.1:8443",
//	)
//
//	// Client-side: Create mTLS client
//	conn, err := auth.NewMTLSClient(
//	    certFile: "/home/user/.acm/certs/client-tui.pem",
//	    keyFile:  "/home/user/.acm/certs/client-tui-key.pem",
//	    caFile:   "/home/user/.acm/certs/ca.pem",
//	    addr:     "127.0.0.1:8443",
//	)
//
//	// Issue JWT token after mTLS handshake
//	token, err := auth.IssueToken(clientCert, signingKey, 30*time.Minute)
//
//	// Validate JWT token in request
//	claims, err := auth.ValidateToken(token, publicKey)
//
// # Phase I Implementation
//
// Phase I focuses on:
//   - Self-signed CA generation
//   - X.509 client certificate generation
//   - mTLS server and client configuration
//   - Ed25519 JWT token issuance and validation
//   - Basic OS keychain integration (macOS, Linux)
//
// Future phases will add:
//   - Hardware security module (HSM) support
//   - TPM integration for key storage
//   - Certificate revocation lists (CRL)
//   - OCSP stapling for certificate validation
//   - Automated certificate rotation
package auth
