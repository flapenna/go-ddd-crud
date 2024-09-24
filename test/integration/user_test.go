package integration

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	pb "github.com/flapenna/go-ddd-crud/pkg/pb/user/v1"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	tc "github.com/testcontainers/testcontainers-go/modules/compose"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"os"
	"testing"
	"time"
)

type UserIntegrationTestSuite struct {
	suite.Suite
	compose     tc.ComposeStack
	grpcClient  pb.UserServiceClient
	consumer    *kafka.Consumer
	conn        *grpc.ClientConn
	ctx         context.Context
	cancel      context.CancelFunc
	createdUser *pb.User
	updatedUser *pb.User
}

func (suite *UserIntegrationTestSuite) SetupSuite() {
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	compose, err := tc.NewDockerCompose("../../docker-compose.test.yaml")
	suite.NoError(err, "NewDockerComposeAPI()")

	compose.WithOsEnv()
	ctx := context.Background()

	err = compose.Up(ctx, tc.Wait(true))
	suite.NoError(err, "compose.Up()")

	// Create a gRPC connection to the server
	conn, err := grpc.NewClient(
		fmt.Sprintf("0.0.0.0:%s", "9091"),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	suite.Require().NoError(err)

	grpcClient := pb.NewUserServiceClient(conn)

	// Create a Kafka consumer
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("localhost:%s", "9192"),
		"group.id":          "test-group",
		"auto.offset.reset": "earliest",
	})
	suite.Require().NoError(err)

	err = consumer.Subscribe("go-ddd-crud_user-event", nil)
	suite.Require().NoError(err)

	suite.grpcClient = grpcClient
	suite.consumer = consumer
	suite.compose = compose.WithOsEnv()
	suite.ctx, suite.cancel = context.WithTimeout(ctx, 30*time.Second)
}

func (suite *UserIntegrationTestSuite) TearDownSuite() {
	suite.cancel()
	suite.consumer.Close()
	err := suite.compose.Down(context.Background(), tc.RemoveOrphans(true), tc.RemoveImagesLocal)
	suite.NoError(err, "compose.Down()")
}

func (suite *UserIntegrationTestSuite) TestUserIntegration_a_CreateUser() {
	req := &pb.CreateUserRequest{
		FirstName: "Federico",
		LastName:  "La Penna",
		Email:     "flapenna@email.com",
		Password:  "password",
		Country:   "IT",
		Nickname:  "Pennino",
	}
	wantedRes := &pb.User{
		Id:        uuid.NewString(),
		FirstName: "Federico",
		LastName:  "La Penna",
		Email:     "flapenna@email.com",
		Country:   "IT",
		Nickname:  "Pennino",
		CreatedAt: timestamppb.New(time.Now().UTC()),
		UpdatedAt: timestamppb.New(time.Now().UTC()),
	}
	wantedEvent := &pb.UserEvent{
		Id:            uuid.NewString(),
		UserId:        wantedRes.Id,
		BeforeChange:  nil,
		AfterChange:   wantedRes,
		OperationType: pb.OperationType_OPERATION_CREATE,
	}

	resp, err := suite.grpcClient.CreateUser(suite.ctx, req)
	suite.Require().NoError(err)

	suite.Equal(wantedRes.FirstName, resp.FirstName)
	suite.Equal(wantedRes.LastName, resp.LastName)
	suite.Equal(wantedRes.Country, resp.Country)
	suite.Equal(wantedRes.Email, resp.Email)
	suite.Equal(wantedRes.Nickname, resp.Nickname)
	suite.WithinDuration(wantedRes.CreatedAt.AsTime(), resp.CreatedAt.AsTime(), 1*time.Second)
	suite.WithinDuration(wantedRes.UpdatedAt.AsTime(), resp.UpdatedAt.AsTime(), 1*time.Second)

	// Consume the message
	message, err := suite.consumer.ReadMessage(10 * time.Second)
	suite.Require().NoError(err)

	userEventReceived := &pb.UserEvent{}
	err = proto.Unmarshal(message.Value, userEventReceived)
	suite.Require().NoError(err)

	suite.Equal(wantedEvent.OperationType, userEventReceived.OperationType)
	suite.Equal(wantedEvent.BeforeChange, userEventReceived.BeforeChange)
	// assert user
	suite.Equal(wantedEvent.AfterChange.FirstName, userEventReceived.AfterChange.FirstName)
	suite.Equal(wantedEvent.AfterChange.LastName, userEventReceived.AfterChange.LastName)
	suite.Equal(wantedEvent.AfterChange.Country, userEventReceived.AfterChange.Country)
	suite.Equal(wantedEvent.AfterChange.Email, userEventReceived.AfterChange.Email)
	suite.Equal(wantedEvent.AfterChange.Nickname, userEventReceived.AfterChange.Nickname)

	// Store the created user for update tests
	suite.createdUser = resp
}

func (suite *UserIntegrationTestSuite) TestUserIntegration_b_UpdateUser() {
	suite.Require().NotNil(suite.createdUser, "User must be created first")

	req := &pb.UpdateUserRequest{
		Id:        suite.createdUser.Id,
		FirstName: "UpdatedName",
		LastName:  "UpdatedLastName",
		Email:     suite.createdUser.Email,
		Country:   suite.createdUser.Country,
		Nickname:  "UpdatedNick",
	}
	wantedRes := &pb.User{
		Id:        suite.createdUser.Id,
		FirstName: "UpdatedName",
		LastName:  "UpdatedLastName",
		Email:     suite.createdUser.Email,
		Country:   suite.createdUser.Country,
		Nickname:  "UpdatedNick",
		CreatedAt: suite.createdUser.CreatedAt,
		UpdatedAt: timestamppb.New(time.Now().UTC()),
	}
	wantedEvent := &pb.UserEvent{
		Id:            uuid.NewString(),
		UserId:        suite.createdUser.Id,
		BeforeChange:  suite.createdUser,
		AfterChange:   wantedRes,
		OperationType: pb.OperationType_OPERATION_UPDATE,
	}

	resp, err := suite.grpcClient.UpdateUser(suite.ctx, req)
	suite.Require().NoError(err)

	suite.Equal(wantedRes.FirstName, resp.FirstName)
	suite.Equal(wantedRes.LastName, resp.LastName)
	suite.Equal(wantedRes.Country, resp.Country)
	suite.Equal(wantedRes.Email, resp.Email)
	suite.Equal(wantedRes.Nickname, resp.Nickname)
	suite.WithinDuration(wantedRes.CreatedAt.AsTime(), resp.CreatedAt.AsTime(), 1*time.Second)
	suite.WithinDuration(wantedRes.UpdatedAt.AsTime(), resp.UpdatedAt.AsTime(), 1*time.Second)

	time.Sleep(5 * time.Second)
	// Consume the message
	message, err := suite.consumer.ReadMessage(10 * time.Second)
	suite.Require().NoError(err)

	userEventReceived := &pb.UserEvent{}
	err = proto.Unmarshal(message.Value, userEventReceived)
	suite.Require().NoError(err)

	suite.Equal(wantedEvent.OperationType, userEventReceived.OperationType)
	// assert user
	suite.Equal(wantedEvent.AfterChange.FirstName, userEventReceived.AfterChange.FirstName)
	suite.Equal(wantedEvent.AfterChange.LastName, userEventReceived.AfterChange.LastName)
	suite.Equal(wantedEvent.AfterChange.Country, userEventReceived.AfterChange.Country)
	suite.Equal(wantedEvent.AfterChange.Email, userEventReceived.AfterChange.Email)
	suite.Equal(wantedEvent.AfterChange.Nickname, userEventReceived.AfterChange.Nickname)

	suite.createdUser = resp
}

func (suite *UserIntegrationTestSuite) TestUserIntegration_c_ListUsers() {
	suite.Require().NotNil(suite.createdUser, "User must be created first")

	req := &pb.ListUsersRequest{}

	wantedRes := &pb.ListUsersResponse{
		Page:       0,
		PageSize:   10,
		TotalCount: 1,
		Results:    []*pb.User{suite.createdUser},
	}
	resp, err := suite.grpcClient.ListUsers(suite.ctx, req)
	suite.Require().NoError(err)

	suite.Equal(wantedRes.Page, resp.Page)
	suite.Equal(wantedRes.PageSize, resp.PageSize)
	suite.Equal(wantedRes.TotalCount, resp.TotalCount)

	suite.Equal(suite.createdUser.FirstName, resp.Results[0].FirstName)
	suite.Equal(suite.createdUser.LastName, resp.Results[0].LastName)
	suite.Equal(suite.createdUser.Country, resp.Results[0].Country)
	suite.Equal(suite.createdUser.Email, resp.Results[0].Email)
	suite.Equal(suite.createdUser.Nickname, resp.Results[0].Nickname)
	suite.Equal(suite.createdUser.CreatedAt, resp.Results[0].CreatedAt)
	suite.Equal(suite.createdUser.UpdatedAt, resp.Results[0].UpdatedAt)
}

func (suite *UserIntegrationTestSuite) TestUserIntegration_d_DeleteUser() {
	suite.Require().NotNil(suite.createdUser, "User must be created first")

	req := &pb.DeleteUserRequest{
		Id: suite.createdUser.Id,
	}
	wantedEvent := &pb.UserEvent{
		Id:            uuid.NewString(),
		UserId:        suite.createdUser.Id,
		BeforeChange:  suite.createdUser,
		AfterChange:   nil,
		OperationType: pb.OperationType_OPERATION_DELETE,
	}

	resp, err := suite.grpcClient.DeleteUser(suite.ctx, req)
	suite.Require().NoError(err)

	suite.IsType(&emptypb.Empty{}, resp)

	// Consume the message
	message, err := suite.consumer.ReadMessage(15 * time.Second)
	suite.Require().NoError(err)

	userEventReceived := &pb.UserEvent{}
	err = proto.Unmarshal(message.Value, userEventReceived)
	suite.Require().NoError(err)

	suite.Equal(wantedEvent.OperationType, userEventReceived.OperationType)
	// assert user deleted
	suite.Equal(wantedEvent.BeforeChange.FirstName, userEventReceived.BeforeChange.FirstName)
	suite.Equal(wantedEvent.BeforeChange.LastName, userEventReceived.BeforeChange.LastName)
	suite.Equal(wantedEvent.BeforeChange.Country, userEventReceived.BeforeChange.Country)
	suite.Equal(wantedEvent.BeforeChange.Email, userEventReceived.BeforeChange.Email)
	suite.Equal(wantedEvent.BeforeChange.Nickname, userEventReceived.BeforeChange.Nickname)
}

func TestUserIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(UserIntegrationTestSuite))
}
