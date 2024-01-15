package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const TrackingCollectionName = "trackings"

type Tracking struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	IDShopee  int64              `json:"id_shopee,omitempty" bson:"id_shopee,omitempty"`
	UserID    primitive.ObjectID `json:"user_id,omitempty" bson:"user_id,omitempty"`
	ShopeeUrl string             `json:"shopee_url,omitempty" bson:"shopee_url,omitempty"`
	Status    bool               `json:"status,omitempty" bson:"status,omitempty"`
	CreatedAt time.Time          `bson:"created_at,omitempty"`
	UpdatedAt time.Time          `bson:"updated_at,omitempty"`
}

type TrackingRepository interface {
	Insert(ctx context.Context, tracking Tracking) (any, error)
	FindByIDShopee(ctx context.Context, id int64, user_id primitive.ObjectID) (Tracking, error)
	FindByUserID(ctx context.Context, id primitive.ObjectID) ([]Tracking, error)
	Remove(ctx context.Context, id string) (bool, error)
	Update(ctx context.Context, id string, tracking Tracking) (Tracking, error)
	FindAll(ctx context.Context, limit int64, page int64) (DataWithPagination[Tracking], error)
}

type MongoTrackingRepository struct {
	collection *mongo.Collection
}

func NewMongoTrackingRepository(collection *mongo.Collection) *MongoTrackingRepository {
	return &MongoTrackingRepository{collection}
}

func (r *MongoTrackingRepository) Insert(ctx context.Context, tracking Tracking) (any, error) {
	result, err := r.collection.InsertOne(ctx, bson.M{
		"id_shopee":  tracking.IDShopee,
		"user_id":    tracking.UserID,
		"status":     tracking.Status,
		"shopee_url": tracking.ShopeeUrl,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	})
	if err != nil {
		return nil, err
	}
	return result.InsertedID, nil
}

func (r *MongoTrackingRepository) FindByIDShopee(ctx context.Context, id int64, user_id primitive.ObjectID) (Tracking, error) {
	var tracking Tracking
	err := r.collection.FindOne(ctx, bson.M{"id_shopee": id, "user_id": user_id}).Decode(&tracking)
	if err != nil {
		return Tracking{}, err
	}
	return tracking, nil
}

func (r *MongoTrackingRepository) FindByUserID(ctx context.Context, id primitive.ObjectID) ([]Tracking, error) {
	var trackings []Tracking
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": id})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &trackings); err != nil {
		return nil, err
	}
	return trackings, nil
}

func (r *MongoTrackingRepository) Remove(ctx context.Context, id string) (bool, error) {
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

func (r *MongoTrackingRepository) Update(ctx context.Context, id string, tracking Tracking) (Tracking, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Tracking{}, err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": objectId}, bson.M{
		"$set": bson.M{
			"status":     tracking.Status,
			"shopee_url": tracking.ShopeeUrl,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return Tracking{}, err
	}
	return tracking, nil
}

func (r *MongoTrackingRepository) FindAll(ctx context.Context, limit int64, page int64) (DataWithPagination[Tracking], error) {
	var trackings []Tracking
	skip := (page - 1) * limit

	total, err := r.collection.CountDocuments(ctx, bson.D{{}})
	if err != nil {
		return DataWithPagination[Tracking]{}, err
	}

	cursor, err := r.collection.Find(ctx, bson.M{}, &options.FindOptions{
		Limit: &limit,
		Skip:  &skip,
	})
	if err != nil {
		return DataWithPagination[Tracking]{}, err
	}
	if err = cursor.All(ctx, &trackings); err != nil {
		return DataWithPagination[Tracking]{}, err
	}
	return DataWithPagination[Tracking]{
		Data:        trackings,
		TotalItems:  int(total),
		TotalPages:  int(total/limit) + 1,
		CurrentPage: int(page),
		Limit:       int(limit),
	}, nil
}

type TrackingService struct {
	repository TrackingRepository
}

func NewTrackingService(repository TrackingRepository) *TrackingService {
	return &TrackingService{repository}
}

func (s *TrackingService) Insert(ctx context.Context, tracking Tracking) (any, error) {
	return s.repository.Insert(ctx, tracking)
}

func (s *TrackingService) FindByIDShopee(ctx context.Context, id int64, user_id primitive.ObjectID) (Tracking, error) {
	return s.repository.FindByIDShopee(ctx, id, user_id)
}

func (s *TrackingService) FindByUserID(ctx context.Context, id primitive.ObjectID) ([]Tracking, error) {
	return s.repository.FindByUserID(ctx, id)
}

func (s *TrackingService) Remove(ctx context.Context, id string) (bool, error) {
	return s.repository.Remove(ctx, id)
}

func (s *TrackingService) Update(ctx context.Context, id string, tracking Tracking) (Tracking, error) {
	return s.repository.Update(ctx, id, tracking)
}

func (s *TrackingService) FindAll(ctx context.Context, limit int64, page int64) (DataWithPagination[Tracking], error) {
	return s.repository.FindAll(ctx, limit, page)
}
