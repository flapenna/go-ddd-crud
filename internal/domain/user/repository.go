package domain

import (
	"context"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	DeleteUserById(ctx context.Context, id string) error
	ListUsers(ctx context.Context, request *ListUsersQueryRequest) (*ListUsersQueryResponse, error)
}
