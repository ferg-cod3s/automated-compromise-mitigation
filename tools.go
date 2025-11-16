//go:build tools
// +build tools

// Package tools tracks development tool dependencies
// This file ensures that `go mod` captures tool dependencies
package tools

import (
	// Code generation tools
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"

	// Testing and mocking tools
	_ "go.uber.org/mock/mockgen"

	// Security scanning
	_ "github.com/securego/gosec/v2/cmd/gosec"

	// Vulnerability checking
	_ "golang.org/x/vuln/cmd/govulncheck"

	// Code formatting and imports
	_ "golang.org/x/tools/cmd/goimports"

	// Better test output
	_ "gotest.tools/gotestsum"
)

// To install all tools, run:
//   go install $(go list -f '{{join .Imports " "}}' tools.go)
