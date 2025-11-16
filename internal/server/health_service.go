package server

import (
	"context"
	"time"

	acmv1 "github.com/ferg-cod3s/automated-compromise-mitigation/api/proto/acm/v1"
)

// HealthServiceServer implements the gRPC HealthService.
type HealthServiceServer struct {
	acmv1.UnimplementedHealthServiceServer
}

// Check returns the health status of the service.
func (s *HealthServiceServer) Check(ctx context.Context, req *acmv1.HealthRequest) (*acmv1.HealthResponse, error) {
	return &acmv1.HealthResponse{
		Status: acmv1.HealthResponse_HEALTH_STATUS_HEALTHY,
		Components: map[string]bool{
			"credential_service": true,
			"audit_logger":       true,
			"him_service":        true,
		},
		Message:   "All systems operational",
		Timestamp: time.Now().Unix(),
	}, nil
}
