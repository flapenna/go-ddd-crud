//go:build integration

package mongodb_test

import (
	"context"
	domain "github.com/flapenna/go-ddd-crud/internal/domain/user"
	"github.com/flapenna/go-ddd-crud/internal/infrastructure/mongodb"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	tc "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
	"testing"
	"time"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	mongoC     testcontainers.Container
	client     *mongo.Client
	collection *mongo.Collection
	repo       *mongodb.UserRepository
	ctx        context.Context
	cancel     context.CancelFunc
}

func (suite *UserRepositoryTestSuite) SetupSuite() {
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	ctx := context.Background()
	mongoC, err := tc.RunContainer(ctx,
		testcontainers.WithImage("mongo:7"),
		tc.WithReplicaSet(),
	)
	suite.Require().NoError(err)

	connStr, err := mongoC.ConnectionString(ctx)
	suite.Require().NoError(err)

	clientOpts := options.Client().ApplyURI(connStr).SetDirect(true)
	client, err := mongo.Connect(ctx, clientOpts)
	suite.Require().NoError(err)

	mongoDb := client.Database("testdb")
	collOpts := options.CreateCollection().
		SetChangeStreamPreAndPostImages(bson.M{"enabled": true})
	err = mongoDb.CreateCollection(ctx, "test", collOpts)
	suite.Require().NoError(err)

	collection := mongoDb.Collection("test")

	suite.mongoC = mongoC
	suite.client = client
	suite.collection = collection
	suite.repo = mongodb.NewUserRepository(collection)
	suite.ctx, suite.cancel = context.WithTimeout(ctx, 5*time.Second)
}

func (suite *UserRepositoryTestSuite) TearDownSuite() {
	suite.client.Disconnect(suite.ctx)
	suite.mongoC.Terminate(suite.ctx)
	suite.cancel()
}

func (suite *UserRepositoryTestSuite) SetupTest() {
	// Clean up the collection before each test
	suite.collection.Drop(suite.ctx)
}

func (suite *UserRepositoryTestSuite) TestUserRepository_CreateUser() {
	now := time.Now().UTC().Round(time.Millisecond)
	idIt := uuid.NewString()
	idUk := uuid.NewString()
	tests := []struct {
		name      string
		req       *domain.User
		wantedErr error
	}{
		{
			name: "insert new user IT",
			req: &domain.User{
				ID:             idIt,
				FirstName:      "Federico",
				LastName:       "La Penna",
				Email:          "flapenna@email.com",
				HashedPassword: "password",
				Country:        "IT",
				Nickname:       "Pennino",
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			wantedErr: nil,
		},
		{
			name: "insert new user UK",
			req: &domain.User{
				ID:             idUk,
				FirstName:      "John",
				LastName:       "Doe",
				Email:          "jdoe@email.com",
				HashedPassword: "password",
				Country:        "UK",
				Nickname:       "JDoe",
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			wantedErr: nil,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := suite.repo.CreateUser(suite.ctx, tt.req)
			if tt.wantedErr != nil {
				suite.Error(err)
				suite.Equal(tt.wantedErr, err)
			} else {
				suite.Require().NoError(err)

				// check the saved user by retrieving it from the DB
				var saved mongodb.UserEntity
				err = suite.collection.FindOne(context.Background(), bson.M{"_id": tt.req.ID}).Decode(&saved)
				suite.Require().NoError(err)
				suite.Equal(tt.req.ID, saved.ID)
				suite.Equal(tt.req.FirstName, saved.FirstName)
				suite.Equal(tt.req.LastName, saved.LastName)
				suite.Equal(tt.req.Nickname, saved.Nickname)
				suite.Equal(tt.req.Email, saved.Email)
				suite.Equal(tt.req.HashedPassword, saved.HashedPassword)
				suite.Equal(tt.req.Country, saved.Country)
				suite.WithinDuration(tt.req.CreatedAt, saved.CreatedAt, time.Millisecond)
				suite.WithinDuration(tt.req.UpdatedAt, saved.UpdatedAt, time.Millisecond)
			}
		})
	}
}

func (suite *UserRepositoryTestSuite) TestUserRepository_UpdateUser() {
	createdAt := time.Now().UTC().Add(-time.Hour)
	id := uuid.NewString()
	tests := []struct {
		name      string
		seed      *domain.User
		req       *domain.User
		wantedErr error
	}{
		{
			name: "update user",
			seed: &domain.User{
				ID:             id,
				FirstName:      "Federico",
				LastName:       "La Penna",
				Email:          "flapenna@email.com",
				HashedPassword: "password",
				Country:        "IT",
				Nickname:       "Pennino",
				CreatedAt:      createdAt,
				UpdatedAt:      createdAt,
			},
			req: &domain.User{
				ID:             id,
				FirstName:      "Federico Updated",
				LastName:       "La Penna Updated",
				Email:          "updated@email.com",
				HashedPassword: "updated_password",
				Country:        "UK",
				Nickname:       "Penninov2",
				CreatedAt:      createdAt,
				UpdatedAt:      time.Now().UTC(),
			},
			wantedErr: nil,
		},
		{
			name: "trying to update not existing user",
			seed: nil,
			req: &domain.User{
				ID:             uuid.NewString(),
				FirstName:      "John",
				LastName:       "Doe",
				Email:          "jdoe@email.com",
				HashedPassword: "password",
				Country:        "UK",
				Nickname:       "JDoe",
				CreatedAt:      createdAt,
				UpdatedAt:      createdAt,
			},
			wantedErr: domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.seed != nil {
				// seed user as pre-requisite
				err := suite.repo.CreateUser(suite.ctx, tt.seed)
				suite.Require().NoError(err)
			}

			err := suite.repo.UpdateUser(suite.ctx, tt.req)
			if tt.wantedErr != nil {
				suite.Error(err)
				suite.Equal(tt.wantedErr, err)
			} else {
				suite.Require().NoError(err)

				// check the saved user by retrieving it from the DB
				var updated mongodb.UserEntity
				err = suite.collection.FindOne(context.Background(), bson.M{"_id": tt.req.ID}).Decode(&updated)
				suite.Require().NoError(err)
				suite.Equal(tt.req.ID, updated.ID)
				suite.Equal(tt.req.FirstName, updated.FirstName)
				suite.Equal(tt.req.LastName, updated.LastName)
				suite.Equal(tt.req.Nickname, updated.Nickname)
				suite.Equal(tt.req.Email, updated.Email)
				suite.Equal(tt.req.HashedPassword, updated.HashedPassword)
				suite.Equal(tt.req.Country, updated.Country)
				suite.WithinDuration(tt.req.CreatedAt, updated.CreatedAt, time.Millisecond)
				suite.WithinDuration(tt.req.UpdatedAt, updated.UpdatedAt, time.Millisecond)
			}
		})
	}
}

func (suite *UserRepositoryTestSuite) TestUserRepository_DeleteUserByID() {
	id := uuid.NewString()
	now := time.Now().UTC()
	tests := []struct {
		name      string
		seed      *domain.User
		req       string
		wantedErr error
	}{
		{
			name: "delete user by id",
			seed: &domain.User{
				ID:             id,
				FirstName:      "Federico",
				LastName:       "La Penna",
				Email:          "flapenna@email.com",
				HashedPassword: "password",
				Country:        "IT",
				Nickname:       "Pennino",
				CreatedAt:      now,
				UpdatedAt:      now,
			},
			req:       id,
			wantedErr: nil,
		},
		{
			name:      "trying to delete not existing user",
			seed:      nil,
			req:       uuid.NewString(),
			wantedErr: domain.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			if tt.seed != nil {
				// seed user as pre-requisite
				err := suite.repo.CreateUser(suite.ctx, tt.seed)
				suite.Require().NoError(err)
			}

			err := suite.repo.DeleteUserById(suite.ctx, tt.req)
			if tt.wantedErr != nil {
				suite.Error(err)
				suite.Equal(tt.wantedErr, err)
			} else {
				suite.Require().NoError(err)

				// check the user has been deleted
				var updated mongodb.UserEntity
				err = suite.collection.FindOne(context.Background(), bson.M{"_id": tt.req}).Decode(&updated)
				suite.Require().Error(err)
				suite.Equal(mongo.ErrNoDocuments, err)
			}
		})
	}
}

func (suite *UserRepositoryTestSuite) TestUserRepository_ListUsers() {
	idIt := uuid.NewString()
	idUk := uuid.NewString()
	idDe := uuid.NewString()
	idEs := uuid.NewString()

	countryIt := "IT"
	firstNameIt := "Federico"
	lastNameIt := "La Penna"
	nicknameIt := "Pennino"
	emailIt := "flapenna@email.com"
	// round due to bson spec https://bsonspec.org/spec.html
	now := time.Now().UTC().Round(time.Millisecond)
	tests := []struct {
		name      string
		seed      []*domain.User
		req       *domain.ListUsersQueryRequest
		wantedRes *domain.ListUsersQueryResponse
		wantedErr error
	}{
		{
			name: "list users without filter and with default pagination",
			seed: []*domain.User{
				{
					ID:             idIt,
					FirstName:      "Federico",
					LastName:       "La Penna",
					Email:          "flapenna@email.com",
					HashedPassword: "password",
					Country:        "IT",
					Nickname:       "Pennino",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
				{
					ID:             idUk,
					FirstName:      "John",
					LastName:       "Doe",
					Email:          "jdoe@email.com",
					HashedPassword: "password",
					Country:        "UK",
					Nickname:       "Jdoe",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			},
			req: &domain.ListUsersQueryRequest{},
			wantedRes: &domain.ListUsersQueryResponse{
				Page:       0,
				PageSize:   10,
				TotalCount: 2,
				Results: []*domain.User{
					{
						ID:             idIt,
						FirstName:      "Federico",
						LastName:       "La Penna",
						Email:          "flapenna@email.com",
						HashedPassword: "",
						Country:        "IT",
						Nickname:       "Pennino",
						CreatedAt:      now,
						UpdatedAt:      now,
					},
					{
						ID:             idUk,
						FirstName:      "John",
						LastName:       "Doe",
						Email:          "jdoe@email.com",
						HashedPassword: "",
						Country:        "UK",
						Nickname:       "Jdoe",
						CreatedAt:      now,
						UpdatedAt:      now,
					},
				},
			},
			wantedErr: nil,
		},
		{
			name: "returns 1 user of second page with size 3",
			seed: []*domain.User{
				{
					ID:             idIt,
					FirstName:      "Federico",
					LastName:       "La Penna",
					Email:          "flapenna@email.com",
					HashedPassword: "password",
					Country:        "IT",
					Nickname:       "Pennino",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
				{
					ID:             idUk,
					FirstName:      "John",
					LastName:       "Doe",
					Email:          "jdoe@email.com",
					HashedPassword: "password",
					Country:        "UK",
					Nickname:       "Jdoe",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
				{
					ID:             idDe,
					FirstName:      "Federico",
					LastName:       "La Penna",
					Email:          "flapenna@email.com",
					HashedPassword: "password",
					Country:        "DE",
					Nickname:       "Pennino",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
				{
					ID:             idEs,
					FirstName:      "John",
					LastName:       "Doe",
					Email:          "jdoe@email.com",
					HashedPassword: "password",
					Country:        "ES",
					Nickname:       "Jdoe",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			},
			req: &domain.ListUsersQueryRequest{
				Page:     1,
				PageSize: 3,
			},
			wantedRes: &domain.ListUsersQueryResponse{
				Page:       1,
				PageSize:   3,
				TotalCount: 4,
				Results: []*domain.User{
					{
						ID:             idEs,
						FirstName:      "John",
						LastName:       "Doe",
						Email:          "jdoe@email.com",
						HashedPassword: "",
						Country:        "ES",
						Nickname:       "Jdoe",
						CreatedAt:      now,
						UpdatedAt:      now,
					},
				},
			},
			wantedErr: nil,
		},
		{
			name: "returns 1 user with all filters",
			seed: []*domain.User{
				{
					ID:             idIt,
					FirstName:      "Federico",
					LastName:       "La Penna",
					Email:          "flapenna@email.com",
					HashedPassword: "password",
					Country:        "IT",
					Nickname:       "Pennino",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
				{
					ID:             idUk,
					FirstName:      "John",
					LastName:       "Doe",
					Email:          "jdoe@email.com",
					HashedPassword: "password",
					Country:        "UK",
					Nickname:       "Jdoe",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
				{
					ID:             idDe,
					FirstName:      "Federico",
					LastName:       "La Penna",
					Email:          "flapenna@email.com",
					HashedPassword: "password",
					Country:        "DE",
					Nickname:       "Pennino",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
				{
					ID:             idEs,
					FirstName:      "John",
					LastName:       "Doe",
					Email:          "jdoe@email.com",
					HashedPassword: "password",
					Country:        "ES",
					Nickname:       "Jdoe",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			},
			req: &domain.ListUsersQueryRequest{
				Page:      0,
				PageSize:  10,
				Country:   &countryIt,
				FirstName: &firstNameIt,
				LastName:  &lastNameIt,
				Nickname:  &nicknameIt,
				Email:     &emailIt,
			},
			wantedRes: &domain.ListUsersQueryResponse{
				Page:       0,
				PageSize:   10,
				TotalCount: 1,
				Results: []*domain.User{
					{
						ID:             idIt,
						FirstName:      "Federico",
						LastName:       "La Penna",
						Email:          "flapenna@email.com",
						HashedPassword: "",
						Country:        "IT",
						Nickname:       "Pennino",
						CreatedAt:      now,
						UpdatedAt:      now,
					},
				},
			},
			wantedErr: nil,
		},
		{
			name: "returns empty result for page with no users",
			seed: []*domain.User{
				{
					ID:             idIt,
					FirstName:      "Federico",
					LastName:       "La Penna",
					Email:          "flapenna@email.com",
					HashedPassword: "password",
					Country:        "IT",
					Nickname:       "Pennino",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
				{
					ID:             idUk,
					FirstName:      "John",
					LastName:       "Doe",
					Email:          "jdoe@email.com",
					HashedPassword: "password",
					Country:        "UK",
					Nickname:       "Jdoe",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			},
			req: &domain.ListUsersQueryRequest{
				Page:     10,
				PageSize: 3,
			},
			wantedRes: &domain.ListUsersQueryResponse{
				Page:       10,
				PageSize:   3,
				TotalCount: 2,
				Results:    []*domain.User{},
			},
			wantedErr: nil,
		},
		{
			name: "country filter returns 1 user",
			seed: []*domain.User{
				{
					ID:             idIt,
					FirstName:      "Federico",
					LastName:       "La Penna",
					Email:          "flapenna@email.com",
					HashedPassword: "password",
					Country:        "IT",
					Nickname:       "Pennino",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
				{
					ID:             idUk,
					FirstName:      "John",
					LastName:       "Doe",
					Email:          "jdoe@email.com",
					HashedPassword: "password",
					Country:        "UK",
					Nickname:       "Jdoe",
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			},
			req: &domain.ListUsersQueryRequest{
				Country: &countryIt,
			},
			wantedRes: &domain.ListUsersQueryResponse{
				Page:       0,
				PageSize:   10,
				TotalCount: 1,
				Results: []*domain.User{
					{
						ID:             idIt,
						FirstName:      "Federico",
						LastName:       "La Penna",
						Email:          "flapenna@email.com",
						HashedPassword: "",
						Country:        "IT",
						Nickname:       "Pennino",
						CreatedAt:      now,
						UpdatedAt:      now,
					},
				},
			},
			wantedErr: nil,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			// Clean up the collection
			suite.collection.Drop(suite.ctx)

			if tt.seed != nil {
				// seed user as pre-requisite
				for _, u := range tt.seed {
					err := suite.repo.CreateUser(suite.ctx, u)
					suite.Require().NoError(err)
				}
			}

			res, err := suite.repo.ListUsers(suite.ctx, tt.req)
			if tt.wantedErr != nil {
				suite.Error(err)
				suite.Equal(tt.wantedErr, err)
			} else {
				suite.Require().NoError(err)
				suite.Equal(tt.wantedRes, res)
			}
		})
	}
}

func TestUserRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
