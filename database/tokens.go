package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	VerifyEmail         = "verify_email"
	ResetPassword       = "reset_password"
	TokenCollectionName = "tokens"
)

type Token struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Token     string             `json:"token,omitempty" bson:"token,omitempty"`
	Type      string             `json:"type,omitempty" bson:"type,omitempty"`
	UserId    primitive.ObjectID `json:"user_id,omitempty" bson:"user_id,omitempty"`
	User      bson.D             `json:"user,omitempty" bson:"user,omitempty"`
	ExpiredAt time.Time          `json:"expired_at,omitempty" bson:"expired_at,omitempty"`
	CreatedAt time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
}

type TokenRepository interface {
	Insert(ctx context.Context, token Token) (Token, error)
	FindOneByFilter(ctx context.Context, filter bson.M) (Token, error)
	Remove(ctx context.Context, filter bson.M) (bool, error)
}

type MongoTokenRepository struct {
	collection *mongo.Collection
}

func NewMongoTokenRepository(collection *mongo.Collection) *MongoTokenRepository {
	return &MongoTokenRepository{collection}
}

func (r *MongoTokenRepository) Insert(ctx context.Context, token Token) (Token, error) {
	_, err := r.collection.InsertOne(ctx, bson.M{
		"token": token.Token,
		"user": bson.D{
			{Key: "$ref", Value: UserCollectionName},
			{Key: "$id", Value: token.UserId},
		},
		"type":       token.Type,
		"expired_at": token.ExpiredAt,
		"created_at": time.Now(),
	})
	if err != nil {
		return Token{}, err
	}
	return token, nil
}

func (r *MongoTokenRepository) FindOneByFilter(ctx context.Context, filter bson.M) (Token, error) {
	var token Token
	err := r.collection.FindOne(ctx, filter).Decode(&token)
	if err != nil {
		return Token{}, err
	}
	return token, nil
}

func (r *MongoTokenRepository) Remove(ctx context.Context, filter bson.M) (bool, error) {
	_, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return false, err
	}
	return true, nil
}

type TokenService struct {
	repo TokenRepository
}

func NewTokenService(repo TokenRepository) *TokenService {
	return &TokenService{repo}
}

func (s *TokenService) Insert(ctx context.Context, token Token) (Token, error) {
	return s.repo.Insert(ctx, token)
}

func (s *TokenService) FindOneByFilter(ctx context.Context, filter bson.M) (Token, error) {
	return s.repo.FindOneByFilter(ctx, filter)
}

func (s *TokenService) Remove(ctx context.Context, filter bson.M) (bool, error) {
	return s.repo.Remove(ctx, filter)
}
