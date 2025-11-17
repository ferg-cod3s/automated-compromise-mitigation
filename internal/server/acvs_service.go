// Package server provides gRPC service implementations.
package server

import (
	"context"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
	"github.com/ferg-cod3s/automated-compromise-mitigation/internal/acvs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ACVSServiceServer implements the gRPC ACVSService.
type ACVSServiceServer struct {
	acmv1.UnimplementedACVSServiceServer
	acvsService acvs.Service
}

// NewACVSServiceServer creates a new ACVS gRPC service server.
func NewACVSServiceServer(acvsService acvs.Service) *ACVSServiceServer {
	return &ACVSServiceServer{
		acvsService: acvsService,
	}
}

// AnalyzeToS implements ACVSService.AnalyzeToS.
func (s *ACVSServiceServer) AnalyzeToS(ctx context.Context, req *acmv1.AnalyzeToSRequest) (*acmv1.AnalyzeToSResponse, error) {
	if req.Site == "" {
		return nil, status.Error(codes.InvalidArgument, "site is required")
	}

	timeoutSecs := req.TimeoutSeconds
	if timeoutSecs == 0 {
		timeoutSecs = 30
	}

	crc, err := s.acvsService.AnalyzeToS(ctx, req.Site, req.TosUrl, req.ForceRefresh, timeoutSecs)
	if err != nil {
		return &acmv1.AnalyzeToSResponse{
			Status:       acmv1.AnalysisStatus_ANALYSIS_STATUS_FAILED,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &acmv1.AnalyzeToSResponse{
		Status: acmv1.AnalysisStatus_ANALYSIS_STATUS_SUCCESS,
		Crc:    crc,
		Metadata: &acmv1.AnalysisMetadata{
			FromCache:        false,
			NlpModelVersion:  "stub-v1.0.0",
			SectionsAnalyzed: int32(len(crc.Rules)),
			RulesExtracted:   int32(len(crc.Rules)),
		},
	}, nil
}

// ValidateAction implements ACVSService.ValidateAction.
func (s *ACVSServiceServer) ValidateAction(ctx context.Context, req *acmv1.ValidateActionRequest) (*acmv1.ValidateActionResponse, error) {
	if req.Site == "" {
		return nil, status.Error(codes.InvalidArgument, "site is required")
	}

	if req.Action == nil {
		return nil, status.Error(codes.InvalidArgument, "action is required")
	}

	result, err := s.acvsService.ValidateAction(ctx, req.Site, req.Action, req.CredentialId, req.ForceRefresh)
	if err != nil {
		return &acmv1.ValidateActionResponse{
			Result:       acmv1.ValidationResult_VALIDATION_RESULT_DISABLED,
			ErrorMessage: err.Error(),
		}, nil
	}

	return &acmv1.ValidateActionResponse{
		Result:            result.Result,
		RecommendedMethod: result.RecommendedMethod,
		ApplicableRuleIds: result.ApplicableRuleIDs,
		Reasoning:         result.Reasoning,
		EvidenceEntryId:   result.EvidenceEntryID,
		ErrorMessage:      result.ErrorMessage,
	}, nil
}

// GetCRC implements ACVSService.GetCRC.
func (s *ACVSServiceServer) GetCRC(ctx context.Context, req *acmv1.GetCRCRequest) (*acmv1.GetCRCResponse, error) {
	if req.Site == "" {
		return nil, status.Error(codes.InvalidArgument, "site is required")
	}

	crc, found, err := s.acvsService.GetCRC(ctx, req.Site)
	if err != nil {
		return &acmv1.GetCRCResponse{
			Status:       acmv1.CRCStatus_CRC_STATUS_DISABLED,
			ErrorMessage: err.Error(),
		}, nil
	}

	if !found {
		return &acmv1.GetCRCResponse{
			Status:       acmv1.CRCStatus_CRC_STATUS_NOT_FOUND,
			ErrorMessage: "CRC not found in cache",
		}, nil
	}

	return &acmv1.GetCRCResponse{
		Status: acmv1.CRCStatus_CRC_STATUS_FOUND,
		Crc:    crc,
	}, nil
}

// ListCRCs implements ACVSService.ListCRCs.
func (s *ACVSServiceServer) ListCRCs(ctx context.Context, req *acmv1.ListCRCsRequest) (*acmv1.ListCRCsResponse, error) {
	summaries, err := s.acvsService.ListCRCs(ctx, req.SiteFilter, req.IncludeExpired)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list CRCs: %v", err)
	}

	protoSummaries := make([]*acmv1.CRCSummary, len(summaries))
	for i, summary := range summaries {
		protoSummaries[i] = &acmv1.CRCSummary{
			Id:             summary.ID,
			Site:           summary.Site,
			ParsedAt:       timestamppb.New(summary.ParsedAt),
			ExpiresAt:      timestamppb.New(summary.ExpiresAt),
			Recommendation: summary.Recommendation,
			RuleCount:      summary.RuleCount,
			Expired:        summary.Expired,
		}
	}

	return &acmv1.ListCRCsResponse{
		Crcs: protoSummaries,
	}, nil
}

// InvalidateCRC implements ACVSService.InvalidateCRC.
func (s *ACVSServiceServer) InvalidateCRC(ctx context.Context, req *acmv1.InvalidateCRCRequest) (*acmv1.InvalidateCRCResponse, error) {
	if req.Site == "" {
		return nil, status.Error(codes.InvalidArgument, "site is required")
	}

	err := s.acvsService.InvalidateCRC(ctx, req.Site)
	if err != nil {
		return &acmv1.InvalidateCRCResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &acmv1.InvalidateCRCResponse{
		Success: true,
		Message: "CRC invalidated successfully",
	}, nil
}

// ExportEvidenceChain implements ACVSService.ExportEvidenceChain.
func (s *ACVSServiceServer) ExportEvidenceChain(req *acmv1.ExportEvidenceChainRequest, stream acmv1.ACVSService_ExportEvidenceChainServer) error {
	ctx := stream.Context()

	exportReq := &acvs.ExportRequest{
		CredentialID:        req.CredentialId,
		Format:              req.Format,
		IncludeCRCSnapshots: req.IncludeCrcSnapshots,
	}

	if req.StartTime != nil {
		exportReq.StartTime = req.StartTime.AsTime()
	}

	if req.EndTime != nil {
		exportReq.EndTime = req.EndTime.AsTime()
	}

	entries, err := s.acvsService.ExportEvidenceChain(ctx, exportReq)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to export evidence chain: %v", err)
	}

	// Stream entries to client
	for _, entry := range entries {
		if err := stream.Send(entry); err != nil {
			return status.Errorf(codes.Internal, "failed to send entry: %v", err)
		}
	}

	return nil
}

// GetACVSStatus implements ACVSService.GetACVSStatus.
func (s *ACVSServiceServer) GetACVSStatus(ctx context.Context, req *acmv1.GetACVSStatusRequest) (*acmv1.GetACVSStatusResponse, error) {
	acvsStatus, err := s.acvsService.GetStatus(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get status: %v", err)
	}

	stats, err := s.acvsService.GetStatistics(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get statistics: %v", err)
	}

	response := &acmv1.GetACVSStatusResponse{
		Enabled:     acvsStatus.Enabled,
		EulaVersion: acvsStatus.EULAVersion,
		Config: &acmv1.ACVSConfiguration{
			NlpModelVersion:      acvsStatus.Configuration.NLPModelVersion,
			CacheTtlSeconds:      acvsStatus.Configuration.CacheTTLSeconds,
			EvidenceChainEnabled: acvsStatus.Configuration.EvidenceChainEnabled,
			DefaultOnUncertain:   acvsStatus.Configuration.DefaultOnUncertain,
			ModelPath:            acvsStatus.Configuration.ModelPath,
		},
		Stats: &acmv1.ACVSStatistics{
			TotalAnalyses:          stats.TotalAnalyses,
			TotalValidations:       stats.TotalValidations,
			ValidationsAllowed:     stats.ValidationsAllowed,
			ValidationsHimRequired: stats.ValidationsHIMRequired,
			ValidationsBlocked:     stats.ValidationsBlocked,
			CrcsCached:             stats.CRCsCached,
			EvidenceEntries:        stats.EvidenceEntries,
		},
	}

	if !acvsStatus.EnabledAt.IsZero() {
		response.EnabledAt = timestamppb.New(acvsStatus.EnabledAt)
	}

	return response, nil
}

// EnableACVS implements ACVSService.EnableACVS.
func (s *ACVSServiceServer) EnableACVS(ctx context.Context, req *acmv1.EnableACVSRequest) (*acmv1.EnableACVSResponse, error) {
	if req.EulaVersion == "" {
		return nil, status.Error(codes.InvalidArgument, "EULA version is required")
	}

	if !req.UserConsent {
		return nil, status.Error(codes.FailedPrecondition, "user consent is required")
	}

	err := s.acvsService.Enable(ctx, req.EulaVersion, req.UserConsent)
	if err != nil {
		return &acmv1.EnableACVSResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	// Get updated status
	statusResp, err := s.GetACVSStatus(ctx, &acmv1.GetACVSStatusRequest{})
	if err != nil {
		statusResp = nil
	}

	return &acmv1.EnableACVSResponse{
		Success: true,
		Message: "ACVS enabled successfully",
		Status:  statusResp,
	}, nil
}

// DisableACVS implements ACVSService.DisableACVS.
func (s *ACVSServiceServer) DisableACVS(ctx context.Context, req *acmv1.DisableACVSRequest) (*acmv1.DisableACVSResponse, error) {
	err := s.acvsService.Disable(ctx, req.ClearCache, req.PreserveEvidence)
	if err != nil {
		return &acmv1.DisableACVSResponse{
			Success: false,
			Message: err.Error(),
		}, nil
	}

	return &acmv1.DisableACVSResponse{
		Success: true,
		Message: "ACVS disabled successfully",
	}, nil
}
