// Package server implements the gRPC service handlers for ACM.
package server

import (
	"context"
	"fmt"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/rotation/github"
)

// RotationServiceServer implements the gRPC RotationService.
type RotationServiceServer struct {
	acmv1.UnimplementedRotationServiceServer
	githubRotator *github.Rotator
}

// NewRotationServiceServer creates a new rotation service server.
func NewRotationServiceServer(githubRotator *github.Rotator) *RotationServiceServer {
	return &RotationServiceServer{
		githubRotator: githubRotator,
	}
}

// StartGitHubRotation initiates a GitHub PAT rotation workflow.
func (s *RotationServiceServer) StartGitHubRotation(ctx context.Context, req *acmv1.StartGitHubRotationRequest) (*acmv1.StartGitHubRotationResponse, error) {
	// Validate request
	if req.CredentialId == "" {
		return &acmv1.StartGitHubRotationResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: "credential_id is required",
			},
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_INVALID_REQUEST,
				Message: "credential_id is required",
			},
		}, nil
	}

	if req.CurrentToken == "" {
		return &acmv1.StartGitHubRotationResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: "current_token is required",
			},
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_INVALID_REQUEST,
				Message: "current_token is required",
			},
		}, nil
	}

	// Create rotation request
	rotationReq := github.RotationRequest{
		CredentialID: req.CredentialId,
		CurrentToken: req.CurrentToken,
		Site:         req.Site,
		Username:     req.Username,
	}

	// Start rotation
	result, err := s.githubRotator.StartRotation(ctx, rotationReq)
	if err != nil {
		return &acmv1.StartGitHubRotationResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: fmt.Sprintf("Failed to start rotation: %v", err),
			},
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_UNKNOWN,
				Message: err.Error(),
			},
		}, nil
	}

	// Check if rotation failed
	if !result.Success {
		errorCode := acmv1.ErrorCode_ERROR_CODE_UNKNOWN
		statusCode := acmv1.StatusCode_STATUS_CODE_FAILURE

		if result.Error != nil {
			// Map specific errors
			errMsg := result.Error.Error()
			if contains(errMsg, "validation failed") {
				errorCode = acmv1.ErrorCode_ERROR_CODE_INVALID_REQUEST
				statusCode = acmv1.StatusCode_STATUS_CODE_FAILURE
			} else if contains(errMsg, "ToS prohibits") {
				errorCode = acmv1.ErrorCode_ERROR_CODE_PERMISSION_DENIED
				statusCode = acmv1.StatusCode_STATUS_CODE_COMPLIANCE_BLOCKED
			}
		}

		return &acmv1.StartGitHubRotationResponse{
			Status: &acmv1.Status{
				Code:    statusCode,
				Message: result.Error.Error(),
			},
			Error: &acmv1.Error{
				Code:    errorCode,
				Message: result.Error.Error(),
			},
		}, nil
	}

	// Convert step to proto enum
	nextStep := mapRotationStepToProto(result.NextStep)

	return &acmv1.StartGitHubRotationResponse{
		Status: &acmv1.Status{
			Code:    acmv1.StatusCode_STATUS_CODE_SUCCESS,
			Message: "GitHub rotation started successfully",
		},
		StateId:      result.State.ID,
		NextStep:     nextStep,
		Instructions: result.Instructions,
		Username:     result.State.Metadata["username"],
		CrcId:        result.State.Metadata["crc_id"],
	}, nil
}

// VerifyNewToken verifies a newly created token.
func (s *RotationServiceServer) VerifyNewToken(ctx context.Context, req *acmv1.VerifyNewTokenRequest) (*acmv1.VerifyNewTokenResponse, error) {
	// Validate request
	if req.StateId == "" {
		return &acmv1.VerifyNewTokenResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: "state_id is required",
			},
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_INVALID_REQUEST,
				Message: "state_id is required",
			},
		}, nil
	}

	if req.NewToken == "" {
		return &acmv1.VerifyNewTokenResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: "new_token is required",
			},
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_INVALID_REQUEST,
				Message: "new_token is required",
			},
		}, nil
	}

	// Verify new token
	result, err := s.githubRotator.VerifyNewToken(ctx, req.StateId, req.NewToken)
	if err != nil {
		return &acmv1.VerifyNewTokenResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: fmt.Sprintf("Failed to verify token: %v", err),
			},
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_UNKNOWN,
				Message: err.Error(),
			},
		}, nil
	}

	// Check if verification failed
	if !result.Success {
		return &acmv1.VerifyNewTokenResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: result.Error.Error(),
			},
			StateId: req.StateId,
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_INVALID_REQUEST,
				Message: result.Error.Error(),
			},
		}, nil
	}

	// Convert step to proto enum
	nextStep := mapRotationStepToProto(result.NextStep)

	return &acmv1.VerifyNewTokenResponse{
		Status: &acmv1.Status{
			Code:    acmv1.StatusCode_STATUS_CODE_SUCCESS,
			Message: "New token verified successfully",
		},
		StateId:      result.State.ID,
		NextStep:     nextStep,
		Instructions: result.Instructions,
	}, nil
}

// ConfirmDeletion confirms that the old token has been deleted.
func (s *RotationServiceServer) ConfirmDeletion(ctx context.Context, req *acmv1.ConfirmDeletionRequest) (*acmv1.ConfirmDeletionResponse, error) {
	// Validate request
	if req.StateId == "" {
		return &acmv1.ConfirmDeletionResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: "state_id is required",
			},
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_INVALID_REQUEST,
				Message: "state_id is required",
			},
		}, nil
	}

	// Confirm deletion
	result, err := s.githubRotator.ConfirmDeletion(ctx, req.StateId)
	if err != nil {
		return &acmv1.ConfirmDeletionResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: fmt.Sprintf("Failed to confirm deletion: %v", err),
			},
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_UNKNOWN,
				Message: err.Error(),
			},
		}, nil
	}

	// Check if confirmation failed
	if !result.Success {
		return &acmv1.ConfirmDeletionResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: "Failed to complete rotation",
			},
			StateId: req.StateId,
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_UNKNOWN,
				Message: "Failed to complete rotation",
			},
		}, nil
	}

	return &acmv1.ConfirmDeletionResponse{
		Status: &acmv1.Status{
			Code:    acmv1.StatusCode_STATUS_CODE_SUCCESS,
			Message: "Rotation completed successfully",
		},
		StateId:     result.State.ID,
		CompletedAt: result.CompletedAt.Unix(),
		// EvidenceId would come from ACVS integration if enabled
	}, nil
}

// GetGitHubRotationStatus gets the current status of a rotation.
func (s *RotationServiceServer) GetGitHubRotationStatus(ctx context.Context, req *acmv1.GetGitHubRotationStatusRequest) (*acmv1.GetGitHubRotationStatusResponse, error) {
	// Validate request
	if req.StateId == "" {
		return &acmv1.GetGitHubRotationStatusResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: "state_id is required",
			},
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_INVALID_REQUEST,
				Message: "state_id is required",
			},
		}, nil
	}

	// Get rotation status
	result, err := s.githubRotator.GetRotationStatus(ctx, req.StateId)
	if err != nil {
		return &acmv1.GetGitHubRotationStatusResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: fmt.Sprintf("Failed to get rotation status: %v", err),
			},
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_UNKNOWN,
				Message: err.Error(),
			},
		}, nil
	}

	// Convert step to proto enum
	currentStep := mapRotationStepToProto(result.NextStep)

	response := &acmv1.GetGitHubRotationStatusResponse{
		Status: &acmv1.Status{
			Code:    acmv1.StatusCode_STATUS_CODE_SUCCESS,
			Message: "Rotation status retrieved",
		},
		StateId:     result.State.ID,
		CurrentStep: currentStep,
		Site:        result.State.Metadata["site"],
		Username:    result.State.Metadata["username"],
		StartedAt:   result.State.StartedAt.Unix(),
		UpdatedAt:   result.State.UpdatedAt.Unix(),
		ExpiresAt:   result.State.ExpiresAt.Unix(),
	}

	// Include error if rotation failed
	if !result.Success && result.Error != nil {
		response.Error = &acmv1.Error{
			Code:    acmv1.ErrorCode_ERROR_CODE_UNKNOWN,
			Message: result.Error.Error(),
		}
	}

	return response, nil
}

// CancelGitHubRotation cancels an in-progress rotation.
func (s *RotationServiceServer) CancelGitHubRotation(ctx context.Context, req *acmv1.CancelGitHubRotationRequest) (*acmv1.CancelGitHubRotationResponse, error) {
	// Validate request
	if req.StateId == "" {
		return &acmv1.CancelGitHubRotationResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: "state_id is required",
			},
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_INVALID_REQUEST,
				Message: "state_id is required",
			},
		}, nil
	}

	// Cancel rotation
	err := s.githubRotator.CancelRotation(ctx, req.StateId)
	if err != nil {
		return &acmv1.CancelGitHubRotationResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: fmt.Sprintf("Failed to cancel rotation: %v", err),
			},
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_UNKNOWN,
				Message: err.Error(),
			},
		}, nil
	}

	return &acmv1.CancelGitHubRotationResponse{
		Status: &acmv1.Status{
			Code:    acmv1.StatusCode_STATUS_CODE_SUCCESS,
			Message: "Rotation cancelled successfully",
		},
		StateId: req.StateId,
	}, nil
}

// ListActiveGitHubRotations lists all active rotations.
func (s *RotationServiceServer) ListActiveGitHubRotations(ctx context.Context, req *acmv1.ListActiveGitHubRotationsRequest) (*acmv1.ListActiveGitHubRotationsResponse, error) {
	// List active rotations
	states, err := s.githubRotator.ListActiveRotations(ctx)
	if err != nil {
		return &acmv1.ListActiveGitHubRotationsResponse{
			Status: &acmv1.Status{
				Code:    acmv1.StatusCode_STATUS_CODE_FAILURE,
				Message: fmt.Sprintf("Failed to list rotations: %v", err),
			},
			Error: &acmv1.Error{
				Code:    acmv1.ErrorCode_ERROR_CODE_UNKNOWN,
				Message: err.Error(),
			},
		}, nil
	}

	// Convert to proto format
	rotations := make([]*acmv1.GitHubRotationState, 0, len(states))
	for _, state := range states {
		// Apply filters if specified
		if req.Site != "" && state.Metadata["site"] != req.Site {
			continue
		}
		if req.Username != "" && state.Metadata["username"] != req.Username {
			continue
		}

		// Map state string to proto enum
		var currentStep acmv1.GitHubRotationStep
		switch state.State {
		case "validating":
			currentStep = acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_VALIDATING
		case "guiding":
			currentStep = acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_GUIDING
		case "waiting_deletion":
			currentStep = acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_WAITING_TOKEN
		case "verifying":
			currentStep = acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_VERIFYING
		case "complete":
			currentStep = acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_COMPLETE
		case "failed":
			currentStep = acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_FAILED
		default:
			currentStep = acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_UNSPECIFIED
		}

		rotations = append(rotations, &acmv1.GitHubRotationState{
			StateId:           state.ID,
			CredentialIdHash:  hashString(state.CredentialID),
			Site:              state.Metadata["site"],
			Username:          state.Metadata["username"],
			CurrentStep:       currentStep,
			StartedAt:         state.StartedAt.Unix(),
			UpdatedAt:         state.UpdatedAt.Unix(),
			ExpiresAt:         state.ExpiresAt.Unix(),
			CrcId:             state.Metadata["crc_id"],
		})
	}

	return &acmv1.ListActiveGitHubRotationsResponse{
		Status: &acmv1.Status{
			Code:    acmv1.StatusCode_STATUS_CODE_SUCCESS,
			Message: fmt.Sprintf("Found %d active rotations", len(rotations)),
		},
		Rotations:  rotations,
		TotalCount: int32(len(rotations)),
	}, nil
}

// mapRotationStepToProto converts internal rotation step to proto enum.
func mapRotationStepToProto(step github.RotationStep) acmv1.GitHubRotationStep {
	switch step {
	case github.StepValidating:
		return acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_VALIDATING
	case github.StepGuiding:
		return acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_GUIDING
	case github.StepWaitingToken:
		return acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_WAITING_TOKEN
	case github.StepVerifying:
		return acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_VERIFYING
	case github.StepComplete:
		return acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_COMPLETE
	case github.StepFailed:
		return acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_FAILED
	default:
		return acmv1.GitHubRotationStep_GITHUB_ROTATION_STEP_UNSPECIFIED
	}
}

// contains checks if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// hashString creates a simple hash of a string for display purposes.
// In production, this should use a proper cryptographic hash like SHA-256.
func hashString(s string) string {
	// Simple hash for now - in production use crypto/sha256
	if len(s) > 8 {
		return s[:8] + "..."
	}
	return s
}
