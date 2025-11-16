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
			IdHash:     hashID(cred.ID),
			Site:       cred.Site,
			Username:   cred.Username,
			BreachName: cred.BreachName,
			BreachDate: cred.BreachDate.Unix(),
			Severity:   mapSeverity(cred),
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
		ExcludeAmbiguous: req.Policy.ExcludeAmbiguous,
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

	// Create a compromised credential struct (we need the actual ID, not the hash)
	// In real implementation, we'd need to look this up from a mapping table
	cred := pwmanager.CompromisedCredential{
		ID: req.CredentialIdHash, // This should be unhashed in production
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
		NewPassword:  newPassword,
		RotationTime: result.Duration.Milliseconds(),
		AuditEventId: result.AuditEventID,
	}, nil
}

// GetRotationStatus retrieves the status of a rotation operation.
func (s *CredentialServiceServer) GetRotationStatus(ctx context.Context, req *acmv1.StatusRequest) (*acmv1.StatusResponse, error) {
	// For Phase I, rotations are synchronous, so we just return the completed status
	// Phase II could implement async rotations with status tracking

	return &acmv1.StatusResponse{
		Status: &acmv1.Status{
			Code:    acmv1.StatusCode_STATUS_CODE_SUCCESS,
			Message: "Rotation status check not yet implemented",
		},
		RotationState: acmv1.RotationState_ROTATION_STATE_COMPLETED,
	}, nil
}

// ListCredentials retrieves all credentials from the password vault.
func (s *CredentialServiceServer) ListCredentials(ctx context.Context, req *acmv1.ListRequest) (*acmv1.ListResponse) {
	// This would query the password manager for all credentials
	// For Phase I, we return a not implemented message

	return &acmv1.ListResponse{
		Status: &acmv1.Status{
			Code:    acmv1.StatusCode_STATUS_CODE_NOT_IMPLEMENTED,
			Message: "List credentials not yet implemented",
		},
	}
}

// GeneratePassword generates a secure password based on policy.
func (s *CredentialServiceServer) GeneratePassword(ctx context.Context, req *acmv1.GenerateRequest) (*acmv1.GenerateResponse, error) {
	policy := pwmanager.PasswordPolicy{
		Length:           int(req.Policy.Length),
		RequireUppercase: req.Policy.RequireUppercase,
		RequireLowercase: req.Policy.RequireLowercase,
		RequireNumbers:   req.Policy.RequireNumbers,
		RequireSymbols:   req.Policy.RequireSymbols,
		ExcludeAmbiguous: req.Policy.ExcludeAmbiguous,
	}

	password, err := s.crs.GeneratePassword(ctx, policy)
	if err != nil {
		return &acmv1.GenerateResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: fmt.Sprintf("Failed to generate password: %v", err),
			},
		}, nil
	}

	return &acmv1.GenerateResponse{
		Status: &acmv1.Status{
			Code:    acmv1.StatusCode_STATUS_CODE_SUCCESS,
			Message: "Password generated successfully",
		},
		Password: password,
	}, nil
}

// Helper functions

func hashID(id string) string {
	// In production, use SHA-256 hashing
	// For now, return as-is (should be fixed in Phase I cleanup)
	return id
}

func mapSeverity(cred pwmanager.CompromisedCredential) acmv1.BreachSeverity {
	// Simple severity mapping based on breach age
	// In Phase II, integrate with breach database for actual severity
	if cred.BreachDate.IsZero() {
		return acmv1.BreachSeverity_BREACH_SEVERITY_MEDIUM
	}

	// Breaches within the last year are considered high severity
	if cred.BreachDate.Year() >= 2024 {
		return acmv1.BreachSeverity_BREACH_SEVERITY_HIGH
	}

	return acmv1.BreachSeverity_BREACH_SEVERITY_MEDIUM
}
