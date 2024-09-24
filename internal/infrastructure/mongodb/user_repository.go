package mongodb

import (
	"context"
	"errors"
	"fmt"
	"github.com/flapenna/go-ddd-crud/internal/domain/user"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(collection *mongo.Collection) *UserRepository {
	return &UserRepository{
		collection: collection,
	}
}

func (r *UserRepository) CreateUser(ctx context.Context, user *domain.User) error {
	_, err := r.collection.InsertOne(ctx, toEntity(user))
	if err != nil {
		return err
	}
	return nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user *domain.User) error {
	filter := bson.M{"_id": user.ID}
	update := bson.M{"$set": toEntity(user)}
	// Configure options to return the updated document
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedUser *UserEntity
	result := r.collection.FindOneAndUpdate(ctx, filter, update, opts)
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return domain.ErrUserNotFound
	}
	if err := result.Decode(&updatedUser); err != nil {
		return err
	}

	user.CreatedAt = updatedUser.CreatedAt

	return nil
}

func (r *UserRepository) DeleteUserById(ctx context.Context, id string) error {
	filter := bson.M{"_id": id}
	result := r.collection.FindOneAndDelete(ctx, filter)
	if errors.Is(result.Err(), mongo.ErrNoDocuments) {
		return domain.ErrUserNotFound
	}
	return result.Err()
}

func (r *UserRepository) ListUsers(ctx context.Context, request *domain.ListUsersQueryRequest) (*domain.ListUsersQueryResponse, error) {
	if request.PageSize == 0 {
		request.PageSize = 10
	}

	filter := bson.M{}

	// Add filters based on the request
	if request.Country != nil {
		filter["country"] = request.Country
	}
	if request.FirstName != nil {
		filter["first_name"] = request.FirstName
	}
	if request.LastName != nil {
		filter["last_name"] = request.LastName
	}
	if request.Nickname != nil {
		filter["nickname"] = request.Nickname
	}
	if request.Email != nil {
		filter["email"] = request.Email
	}

	// Get the total count of documents matching the filter
	totalCount, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %v", err)
	}

	// Pagination options
	findOptions := options.Find()
	findOptions.SetSkip(int64(request.Page * request.PageSize))
	findOptions.SetLimit(int64(request.PageSize))

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			log.Warnf("failed to close cursor: %v", err)
		}
	}()

	var users []*UserEntity
	if err = cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return &domain.ListUsersQueryResponse{
		Page:       request.Page,
		PageSize:   request.PageSize,
		TotalCount: uint32(totalCount),
		Results:    usersToDomain(users),
	}, nil
}

func usersToDomain(ul []*UserEntity) []*domain.User {
	users := make([]*domain.User, len(ul))
	for i, u := range ul {
		users[i] = userToDomain(u)
	}
	return users
}

func userToDomain(u *UserEntity) *domain.User {
	if u == nil {
		return nil
	}
	return &domain.User{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Country:   u.Country,
		Nickname:  u.Nickname,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func toEntity(user *domain.User) *UserEntity {
	return &UserEntity{
		ID:             user.ID,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		Email:          user.Email,
		HashedPassword: user.HashedPassword,
		Country:        user.Country,
		Nickname:       user.Nickname,
		CreatedAt:      user.CreatedAt,
		UpdatedAt:      user.UpdatedAt,
	}
}
