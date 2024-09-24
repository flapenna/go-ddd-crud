package grpc

import (
	"context"
	pb "github.com/flapenna/go-ddd-crud/pkg/pb/health/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

type HealthServiceServer struct {
	pb.UnimplementedHealthServiceServer
}

func NewHealthServiceServer() *HealthServiceServer {
	return &HealthServiceServer{}
}
func (s *HealthServiceServer) Health(context.Context, *emptypb.Empty) (*pb.HealthResponse, error) {
	return &pb.HealthResponse{Status: "OK"}, nil
}
