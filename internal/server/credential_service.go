// Package server implements the gRPC service handlers for ACM.
package server

import (
	"context"
	"fmt"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/crs"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/pwmanager"
)

// CredentialServiceServer implements the gRPC CredentialService.
type CredentialServiceServer struct {
	acmv1.UnimplementedCredentialServiceServer
	crs *crs.Service
}

// NewCredentialServiceServer creates a new credential service server.
func NewCredentialServiceServer(crsService *crs.Service) *CredentialServiceServer {
	return &CredentialServiceServer{
		crs: crsService,
	}
}

// DetectCompromised queries the password manager for compromised credentials.
func (s *CredentialServiceServer) DetectCompromised(ctx context.Context, req *acmv1.DetectRequest) (*acmv1.DetectResponse, error) {
	// Call CRS to detect compromised credentials
	creds, err := s.crs.DetectCompromised(ctx)
	if err != nil {
		return &acmv1.DetectResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: fmt.Sprintf("Failed to detect compromised credentials: %v", err),
			},
		}, nil
	}

	// Convert to proto format
	protoCredentials := make([]*acmv1.CompromisedCredential, 0, len(creds))
	for _, cred := range creds {
		protoCred := &acmv1.CompromisedCredential{
			IdHash:     cred.ID, // Should be hashed in production
			Site:       cred.Site,
			Username:   cred.Username,
			BreachName: cred.BreachName,
			BreachDate: cred.BreachDate.Unix(),
			Severity:   acmv1.BreachSeverity_BREACH_SEVERITY_MEDIUM,
		}
		protoCredentials = append(protoCredentials, protoCred)
	}

	return &acmv1.DetectResponse{
		Status: &acmv1.Status{
			Code:    acmv1.StatusCode_STATUS_CODE_SUCCESS,
			Message: fmt.Sprintf("Found %d compromised credentials", len(creds)),
		},
		Credentials: protoCredentials,
		TotalCount:  int32(len(creds)),
	}, nil
}

// RotateCredential performs a credential rotation operation.
func (s *CredentialServiceServer) RotateCredential(ctx context.Context, req *acmv1.RotateRequest) (*acmv1.RotateResponse, error) {
	// Generate password based on policy
	policy := pwmanager.PasswordPolicy{
		Length:           int(req.Policy.Length),
		RequireUppercase: req.Policy.RequireUppercase,
		RequireLowercase: req.Policy.RequireLowercase,
		RequireNumbers:   req.Policy.RequireNumbers,
		RequireSymbols:   req.Policy.RequireSymbols,
	}

	newPassword, err := s.crs.GeneratePassword(ctx, policy)
	if err != nil {
		return &acmv1.RotateResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: fmt.Sprintf("Failed to generate password: %v", err),
			},
		}, nil
	}

	// Create a compromised credential struct
	cred := pwmanager.CompromisedCredential{
		ID: req.CredentialIdHash,
	}

	// Perform rotation
	result, err := s.crs.RotateCredential(ctx, cred, newPassword)
	if err != nil {
		statusCode := acmv1.StatusCode_STATUS_CODE_FAILURE
		if result != nil && result.Status == crs.RotationHIMRequired {
			statusCode = acmv1.StatusCode_STATUS_CODE_HIM_REQUIRED
		}

		return &acmv1.RotateResponse{
			Status: &acmv1.Status{
				Code:    statusCode,
				Message: err.Error(),
			},
		}, nil
	}

	return &acmv1.RotateResponse{
		Status: &acmv1.Status{
			Code:    acmv1.StatusCode_STATUS_CODE_SUCCESS,
			Message: "Credential rotated successfully",
		},
		NewPassword: newPassword,
	}, nil
}

// GetRotationStatus retrieves the status of a rotation operation.
func (s *CredentialServiceServer) GetRotationStatus(ctx context.Context, req *acmv1.StatusRequest) (*acmv1.StatusResponse, error) {
	// For Phase I, rotations are synchronous
	return &acmv1.StatusResponse{
		Status: &acmv1.Status{
			Code:    acmv1.StatusCode_STATUS_CODE_SUCCESS,
			Message: "Rotation status check not yet implemented",
		},
	}, nil
}

// ListCredentials retrieves all credentials from the password vault.
func (s *CredentialServiceServer) ListCredentials(ctx context.Context, req *acmv1.ListRequest) (*acmv1.ListResponse, error) {
	// For Phase I, return empty list
	return &acmv1.ListResponse{
		Status: &acmv1.Status{
			Code:    acmv1.StatusCode_STATUS_CODE_SUCCESS,
			Message: "List credentials not yet fully implemented",
		},
		Credentials: make([]*acmv1.CredentialMetadata, 0),
	}, nil
}
