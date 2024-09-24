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
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"os"
	"testing"
	"time"
)

type UserWatcherTestSuite struct {
	suite.Suite
	mongoC     testcontainers.Container
	client     *mongo.Client
	collection *mongo.Collection
	watcher    *mongodb.UsersChangeStreamWatcher
	ctx        context.Context
	cancel     context.CancelFunc
}

func (suite *UserWatcherTestSuite) SetupSuite() {
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	ctx := context.Background()

	mongoC, err := tc.RunContainer(ctx,
		testcontainers.WithImage("mongo:7"),
		tc.WithReplicaSet(),
		testcontainers.WithWaitStrategy(wait.ForListeningPort("27017/tcp").WithStartupTimeout(60*time.Second)),
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

	err = client.Ping(ctx, readpref.Primary())
	suite.Require().NoError(err)

	collection := mongoDb.Collection("test")

	suite.mongoC = mongoC
	suite.client = client
	suite.collection = collection
	suite.watcher = mongodb.NewChangeStreamWatcher(collection)
	suite.ctx, suite.cancel = context.WithTimeout(ctx, 60*time.Second)
}

func (suite *UserWatcherTestSuite) TearDownSuite() {
	err := suite.client.Disconnect(suite.ctx)
	suite.Require().NoError(err)
	err = suite.mongoC.Terminate(suite.ctx)
	suite.Require().NoError(err)
	suite.cancel()
}

func (suite *UserWatcherTestSuite) TestWatchUsers() {
	events := suite.watcher.WatchUsers(suite.ctx)

	// Give some time for the watcher to initialize
	time.Sleep(5 * time.Second)

	// Insert a user
	user := &mongodb.UserEntity{
		ID:             uuid.NewString(),
		FirstName:      "Federico",
		LastName:       "La Penna",
		Email:          "flapenna@mail.com",
		Country:        "IT",
		Nickname:       "Pennino",
		HashedPassword: "password",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	_, err := suite.collection.InsertOne(suite.ctx, user)
	suite.Require().NoError(err)

	// Wait for the event to be captured
	select {
	case event := <-events:
		suite.Require().NotNil(event)
		suite.Nil(event.BeforeChange)
		suite.Equal(user.ID, event.AfterChange.ID)
		suite.Equal(user.FirstName, event.AfterChange.FirstName)
		suite.Equal(user.LastName, event.AfterChange.LastName)
		suite.Equal(user.Email, event.AfterChange.Email)
		suite.Equal(user.Country, event.AfterChange.Country)
		suite.Equal(user.Nickname, event.AfterChange.Nickname)
		suite.Equal(domain.OPERATION_CREATE, event.OperationType)
	case <-time.After(15 * time.Second):
		suite.Fail("Timed out waiting for change event")
	}

	// Update the user
	user.FirstName = "Updated Federico"
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": user}
	// Configure options to return the updated document
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	result := suite.collection.FindOneAndUpdate(suite.ctx, filter, update, opts)
	suite.Require().NoError(result.Err())

	// Wait for the event to be captured
	select {
	case event := <-events:
		suite.Require().NotNil(event)
		suite.Equal(user.ID, event.AfterChange.ID)
		suite.Equal(user.FirstName, event.AfterChange.FirstName)
		suite.Equal(user.LastName, event.AfterChange.LastName)
		suite.Equal(user.Email, event.AfterChange.Email)
		suite.Equal(user.Country, event.AfterChange.Country)
		suite.Equal(user.Nickname, event.AfterChange.Nickname)
		suite.Equal(domain.OPERATION_UPDATE, event.OperationType)
	case <-time.After(15 * time.Second):
		suite.Fail("Timed out waiting for change event")
	}

	// Delete the user
	_, err = suite.collection.DeleteOne(suite.ctx, bson.M{"_id": user.ID})
	suite.Require().NoError(err)

	// Wait for the event to be captured
	select {
	case event := <-events:
		suite.Require().NotNil(event)
		suite.Nil(event.AfterChange)
		suite.Equal(domain.OPERATION_DELETE, event.OperationType)
	case <-time.After(15 * time.Second):
		suite.Fail("Timed out waiting for change event")
	}
}

func TestUserWatcherTestSuite(t *testing.T) {
	suite.Run(t, new(UserWatcherTestSuite))
}
