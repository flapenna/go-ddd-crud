package mongodb

import "time"

type UserEntity struct {
	ID             string    `bson:"_id"`
	FirstName      string    `bson:"first_name"`
	LastName       string    `bson:"last_name"`
	Email          string    `bson:"email"`
	HashedPassword string    `bson:"hashed_password,omitempty"`
	Country        string    `bson:"country"`
	Nickname       string    `bson:"nickname"`
	CreatedAt      time.Time `bson:"created_at,omitempty"`
	UpdatedAt      time.Time `bson:"updated_at,omitempty"`
}
