package domain

import (
	"context"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"time"
)

type UserService interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	DeleteUser(ctx context.Context, id string) error
	ListUsers(ctx context.Context, request *ListUsersQueryRequest) (*ListUsersQueryResponse, error)
	StartWatchingUsers(ctx context.Context)
}

type service struct {
	repo     UserRepository
	producer UserProducer
	watcher  UserWatcher
}

func NewUserService(repo UserRepository, producer UserProducer, watcher UserWatcher) UserService {
	return &service{repo: repo, producer: producer, watcher: watcher}
}

func (s *service) CreateUser(ctx context.Context, user *User) (*User, error) {
	user.ID = uuid.NewString()
	user.CreatedAt = time.Now().UTC().Round(time.Millisecond)
	user.UpdatedAt = user.CreatedAt
	err := s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) UpdateUser(ctx context.Context, user *User) (*User, error) {
	user.UpdatedAt = time.Now().UTC().Round(time.Millisecond)
	err := s.repo.UpdateUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *service) DeleteUser(ctx context.Context, id string) error {
	return s.repo.DeleteUserById(ctx, id)
}

func (s *service) ListUsers(ctx context.Context, req *ListUsersQueryRequest) (*ListUsersQueryResponse, error) {
	users, err := s.repo.ListUsers(ctx, req)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *service) StartWatchingUsers(ctx context.Context) {
	userEvents := s.watcher.WatchUsers(ctx)

	go func() {
		for userEvent := range userEvents {
			err := s.producer.SendMessage(userEvent)
			if err != nil {
				log.Errorf("Error sending user event: %v", err)
			}
		}
	}()
}
