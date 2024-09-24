//go:build unit

package grpc_test

import (
	"context"
	"github.com/flapenna/go-ddd-crud/internal/interfaces/grpc"
	pb "github.com/flapenna/go-ddd-crud/pkg/pb/health/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/emptypb"
	"testing"
)

func TestHealthServiceServer_Health(t *testing.T) {
	server := grpc.NewHealthServiceServer()
	ctx := context.TODO()

	resp, err := server.Health(ctx, &emptypb.Empty{})

	assert.Nil(t, err)
	assert.Equal(t, &pb.HealthResponse{Status: "OK"}, resp)
}
