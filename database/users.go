package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	ADMIN_ROLE = "x000001"
	USER_ROLE  = "x000002"
)

const (
	ACTIVE_STATUS   = "x0000001"
	INACTIVE_STATUS = "x0000002"
	PENDING_STATUS  = "x0000003"
)

const UserCollectionName = "users"

type User struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Email     string             `json:"email,omitempty" bson:"email,omitempty"`
	Password  string             `json:"password,omitempty" bson:"password,omitempty"`
	Role      string             `json:"role,omitempty" bson:"role,omitempty"`
	Verified  bool               `json:"verified,omitempty" bson:"verified,omitempty"`
	Status    string             `json:"status,omitempty" bson:"status,omitempty"`
	CreatedAt time.Time          `bson:"created_at,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
}

type UserRepository interface {
	Insert(ctx context.Context, user User) (any, error)
	FindAll(ctx context.Context) ([]User, error)
	FindById(ctx context.Context, id string) (User, error)
	FindByEmail(ctx context.Context, email string) (User, error)
	Remove(ctx context.Context, id string) (bool, error)
	Update(ctx context.Context, id string, user bson.M) error
}

type MongoUserRepository struct {
	collection *mongo.Collection
}

func NewMongoUserRepository(collection *mongo.Collection) *MongoUserRepository {
	return &MongoUserRepository{collection}
}

func (r *MongoUserRepository) Insert(ctx context.Context, user User) (any, error) {
	result, err := r.collection.InsertOne(ctx, bson.M{
		"email":      user.Email,
		"password":   user.Password,
		"role":       user.Role,
		"verified":   user.Verified,
		"status":     PENDING_STATUS,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

func (r *MongoUserRepository) FindAll(ctx context.Context) ([]User, error) {
	var users []User
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *MongoUserRepository) FindById(ctx context.Context, id string) (User, error) {
	var user User
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return user, err
	}
	err = r.collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&user)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (r *MongoUserRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		return user, err
	}
	return user, nil
}

func (r *MongoUserRepository) Remove(ctx context.Context, id string) (bool, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}
	_, err = r.collection.DeleteOne(ctx, bson.M{"_id": objectId})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *MongoUserRepository) Update(ctx context.Context, id string, user bson.M) error {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectId}, bson.M{"$set": user})
	if err != nil {
		return err
	}
	return nil
}

type UserService struct {
	repository UserRepository
}

func NewUserService(repository UserRepository) *UserService {
	return &UserService{repository}
}

func (s *UserService) Insert(ctx context.Context, user User) (any, error) {
	return s.repository.Insert(ctx, user)
}

func (s *UserService) FindAll(ctx context.Context) ([]User, error) {
	return s.repository.FindAll(ctx)
}

func (s *UserService) FindByEmail(ctx context.Context, email string) (User, error) {
	return s.repository.FindByEmail(ctx, email)
}

func (s *UserService) Update(ctx context.Context, id string, user bson.M) error {
	return s.repository.Update(ctx, id, user)
}

func (s *UserService) FindById(ctx context.Context, id string) (User, error) {
	return s.repository.FindById(ctx, id)
}
