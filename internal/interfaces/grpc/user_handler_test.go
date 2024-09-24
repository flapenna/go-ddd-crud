//go:build unit

package grpc_test

import (
	"context"
	"errors"
	"github.com/flapenna/go-ddd-crud/internal/interfaces/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"

	"github.com/flapenna/go-ddd-crud/internal/domain/user"
	grpcServer "github.com/flapenna/go-ddd-crud/internal/interfaces/grpc"
	"github.com/flapenna/go-ddd-crud/mocks"
	pb "github.com/flapenna/go-ddd-crud/pkg/pb/user/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestUserServiceServer_CreateUser(t *testing.T) {
	id := uuid.NewString()
	now := time.Now()
	tests := []struct {
		name         string
		req          *pb.CreateUserRequest
		mockResponse *domain.User
		wantedRes    *pb.User
		mockError    error
		wantedErr    error
	}{
		{
			name: "successful creation",
			req: &pb.CreateUserRequest{
				FirstName: "Federico",
				LastName:  "La Penna",
				Email:     "flapenna@email.com",
				Country:   "IT",
				Nickname:  "Pennino",
				Password:  "password",
			},
			mockResponse: &domain.User{
				ID:        id,
				FirstName: "Federico",
				LastName:  "La Penna",
				Email:     "flapenna@email.com",
				Country:   "IT",
				Nickname:  "Pennino",
				CreatedAt: now,
				UpdatedAt: now,
			},
			wantedRes: &pb.User{
				Id:        id,
				FirstName: "Federico",
				LastName:  "La Penna",
				Email:     "flapenna@email.com",
				Country:   "IT",
				Nickname:  "Pennino",
				CreatedAt: timestamppb.New(now),
				UpdatedAt: timestamppb.New(now),
			},
			mockError: nil,
			wantedErr: nil,
		},
		{
			name: "service error",
			req: &pb.CreateUserRequest{
				FirstName: "Federico",
				LastName:  "La Penna",
				Email:     "flapenna@email.com",
				Country:   "IT",
				Nickname:  "Pennino",
				Password:  "password",
			},
			mockResponse: nil,
			wantedRes:    nil,
			mockError:    errors.New("service error"),
			wantedErr:    status.Error(codes.Internal, "internal server error"),
		},
		{
			name: "validation error",
			req: &pb.CreateUserRequest{
				FirstName: "Federico",
				LastName:  "La Penna",
				Email:     "flapenna@email.com",
				Country:   "ITALIA", // not valid Country (IT)
				Nickname:  "Pennino",
				Password:  "password",
			},
			mockResponse: nil,
			wantedRes:    nil,
			mockError:    nil,
			wantedErr:    status.Error(codes.InvalidArgument, "invalid CreateUserRequest.Country: value does not match regex pattern \"^[A-Z]{2}$\""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(mocks.MockUserService)
			server := grpc.NewUserServiceServer(mockUserService)

			ctx := context.TODO()

			mockUserService.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(tt.mockResponse, tt.mockError).Once()

			resp, err := server.CreateUser(ctx, tt.req)
			if tt.wantedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantedRes, resp)
			}

		})
	}
}

func TestUserServiceServer_UpdateUser(t *testing.T) {
	userId := uuid.NewString()
	now := time.Now()
	tests := []struct {
		name         string
		req          *pb.UpdateUserRequest
		mockResponse *domain.User
		wantedRes    *pb.User
		mockError    error
		wantedErr    error
	}{
		{
			name: "successful update",
			req: &pb.UpdateUserRequest{
				Id:        userId,
				FirstName: "Federico",
				LastName:  "La Penna",
				Email:     "flapenna@email.com",
				Country:   "IT",
				Nickname:  "Pennino",
			},
			mockResponse: &domain.User{
				ID:        userId,
				FirstName: "Federico",
				LastName:  "La Penna",
				Email:     "flapenna@email.com",
				Country:   "IT",
				Nickname:  "Pennino",
				CreatedAt: now.Add(-time.Hour),
				UpdatedAt: now,
			},
			wantedRes: &pb.User{
				Id:        userId,
				FirstName: "Federico",
				LastName:  "La Penna",
				Email:     "flapenna@email.com",
				Country:   "IT",
				Nickname:  "Pennino",
				CreatedAt: timestamppb.New(now.Add(-time.Hour)),
				UpdatedAt: timestamppb.New(now),
			},
			mockError: nil,
			wantedErr: nil,
		},
		{
			name: "return empty slice of users when service returns nil",
			req: &pb.UpdateUserRequest{
				Id:        userId,
				FirstName: "Federico",
				LastName:  "La Penna",
				Email:     "flapenna@email.com",
				Country:   "IT",
				Nickname:  "Pennino",
			},
			mockResponse: nil,
			wantedRes:    nil,
			mockError:    nil,
			wantedErr:    nil,
		},
		{
			name: "user not found",
			req: &pb.UpdateUserRequest{
				Id:        userId,
				FirstName: "Federico",
				LastName:  "La Penna",
				Email:     "flapenna@email.com",
				Country:   "IT",
				Nickname:  "Pennino",
			},
			mockResponse: nil,
			wantedRes:    nil,
			mockError:    domain.ErrUserNotFound,
			wantedErr:    status.Errorf(codes.NotFound, domain.ErrUserNotFound.Error()),
		},
		{
			name: "service error",
			req: &pb.UpdateUserRequest{
				Id:        userId,
				FirstName: "Federico",
				LastName:  "La Penna",
				Email:     "flapenna@email.com",
				Country:   "IT",
				Nickname:  "Pennino",
			},
			mockResponse: nil,
			wantedRes:    nil,
			mockError:    errors.New("service error"),
			wantedErr:    status.Error(codes.Internal, "internal server error"),
		},
		{
			name: "validation error",
			req: &pb.UpdateUserRequest{
				Id:        userId,
				FirstName: "Federico",
				LastName:  "La Penna",
				Email:     "flapenna@email.com",
				Country:   "IT",
				Nickname:  "",
			},
			mockResponse: nil,
			wantedRes:    nil,
			mockError:    nil,
			wantedErr:    status.Error(codes.InvalidArgument, "invalid UpdateUserRequest.Nickname: value length must be between 2 and 50 runes, inclusive"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(mocks.MockUserService)
			server := grpcServer.NewUserServiceServer(mockUserService)
			ctx := context.TODO()

			mockUserService.On("UpdateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(tt.mockResponse, tt.mockError).Once()

			resp, err := server.UpdateUser(ctx, tt.req)
			if tt.wantedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantedRes, resp)
			}

		})
	}
}

func TestUserServiceServer_DeleteUser(t *testing.T) {
	tests := []struct {
		name      string
		req       *pb.DeleteUserRequest
		mockError error
		wantedErr error
	}{
		{
			name:      "successful deletion",
			req:       &pb.DeleteUserRequest{Id: uuid.NewString()},
			mockError: nil,
			wantedErr: nil,
		},
		{
			name:      "user not found",
			req:       &pb.DeleteUserRequest{Id: uuid.NewString()},
			mockError: domain.ErrUserNotFound,
			wantedErr: status.Error(codes.NotFound, domain.ErrUserNotFound.Error()),
		},
		{
			name:      "service error",
			req:       &pb.DeleteUserRequest{Id: uuid.NewString()},
			mockError: errors.New("service error"),
			wantedErr: status.Error(codes.Internal, "internal server error"),
		},
		{
			name:      "validation error",
			req:       &pb.DeleteUserRequest{Id: "not-uuid"}, // not uuid
			mockError: errors.New("service error"),
			wantedErr: status.Error(codes.InvalidArgument, "invalid DeleteUserRequest.Id: value must be a valid UUID | caused by: invalid uuid format"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(mocks.MockUserService)
			server := grpcServer.NewUserServiceServer(mockUserService)

			ctx := context.TODO()

			mockUserService.On("DeleteUser", mock.Anything, tt.req.Id).Return(tt.mockError).Once()

			_, err := server.DeleteUser(ctx, tt.req)
			if tt.wantedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

		})
	}
}

func TestUserServiceServer_ListUsers(t *testing.T) {
	now := time.Now()
	id := uuid.NewString()
	invalidFirstName := ""

	tests := []struct {
		name         string
		req          *pb.ListUsersRequest
		mockResponse *domain.ListUsersQueryResponse
		mockError    error
		wantedRes    *pb.ListUsersResponse
		wantedErr    error
	}{
		{
			name: "successful listing with no users",
			req:  &pb.ListUsersRequest{},
			mockResponse: &domain.ListUsersQueryResponse{
				Page:       0,
				PageSize:   10,
				TotalCount: 0,
				Results:    []*domain.User{},
			},
			mockError: nil,
			wantedRes: &pb.ListUsersResponse{
				Page:       0,
				PageSize:   10,
				TotalCount: 0,
				Results:    []*pb.User{},
			},
			wantedErr: nil,
		},
		{
			name: "successful listing",
			req:  &pb.ListUsersRequest{},
			mockResponse: &domain.ListUsersQueryResponse{
				Page:       0,
				PageSize:   10,
				TotalCount: 1,
				Results: []*domain.User{
					{
						ID:        id,
						FirstName: "Federico",
						LastName:  "La Penna",
						Email:     "email@email.com",
						Country:   "IT",
						Nickname:  "Pennino",
						CreatedAt: now,
						UpdatedAt: now,
					},
				},
			},
			mockError: nil,
			wantedRes: &pb.ListUsersResponse{
				Page:       0,
				PageSize:   10,
				TotalCount: 1,
				Results: []*pb.User{
					{
						Id:        id,
						FirstName: "Federico",
						LastName:  "La Penna",
						Email:     "email@email.com",
						Country:   "IT",
						Nickname:  "Pennino",
						CreatedAt: timestamppb.New(now),
						UpdatedAt: timestamppb.New(now),
					},
				},
			},
			wantedErr: nil,
		},
		{
			name:         "service error",
			req:          &pb.ListUsersRequest{},
			mockResponse: nil,
			mockError:    errors.New("service error"),
			wantedRes:    nil,
			wantedErr:    status.Error(codes.Internal, "internal server error"),
		},
		{
			name: "invalid error",
			req: &pb.ListUsersRequest{
				FirstName: &invalidFirstName, // not valid - at least 2 chars
			},
			mockResponse: nil,
			mockError:    nil,
			wantedRes:    nil,
			wantedErr:    status.Error(codes.InvalidArgument, "invalid ListUsersRequest.FirstName: value length must be between 2 and 50 runes, inclusive"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserService := new(mocks.MockUserService)
			server := grpcServer.NewUserServiceServer(mockUserService)

			ctx := context.TODO()

			mockUserService.On("ListUsers", mock.Anything, mock.AnythingOfType("*domain.ListUsersQueryRequest")).Return(tt.mockResponse, tt.mockError).Once()

			resp, err := server.ListUsers(ctx, tt.req)
			if tt.wantedErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantedErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantedRes.Page, resp.Page)
				assert.Equal(t, tt.wantedRes.PageSize, resp.PageSize)
				assert.Equal(t, tt.wantedRes.TotalCount, resp.TotalCount)
				assert.Equal(t, tt.wantedRes.Results, resp.Results)
			}
		})
	}
}
