package mongodb

import (
	"context"
	"fmt"
	"github.com/flapenna/go-ddd-crud/internal/domain/user"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UsersChangeStreamWatcher struct {
	collection *mongo.Collection
	userEvents chan *domain.UserEvent
}

func NewChangeStreamWatcher(collection *mongo.Collection) *UsersChangeStreamWatcher {
	return &UsersChangeStreamWatcher{
		collection: collection,
		userEvents: make(chan *domain.UserEvent),
	}
}

func (w *UsersChangeStreamWatcher) WatchUsers(ctx context.Context) <-chan *domain.UserEvent {
	go func() {
		defer close(w.userEvents)
		changeStream, err := w.openChangeStream(ctx)
		if err != nil {
			log.Errorf("Failed to open change stream: %v", err)
			return
		}
		defer changeStream.Close(ctx)

		log.Info("Started watching for user changes.")

		for {
			select {
			case <-ctx.Done():
				log.Info("Context canceled, stopping watch.")
				return
			default:
				err = w.processNextChange(ctx, changeStream)
				if err != nil {
					log.Errorf("Error processing next change: %v", err)
					return
				}
			}
		}
	}()

	return w.userEvents
}

func (w *UsersChangeStreamWatcher) openChangeStream(ctx context.Context) (*mongo.ChangeStream, error) {
	return w.collection.Watch(ctx, mongo.Pipeline{},
		options.ChangeStream().SetFullDocument(options.UpdateLookup).
			SetFullDocumentBeforeChange(options.WhenAvailable))
}

func (w *UsersChangeStreamWatcher) processNextChange(ctx context.Context, changeStream *mongo.ChangeStream) error {

	if changeStream.Next(ctx) {
		changeDoc := changeStream.Current
		if changeDoc == nil {
			return nil
		}

		log.Debugf("Change doc: %v", changeDoc)

		afterChange, beforeChange := w.decodeChangeDocuments(changeDoc)
		userID := w.determineUserID(beforeChange, afterChange)
		operationType := changeDoc.Lookup("operationType").StringValue()

		userEvent := &domain.UserEvent{
			Id:            uuid.New().String(),
			UserId:        userID,
			BeforeChange:  userToDomain(beforeChange),
			AfterChange:   userToDomain(afterChange),
			OperationType: operationToEnum(operationType),
		}

		select {
		case w.userEvents <- userEvent:
		case <-ctx.Done():
			return ctx.Err()
		}
	} else if err := changeStream.Err(); err != nil {
		return fmt.Errorf("change stream error: %v", err)
	}
	return nil
}

func (w *UsersChangeStreamWatcher) decodeChangeDocuments(changeDoc bson.Raw) (*UserEntity, *UserEntity) {
	var afterChange, beforeChange *UserEntity

	fullDocument := changeDoc.Lookup("fullDocument")
	if fullDocument.Type == bson.TypeEmbeddedDocument {
		var usr UserEntity
		if err := bson.Unmarshal(fullDocument.Value, &usr); err != nil {
			log.Warnf("Failed to decode user document: %v", err)
		} else {
			afterChange = &usr
			log.Debugf("User decoded successfully: %v", usr)
		}
	} else {
		log.Debug("No fullDocument in change stream.")
	}

	fullDocumentBeforeChange := changeDoc.Lookup("fullDocumentBeforeChange")
	if fullDocumentBeforeChange.Type == bson.TypeEmbeddedDocument {
		var previousUser UserEntity
		if err := bson.Unmarshal(fullDocumentBeforeChange.Value, &previousUser); err != nil {
			log.Warnf("Failed to decode previous user document: %v", err)
		} else {
			beforeChange = &previousUser
			log.Debugf("Previous user decoded successfully: %v", previousUser)
		}
	} else {
		log.Debug("No previous fullDocumentBeforeChange in change stream.")
	}

	return afterChange, beforeChange
}

func (w *UsersChangeStreamWatcher) determineUserID(beforeChange, afterChange *UserEntity) string {
	if beforeChange != nil {
		return beforeChange.ID
	}
	if afterChange != nil {
		return afterChange.ID
	}
	return ""
}

func operationToEnum(operationType string) domain.OperationType {
	switch operationType {
	case "insert":
		return domain.OPERATION_CREATE
	case "update":
		return domain.OPERATION_UPDATE
	case "delete":
		return domain.OPERATION_DELETE
	default:
		return domain.OPERATION_UNSPECIFIED
	}
}
