package grpc

import (
	"errors"
	"github.com/flapenna/go-ddd-crud/internal/domain/user"
	pb "github.com/flapenna/go-ddd-crud/pkg/pb/user/v1"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

import (
	"context"
)

type UserServiceServer struct {
	pb.UnimplementedUserServiceServer
	userService domain.UserService
}

func NewUserServiceServer(userService domain.UserService) *UserServiceServer {
	return &UserServiceServer{userService: userService}
}

func (s *UserServiceServer) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.User, error) {
	log.Info("[GRPC] CreateUser called")
	if err := req.Validate(); err != nil {
		log.Errorf("failed to validate create user request: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Errorf("error hashing password: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	user := &domain.User{
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		Country:        req.Country,
		Nickname:       req.Nickname,
		HashedPassword: string(hashedPassword),
	}

	createdUser, err := s.userService.CreateUser(ctx, user)
	if err != nil {
		log.Errorf("failed to create user: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return userToProto(createdUser), nil
}

func (s *UserServiceServer) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.User, error) {
	log.Infof("[GRPC] UpdateUser called with id %s", req.Id)
	if err := req.Validate(); err != nil {
		log.Errorf("failed to validate update user request: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	user := &domain.User{
		ID:        req.Id,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Country:   req.Country,
		Nickname:  req.Nickname,
	}

	updatedUser, err := s.userService.UpdateUser(ctx, user)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			log.Warn("trying to update user that doesn't exist")
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		log.Errorf("failed to update user: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return userToProto(updatedUser), nil
}

func (s *UserServiceServer) DeleteUser(ctx context.Context, req *pb.DeleteUserRequest) (*emptypb.Empty, error) {
	log.Infof("[GRPC] DeleteUser called with id %s", req.Id)
	if err := req.Validate(); err != nil {
		log.Errorf("failed to validate delete user request: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err := s.userService.DeleteUser(ctx, req.Id)
	if err != nil {
		log.Warn("trying to delete user that doesn't exist")
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Errorf(codes.NotFound, err.Error())
		}
		log.Errorf("failed to update user: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	return &emptypb.Empty{}, nil
}

func (s *UserServiceServer) ListUsers(ctx context.Context, req *pb.ListUsersRequest) (*pb.ListUsersResponse, error) {
	log.Infof("[GRPC] ListUsers called")
	if err := req.Validate(); err != nil {
		log.Errorf("failed to validate list users request: %v", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	listUsersRequest := &domain.ListUsersQueryRequest{
		Page:      req.Page,
		PageSize:  req.PageSize,
		Country:   req.Country,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Nickname:  req.Nickname,
		Email:     req.Email,
	}

	res, err := s.userService.ListUsers(ctx, listUsersRequest)
	if err != nil {
		log.Errorf("failed to list users: %v", err)
		return nil, status.Errorf(codes.Internal, "internal server error")
	}

	users := make([]*pb.User, len(res.Results))
	for i, u := range res.Results {
		users[i] = userToProto(u)
	}
	listUsersResponse := &pb.ListUsersResponse{
		Page:       res.Page,
		PageSize:   res.PageSize,
		TotalCount: res.TotalCount,
		Results:    users,
	}
	return listUsersResponse, nil
}

func userToProto(user *domain.User) *pb.User {
	if user == nil {
		return nil
	}
	return &pb.User{
		Id:        user.ID,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Email:     user.Email,
		Country:   user.Country,
		Nickname:  user.Nickname,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}
