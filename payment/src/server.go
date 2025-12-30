package src

import (
	"common/helpers/logger"
	"context"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// healthServer struct.
// This struct is used to implement the health service gRPC server.
type healthServer struct {
	grpc_health_v1.UnimplementedHealthServer
}

// Check method.
// This method is used to check the health of the service.
//
// Parameters:
//   - ctx (context.Context): The context.
//   - req (*grpc_health_v1.HealthCheckRequest): The health check request.
//
// Returns:
//   - *grpc_health_v1.HealthCheckResponse: The health check response.
//   - error: The error.
func (s *healthServer) Check(ctx context.Context, req *grpc_health_v1.HealthCheckRequest) (*grpc_health_v1.HealthCheckResponse, error) {
	return &grpc_health_v1.HealthCheckResponse{Status: grpc_health_v1.HealthCheckResponse_SERVING}, nil
}

// ListenAndServeGRPC function.
// This function is used to listen and serve the payment service gRPC server.
//
// Parameters:
//   - port (int): The port.
//
// Returns:
//   - error: The error.
func ListenAndServeGRPC(port int) error {
	log := logger.NewLogger()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.WithFields(logrus.Fields{"port": port, "error": err}).Error("Failed to Listen!")
		return fmt.Errorf("failed to listen: %w", err)
	}

	server := grpc.NewServer()

	grpc_health_v1.RegisterHealthServer(server, &healthServer{})
	reflection.Register(server)

	log.WithFields(logrus.Fields{"port": port}).Info("Starting gRPC Server...")
	if err := server.Serve(lis); err != nil {
		log.WithFields(logrus.Fields{"error": err}).Error("Failed to Serve gRPC Server!")
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}
