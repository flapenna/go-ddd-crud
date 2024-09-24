//go:build integration

package kafka_test

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	domain "github.com/flapenna/go-ddd-crud/internal/domain/user"
	kafkaClient "github.com/flapenna/go-ddd-crud/internal/infrastructure/kafka"
	pb "github.com/flapenna/go-ddd-crud/pkg/pb/user/v1"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/protobuf/proto"
	"testing"
	"time"
)

var (
	redpandaImage   = "docker.vectorized.io/vectorized/redpanda"
	redpandaVersion = "v21.8.1"
	testTopic       = "go-ddd-crud_user-event"
)

type UserProducerTestSuite struct {
	suite.Suite
	kafkaC   testcontainers.Container
	producer *kafka.Producer
	consumer *kafka.Consumer
	ctx      context.Context
	cancel   context.CancelFunc
}

func (suite *UserProducerTestSuite) SetupSuite() {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image: fmt.Sprintf("%s:%s", redpandaImage, redpandaVersion),
		ExposedPorts: []string{
			"9092:9092/tcp",
		},
		Cmd:        []string{"redpanda", "start"},
		WaitingFor: wait.ForLog("Successfully started Redpanda!"),
	}

	kafkaC, err := testcontainers.GenericContainer(context.Background(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	suite.Require().NoError(err)

	mPort, err := kafkaC.MappedPort(ctx, "9092")
	suite.Require().NoError(err)

	// Create a Kafka producer
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("localhost:%d", mPort.Int()),
	})
	suite.Require().NoError(err)

	// Create a Kafka consumer
	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": fmt.Sprintf("localhost:%s", mPort.Port()),
		"group.id":          "test-group",
		"auto.offset.reset": "earliest",
	})
	suite.Require().NoError(err)

	err = consumer.Subscribe(testTopic, nil)
	suite.Require().NoError(err)

	suite.kafkaC = kafkaC
	suite.producer = producer
	suite.consumer = consumer
	suite.ctx, suite.cancel = context.WithTimeout(context.Background(), 30*time.Second)
}

func (suite *UserProducerTestSuite) TearDownSuite() {
	suite.cancel()
	suite.producer.Close()
	suite.consumer.Close()
}

func (suite *UserProducerTestSuite) TestUserProducer_SendMessage() {
	// Define a test message
	userEvent := &domain.UserEvent{
		Id:            "1",
		UserId:        "user_123",
		BeforeChange:  &domain.User{},
		AfterChange:   &domain.User{},
		OperationType: domain.OPERATION_CREATE,
	}

	expectedUserEvent := &pb.UserEvent{
		Id:            "1",
		UserId:        "user_123",
		BeforeChange:  &pb.User{},
		AfterChange:   &pb.User{},
		OperationType: pb.OperationType_OPERATION_CREATE,
	}

	// Create a userProducer
	userProducer := kafkaClient.NewUserProducer(suite.producer, testTopic)

	// Send the message
	err := userProducer.SendMessage(userEvent)
	suite.Require().NoError(err)

	// Consume the message
	message, err := suite.consumer.ReadMessage(10 * time.Second)

	userEventReceived := &pb.UserEvent{}
	err = proto.Unmarshal(message.Value, userEventReceived)
	suite.Require().NoError(err)

	// assert the message received is the one sent
	suite.Equal(expectedUserEvent.Id, userEventReceived.Id)
	suite.Equal(expectedUserEvent.UserId, userEventReceived.UserId)
	suite.Equal(expectedUserEvent.OperationType, userEventReceived.OperationType)
}

func TestUserProducerTestSuite(t *testing.T) {
	suite.Run(t, new(UserProducerTestSuite))
}
