#!/usr/bin/env bash
# Certificate generation script for ACM mTLS
# Generates self-signed CA and certificates for local development

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Certificate directory
CERT_DIR="${ROOT_DIR}/certs"

# Certificate parameters
CA_SUBJECT="/C=US/ST=California/L=San Francisco/O=ACM Development/OU=Security/CN=ACM Development CA"
SERVER_SUBJECT="/C=US/ST=California/L=San Francisco/O=ACM Development/OU=Service/CN=localhost"
CLIENT_SUBJECT="/C=US/ST=California/L=San Francisco/O=ACM Development/OU=Client/CN=acm-cli"

# Certificate validity (in days)
CA_DAYS=3650      # 10 years
CERT_DAYS=365     # 1 year

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

# Check prerequisites
check_prerequisites() {
    if ! command -v openssl &> /dev/null; then
        log_error "OpenSSL is not installed"
        log_info "Install OpenSSL:"
        log_info "  macOS:   brew install openssl"
        log_info "  Ubuntu:  sudo apt-get install openssl"
        log_info "  Fedora:  sudo dnf install openssl"
        exit 1
    fi

    OPENSSL_VERSION=$(openssl version | awk '{print $2}')
    log_info "OpenSSL version: $OPENSSL_VERSION"
}

# Create certificate directory
create_cert_dir() {
    log_info "Creating certificate directory..."

    if [ -d "$CERT_DIR" ]; then
        log_warn "Certificate directory already exists: $CERT_DIR"
        read -p "Do you want to regenerate certificates? (yes/no): " -r
        echo

        if [[ ! $REPLY =~ ^[Yy](es)?$ ]]; then
            log_info "Keeping existing certificates"
            exit 0
        fi

        log_warn "Removing existing certificates..."
        rm -rf "$CERT_DIR"
    fi

    mkdir -p "$CERT_DIR"
    log_success "Created certificate directory: $CERT_DIR"
}

# Generate CA certificate
generate_ca() {
    log_info "Generating Certificate Authority (CA)..."

    # Generate CA private key
    openssl genrsa -out "$CERT_DIR/ca-key.pem" 4096 2>/dev/null

    # Generate CA certificate
    openssl req -new -x509 \
        -days $CA_DAYS \
        -key "$CERT_DIR/ca-key.pem" \
        -out "$CERT_DIR/ca-cert.pem" \
        -subj "$CA_SUBJECT" 2>/dev/null

    log_success "Generated CA certificate (valid for $CA_DAYS days)"
}

# Generate server certificate
generate_server_cert() {
    log_info "Generating server certificate..."

    # Generate server private key
    openssl genrsa -out "$CERT_DIR/server-key.pem" 4096 2>/dev/null

    # Generate server CSR
    openssl req -new \
        -key "$CERT_DIR/server-key.pem" \
        -out "$CERT_DIR/server.csr" \
        -subj "$SERVER_SUBJECT" 2>/dev/null

    # Create server certificate extensions file
    cat > "$CERT_DIR/server-ext.cnf" << EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = 127.0.0.1
DNS.3 = ::1
IP.1 = 127.0.0.1
IP.2 = ::1
EOF

    # Sign server certificate with CA
    openssl x509 -req \
        -in "$CERT_DIR/server.csr" \
        -CA "$CERT_DIR/ca-cert.pem" \
        -CAkey "$CERT_DIR/ca-key.pem" \
        -CAcreateserial \
        -out "$CERT_DIR/server-cert.pem" \
        -days $CERT_DAYS \
        -extfile "$CERT_DIR/server-ext.cnf" 2>/dev/null

    # Clean up CSR and extensions file
    rm -f "$CERT_DIR/server.csr" "$CERT_DIR/server-ext.cnf"

    log_success "Generated server certificate (valid for $CERT_DAYS days)"
}

# Generate client certificate
generate_client_cert() {
    log_info "Generating client certificate..."

    # Generate client private key
    openssl genrsa -out "$CERT_DIR/client-key.pem" 4096 2>/dev/null

    # Generate client CSR
    openssl req -new \
        -key "$CERT_DIR/client-key.pem" \
        -out "$CERT_DIR/client.csr" \
        -subj "$CLIENT_SUBJECT" 2>/dev/null

    # Create client certificate extensions file
    cat > "$CERT_DIR/client-ext.cnf" << EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, keyEncipherment
extendedKeyUsage = clientAuth
EOF

    # Sign client certificate with CA
    openssl x509 -req \
        -in "$CERT_DIR/client.csr" \
        -CA "$CERT_DIR/ca-cert.pem" \
        -CAkey "$CERT_DIR/ca-key.pem" \
        -CAcreateserial \
        -out "$CERT_DIR/client-cert.pem" \
        -days $CERT_DAYS \
        -extfile "$CERT_DIR/client-ext.cnf" 2>/dev/null

    # Clean up CSR and extensions file
    rm -f "$CERT_DIR/client.csr" "$CERT_DIR/client-ext.cnf"

    log_success "Generated client certificate (valid for $CERT_DAYS days)"
}

# Set appropriate permissions
set_permissions() {
    log_info "Setting certificate permissions..."

    # Private keys should be readable only by owner
    chmod 600 "$CERT_DIR"/*-key.pem

    # Certificates can be readable by all
    chmod 644 "$CERT_DIR"/*-cert.pem
    chmod 644 "$CERT_DIR"/ca-cert.pem

    log_success "Set certificate permissions"
}

# Verify certificates
verify_certs() {
    log_info "Verifying certificates..."

    # Verify CA certificate
    if ! openssl x509 -in "$CERT_DIR/ca-cert.pem" -noout -text &>/dev/null; then
        log_error "CA certificate verification failed"
        exit 1
    fi

    # Verify server certificate
    if ! openssl verify -CAfile "$CERT_DIR/ca-cert.pem" "$CERT_DIR/server-cert.pem" &>/dev/null; then
        log_error "Server certificate verification failed"
        exit 1
    fi

    # Verify client certificate
    if ! openssl verify -CAfile "$CERT_DIR/ca-cert.pem" "$CERT_DIR/client-cert.pem" &>/dev/null; then
        log_error "Client certificate verification failed"
        exit 1
    fi

    log_success "All certificates verified successfully"
}

# Display certificate information
display_cert_info() {
    echo ""
    log_info "Certificate Information"
    echo ""

    echo "CA Certificate:"
    openssl x509 -in "$CERT_DIR/ca-cert.pem" -noout -subject -issuer -dates | sed 's/^/  /'
    echo ""

    echo "Server Certificate:"
    openssl x509 -in "$CERT_DIR/server-cert.pem" -noout -subject -issuer -dates | sed 's/^/  /'
    openssl x509 -in "$CERT_DIR/server-cert.pem" -noout -ext subjectAltName | sed 's/^/  /'
    echo ""

    echo "Client Certificate:"
    openssl x509 -in "$CERT_DIR/client-cert.pem" -noout -subject -issuer -dates | sed 's/^/  /'
    echo ""
}

# Create .gitignore in certs directory
create_gitignore() {
    cat > "$CERT_DIR/.gitignore" << 'EOF'
# Ignore all certificate files (security-sensitive)
*.pem
*.csr
*.srl

# Keep this .gitignore
!.gitignore

# Keep README
!README.md
EOF

    log_success "Created .gitignore in certs directory"
}

# Create README in certs directory
create_readme() {
    cat > "$CERT_DIR/README.md" << 'EOF'
# ACM mTLS Certificates

This directory contains self-signed certificates for local development.

## Files

- `ca-cert.pem` - Certificate Authority certificate
- `ca-key.pem` - Certificate Authority private key
- `server-cert.pem` - Server certificate (for acm-service)
- `server-key.pem` - Server private key
- `client-cert.pem` - Client certificate (for acm-cli)
- `client-key.pem` - Client private key

## Security Notice

**DO NOT commit these files to version control.**

These certificates are for local development only and should be regenerated for each development environment.

## Regeneration

To regenerate certificates:

```bash
make cert-gen
```

Or run directly:

```bash
./scripts/generate-certs.sh
```

## Production

For production deployments, use certificates from a trusted Certificate Authority or your organization's PKI infrastructure.
EOF

    log_success "Created README.md in certs directory"
}

# Print summary
print_summary() {
    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║              mTLS Certificates Generated                  ║"
    echo "╠════════════════════════════════════════════════════════════╣"
    echo "║  Location: $CERT_DIR"
    echo "║                                                            ║"
    echo "║  Generated Files:                                          ║"
    echo "║    - ca-cert.pem      (CA certificate)                     ║"
    echo "║    - ca-key.pem       (CA private key)                     ║"
    echo "║    - server-cert.pem  (Server certificate)                 ║"
    echo "║    - server-key.pem   (Server private key)                 ║"
    echo "║    - client-cert.pem  (Client certificate)                 ║"
    echo "║    - client-key.pem   (Client private key)                 ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    log_success "Certificates ready for use!"
    echo ""
    log_warn "Security Notice:"
    echo "  - These are self-signed certificates for development only"
    echo "  - DO NOT use in production"
    echo "  - DO NOT commit to version control"
    echo "  - Regenerate if compromised"
    echo ""
    log_info "Next steps:"
    echo "  1. Update config files to use these certificates"
    echo "  2. Start the service: make dev"
    echo "  3. Connect with CLI: make dev-cli"
    echo ""
}

# Main execution
main() {
    log_info "ACM mTLS Certificate Generation"
    echo ""

    check_prerequisites
    create_cert_dir
    generate_ca
    generate_server_cert
    generate_client_cert
    set_permissions
    verify_certs
    create_gitignore
    create_readme
    display_cert_info
    print_summary
}

# Run main function
main "$@"
