#!/usr/bin/env bash
# Cross-platform build script for ACM
# Builds binaries for all supported platforms

set -euo pipefail

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$SCRIPT_DIR")"

# Build configuration
VERSION_FILE="${ROOT_DIR}/VERSION"
VERSION="${VERSION:-$(cat "$VERSION_FILE" 2>/dev/null || echo "0.1.0-dev")}"
GIT_COMMIT="${GIT_COMMIT:-$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")}"
BUILD_DATE="${BUILD_DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"

# Output directory
BUILD_DIR="${ROOT_DIR}/build"
DIST_DIR="${ROOT_DIR}/dist"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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
    log_info "Checking prerequisites..."

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go version: $GO_VERSION"

    log_success "Prerequisites checked"
}

# Clean build directories
clean() {
    log_info "Cleaning build directories..."
    rm -rf "$BUILD_DIR" "$DIST_DIR"
    mkdir -p "$BUILD_DIR" "$DIST_DIR"
    log_success "Cleaned build directories"
}

# Build for a specific platform
build_platform() {
    local GOOS=$1
    local GOARCH=$2
    local BINARY_NAME=$3
    local CMD_PATH=$4

    local OUTPUT_DIR="${DIST_DIR}/${GOOS}_${GOARCH}"
    local OUTPUT_NAME="${BINARY_NAME}"

    # Add .exe extension for Windows
    if [ "$GOOS" = "windows" ]; then
        OUTPUT_NAME="${BINARY_NAME}.exe"
    fi

    mkdir -p "$OUTPUT_DIR"

    log_info "Building ${BINARY_NAME} for ${GOOS}/${GOARCH}..."

    # Build flags
    local LDFLAGS="-s -w"
    LDFLAGS="${LDFLAGS} -X main.Version=${VERSION}"
    LDFLAGS="${LDFLAGS} -X main.GitCommit=${GIT_COMMIT}"
    LDFLAGS="${LDFLAGS} -X main.BuildDate=${BUILD_DATE}"

    # Build
    CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH go build \
        -ldflags "${LDFLAGS}" \
        -tags "netgo osusergo" \
        -trimpath \
        -o "${OUTPUT_DIR}/${OUTPUT_NAME}" \
        "${CMD_PATH}" 2>&1

    if [ $? -eq 0 ]; then
        # Get file size
        local SIZE=$(du -h "${OUTPUT_DIR}/${OUTPUT_NAME}" | cut -f1)
        log_success "Built ${OUTPUT_NAME} for ${GOOS}/${GOARCH} (${SIZE})"
        return 0
    else
        log_error "Failed to build ${BINARY_NAME} for ${GOOS}/${GOARCH}"
        return 1
    fi
}

# Build all platforms
build_all() {
    log_info "Building ACM for all platforms..."
    echo ""

    # Define platforms to build
    declare -a PLATFORMS=(
        "linux:amd64"
        "linux:arm64"
        "darwin:amd64"
        "darwin:arm64"
        "windows:amd64"
    )

    # Define binaries to build
    declare -a BINARIES=(
        "acm-service:./cmd/acm-service"
        "acm-cli:./cmd/acm-cli"
    )

    local TOTAL_BUILDS=$((${#PLATFORMS[@]} * ${#BINARIES[@]}))
    local CURRENT_BUILD=0
    local FAILED_BUILDS=0

    # Build each binary for each platform
    for PLATFORM in "${PLATFORMS[@]}"; do
        IFS=':' read -r GOOS GOARCH <<< "$PLATFORM"

        for BINARY in "${BINARIES[@]}"; do
            IFS=':' read -r BINARY_NAME CMD_PATH <<< "$BINARY"

            CURRENT_BUILD=$((CURRENT_BUILD + 1))
            echo -e "${BLUE}[${CURRENT_BUILD}/${TOTAL_BUILDS}]${NC} Building ${BINARY_NAME} for ${GOOS}/${GOARCH}..."

            if ! build_platform "$GOOS" "$GOARCH" "$BINARY_NAME" "$CMD_PATH"; then
                FAILED_BUILDS=$((FAILED_BUILDS + 1))
            fi
        done
    done

    echo ""
    if [ $FAILED_BUILDS -eq 0 ]; then
        log_success "All builds completed successfully!"
    else
        log_warn "${FAILED_BUILDS} build(s) failed"
    fi

    return $FAILED_BUILDS
}

# Create archives
create_archives() {
    log_info "Creating release archives..."

    cd "$DIST_DIR"

    for PLATFORM_DIR in */; do
        PLATFORM_NAME="${PLATFORM_DIR%/}"
        ARCHIVE_NAME="acm_${VERSION}_${PLATFORM_NAME}"

        # Determine archive format
        if [[ "$PLATFORM_NAME" == windows_* ]]; then
            ARCHIVE_FILE="${ARCHIVE_NAME}.zip"
            log_info "Creating ${ARCHIVE_FILE}..."

            # Create zip for Windows
            (cd "$PLATFORM_NAME" && zip -q -r "../${ARCHIVE_FILE}" .)
        else
            ARCHIVE_FILE="${ARCHIVE_NAME}.tar.gz"
            log_info "Creating ${ARCHIVE_FILE}..."

            # Create tar.gz for Unix-like systems
            tar -czf "$ARCHIVE_FILE" -C "$PLATFORM_NAME" .
        fi

        if [ $? -eq 0 ]; then
            SIZE=$(du -h "$ARCHIVE_FILE" | cut -f1)
            log_success "Created ${ARCHIVE_FILE} (${SIZE})"
        else
            log_error "Failed to create ${ARCHIVE_FILE}"
        fi
    done

    cd "$ROOT_DIR"
}

# Generate checksums
generate_checksums() {
    log_info "Generating checksums..."

    cd "$DIST_DIR"

    # Generate SHA256 checksums
    if command -v sha256sum &> /dev/null; then
        sha256sum *.tar.gz *.zip 2>/dev/null > checksums.txt
    elif command -v shasum &> /dev/null; then
        shasum -a 256 *.tar.gz *.zip 2>/dev/null > checksums.txt
    else
        log_warn "Neither sha256sum nor shasum found, skipping checksums"
        cd "$ROOT_DIR"
        return
    fi

    if [ -f checksums.txt ]; then
        log_success "Generated checksums.txt"
        cat checksums.txt
    fi

    cd "$ROOT_DIR"
}

# Print build summary
print_summary() {
    log_info "Build Summary"
    echo ""
    echo "Version:    $VERSION"
    echo "Git Commit: $GIT_COMMIT"
    echo "Build Date: $BUILD_DATE"
    echo ""
    echo "Build output: ${DIST_DIR}"
    echo ""

    if [ -d "$DIST_DIR" ]; then
        log_info "Built artifacts:"
        ls -lh "$DIST_DIR" | grep -v "^d" | grep -v "^total" || true
    fi
}

# Main execution
main() {
    log_info "ACM Cross-Platform Build Script"
    log_info "Version: $VERSION"
    echo ""

    check_prerequisites
    clean

    if build_all; then
        create_archives
        generate_checksums
        print_summary
        log_success "Build completed successfully!"
        exit 0
    else
        log_error "Build failed!"
        exit 1
    fi
}

# Run main function
main "$@"
