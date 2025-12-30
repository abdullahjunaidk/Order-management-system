package src

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

// healthServer struct.
// This struct is used to implement the health service gRPC server.
//
// Attributes:
//   - grpc_health_v1.UnimplementedHealthServer: The unimplemented health server.
type healthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

// NewHealthServer function.
// This function is used to create a new health server.
//
// Returns:
//   - *healthServer: The health server.
//   - error: The error.
func (s *healthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{
		Status: grpc_health_v1.HealthCheckResponse_SERVING,
	}, nil
}

// Watch method.
// This method is used to watch the health status.
//
// Parameters:
//   - req (*grpc_health_v1.HealthCheckRequest): The health check request.
//   - server grpc_health_v1.Health_WatchServer: The health watch server.
//
// Returns:
//   - error: The error.
func (s *healthServer) Watch(req *grpc_health_v1.HealthCheckRequest, server grpc_health_v1.Health_WatchServer) error {
	for {
		select {
		case <-server.Context().Done():
			return status.Error(codes.Canceled, "Stream Canceled!")
		default:
			err := server.Send(&grpc_health_v1.HealthCheckResponse{
				Status: grpc_health_v1.HealthCheckResponse_SERVING,
			})
			if err != nil {
				return status.Error(codes.Internal, "Failed to Send Health Check Response!")
			}
			time.Sleep(time.Second)
		}
	}
}
