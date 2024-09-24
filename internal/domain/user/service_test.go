//go:build unit

package domain_test

import (
	"context"
	"errors"
	domain "github.com/flapenna/go-ddd-crud/internal/domain/user"
	"testing"
	"time"

	"github.com/flapenna/go-ddd-crud/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestService_CreateUser(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(repository *mocks.MockUserRepository)
		req       *domain.User
		wantErr   bool
	}{
		{
			name: "successful creation",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*domain.User)
					arg.ID = uuid.New().String()
					arg.CreatedAt = time.Now()
					arg.UpdatedAt = arg.CreatedAt
				})
			},
			req: &domain.User{
				FirstName: "Federico",
				Email:     "flapenna@email.com",
			},
			wantErr: false,
		},
		{
			name: "repository error",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("CreateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(errors.New("repository error"))
			},
			req: &domain.User{
				FirstName: "Federico",
				Email:     "flapenna@email.com",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			mockProducer := new(mocks.MockUserProducer)
			mockWatcher := new(mocks.MockUserWatcher)
			service := domain.NewUserService(mockRepo, mockProducer, mockWatcher)
			tt.setupMock(mockRepo)

			ctx := context.TODO()
			createdUser, err := service.CreateUser(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, createdUser)
				assert.NotEmpty(t, createdUser.ID)
				assert.WithinDuration(t, time.Now(), createdUser.CreatedAt, time.Second)
				assert.WithinDuration(t, createdUser.CreatedAt, createdUser.UpdatedAt, time.Second)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_UpdateUser(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(repository *mocks.MockUserRepository)
		req       *domain.User
		wantErr   bool
	}{
		{
			name: "successful update",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("UpdateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil).Run(func(args mock.Arguments) {
					arg := args.Get(1).(*domain.User)
					arg.UpdatedAt = time.Now()
				})
			},
			req: &domain.User{
				ID:        "c4fa0ff4-71a6-4010-8f1c-b9706853f8a0",
				FirstName: "Federico",
				Email:     "flapenna@email.com",
				CreatedAt: time.Now().Add(-time.Hour),
			},
			wantErr: false,
		},
		{
			name: "repository error",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("UpdateUser", mock.Anything, mock.AnythingOfType("*domain.User")).Return(errors.New("repository error"))
			},
			req: &domain.User{
				ID:        "c4fa0ff4-71a6-4010-8f1c-b9706853f8a0",
				FirstName: "Federico",
				Email:     "flapenna@email.com",
				CreatedAt: time.Now().Add(-time.Hour),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			mockProducer := new(mocks.MockUserProducer)
			mockWatcher := new(mocks.MockUserWatcher)
			service := domain.NewUserService(mockRepo, mockProducer, mockWatcher)
			tt.setupMock(mockRepo)

			ctx := context.TODO()
			updatedUser, err := service.UpdateUser(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, updatedUser)
				assert.Equal(t, tt.req.ID, updatedUser.ID)
				assert.WithinDuration(t, time.Now(), updatedUser.UpdatedAt, time.Second)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_DeleteUser(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(repository *mocks.MockUserRepository)
		userID    string
		wantErr   bool
	}{
		{
			name: "successful deletion",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("DeleteUserById", mock.Anything, "user-123").Return(nil)
			},
			userID:  "user-123",
			wantErr: false,
		},
		{
			name: "repository error",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("DeleteUserById", mock.Anything, "user-123").Return(errors.New("repository error"))
			},
			userID:  "user-123",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			mockProducer := new(mocks.MockUserProducer)
			mockWatcher := new(mocks.MockUserWatcher)
			service := domain.NewUserService(mockRepo, mockProducer, mockWatcher)
			tt.setupMock(mockRepo)

			ctx := context.TODO()
			err := service.DeleteUser(ctx, tt.userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_ListUsers(t *testing.T) {
	now := time.Now()
	wantedRes := &domain.ListUsersQueryResponse{
		Page:       0,
		PageSize:   10,
		TotalCount: 1,
		Results: []*domain.User{
			{
				ID:        "1",
				FirstName: "Federico",
				LastName:  "La Penna",
				Email:     "email@email.com",
				Country:   "IT",
				Nickname:  "Pennino",
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	tests := []struct {
		name      string
		setupMock func(repository *mocks.MockUserRepository)
		req       *domain.ListUsersQueryRequest
		wantRes   *domain.ListUsersQueryResponse
		wantErr   bool
	}{
		{
			name:    "successful listing",
			wantRes: wantedRes,
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("ListUsers", mock.Anything, mock.AnythingOfType("*domain.ListUsersQueryRequest")).Return(wantedRes, nil)
			},
			req:     &domain.ListUsersQueryRequest{},
			wantErr: false,
		},
		{
			name: "repository error",
			setupMock: func(mockRepo *mocks.MockUserRepository) {
				mockRepo.On("ListUsers", mock.Anything, mock.AnythingOfType("*domain.ListUsersQueryRequest")).Return(nil, errors.New("repository error"))
			},
			req:     &domain.ListUsersQueryRequest{},
			wantRes: nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(mocks.MockUserRepository)
			mockProducer := new(mocks.MockUserProducer)
			mockWatcher := new(mocks.MockUserWatcher)
			service := domain.NewUserService(mockRepo, mockProducer, mockWatcher)
			tt.setupMock(mockRepo)

			ctx := context.TODO()
			res, err := service.ListUsers(ctx, tt.req)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantRes.Page, res.Page)
				assert.Equal(t, tt.wantRes.PageSize, res.PageSize)
				assert.Equal(t, tt.wantRes.TotalCount, res.TotalCount)
				assert.ElementsMatch(t, tt.wantRes.Results, res.Results)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
