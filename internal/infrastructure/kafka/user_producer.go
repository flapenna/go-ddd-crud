package kafka

import (
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	domain "github.com/flapenna/go-ddd-crud/internal/domain/user"
	pb "github.com/flapenna/go-ddd-crud/pkg/pb/user/v1"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userProducer struct {
	broker *kafka.Producer
	topic  string
}

func NewUserProducer(broker *kafka.Producer, topic string) domain.UserProducer {
	return &userProducer{
		broker,
		topic,
	}
}

func (userProducer *userProducer) SendMessage(message *domain.UserEvent) error {
	value, err := proto.Marshal(userEventToProto(message))
	if err != nil {
		log.Error("unable to marshal message: ", err)
		return err
	}

	err = userProducer.broker.Produce(&kafka.Message{
		Key:            []byte(message.UserId),
		TopicPartition: kafka.TopicPartition{Topic: &userProducer.topic, Partition: kafka.PartitionAny},
		Value:          value,
	}, nil)
	if err != nil {
		log.Error("unable to enqueue message ", message)
		return err
	}

	go func() {
		for e := range userProducer.broker.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Warnf("Failed to deliver message: %v\n", ev.TopicPartition.Error)
				} else {
					log.Infof("Successfully produced record to topic %s partition [%d] @ offset %v\n",
						*ev.TopicPartition.Topic, ev.TopicPartition.Partition, ev.TopicPartition.Offset)
				}
			}
		}
	}()

	return nil
}

func userEventToProto(event *domain.UserEvent) *pb.UserEvent {
	return &pb.UserEvent{
		Id:            event.Id,
		UserId:        event.UserId,
		BeforeChange:  userToProto(event.BeforeChange),
		AfterChange:   userToProto(event.AfterChange),
		OperationType: operationTypeToProto(event.OperationType),
	}
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

func operationTypeToProto(operationType domain.OperationType) pb.OperationType {
	switch operationType {
	case domain.OPERATION_CREATE:
		return pb.OperationType_OPERATION_CREATE
	case domain.OPERATION_UPDATE:
		return pb.OperationType_OPERATION_UPDATE
	case domain.OPERATION_DELETE:
		return pb.OperationType_OPERATION_DELETE
	case domain.OPERATION_UNSPECIFIED:
		return pb.OperationType_OPERATION_UNSPECIFIED
	default:
		return pb.OperationType_OPERATION_UNSPECIFIED
	}
}
