package domain

import (
	"time"
)

type User struct {
	ID             string
	FirstName      string
	LastName       string
	Email          string
	HashedPassword string
	Country        string
	Nickname       string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ListUsersQueryRequest struct {
	Page      uint32
	PageSize  uint32
	Country   *string
	FirstName *string
	LastName  *string
	Nickname  *string
	Email     *string
}

type ListUsersQueryResponse struct {
	Page       uint32
	PageSize   uint32
	TotalCount uint32
	Results    []*User
}

type UserEvent struct {
	Id            string
	UserId        string
	BeforeChange  *User
	AfterChange   *User
	OperationType OperationType
}

type OperationType int32

const (
	OPERATION_UNSPECIFIED OperationType = 0
	OPERATION_CREATE      OperationType = 1
	OPERATION_UPDATE      OperationType = 2
	OPERATION_DELETE      OperationType = 3
)
