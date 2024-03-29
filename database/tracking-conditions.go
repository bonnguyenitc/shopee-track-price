package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	TrackingConditionCollectionName = "tracking_conditions"
	LESS_THAN                       = "less_than"
	GREATER_THAN                    = "greater_than"
	EQUAL                           = "equal"
)

type TrackingCondition struct {
	ID           string             `json:"id" bson:"_id,omitempty"`
	TrackingID   primitive.ObjectID `json:"tracking_id" bson:"tracking_id"`
	Tracking     bson.D             `json:"tracking" bson:"tracking"`
	Condition    string             `json:"condition" bson:"condition"`
	Price        int64              `json:"price" bson:"price"`
	UserID       primitive.ObjectID `json:"user_id" bson:"user_id"`
	User         bson.D             `json:"user" bson:"user"`
	Active       bool               `json:"active" bson:"active"`
	CreatedAt    time.Time          `bson:"created_at,omitempty"`
	UpdatedAt    time.Time          `bson:"updated_at,omitempty"`
	UserInfo     []User             `json:"user_info" bson:"user_info"`
	TrackingInfo []Tracking         `json:"tracking_info" bson:"tracking_info"`
}

type TrackingConditionRepository interface {
	Insert(ctx context.Context, tracking TrackingCondition) (TrackingCondition, error)
	FindOneByFilter(ctx context.Context, filter bson.M) (TrackingCondition, error)
	FindAllByFilter(ctx context.Context, filter bson.M) ([]TrackingCondition, error)
	Remove(ctx context.Context, id primitive.ObjectID) (bool, error)
	Update(ctx context.Context, id primitive.ObjectID, trackingCondition bson.M) (bool, error)
	RemoveByFilter(ctx context.Context, filter bson.M) (bool, error)
	FindAllByFilterWithUser(ctx context.Context, filter bson.M) ([]TrackingCondition, error)
}

type MongoTrackingConditionRepository struct {
	collection *mongo.Collection
}

func NewMongoTrackingConditionRepository(collection *mongo.Collection) *MongoTrackingConditionRepository {
	return &MongoTrackingConditionRepository{collection}
}

func (r *MongoTrackingConditionRepository) Insert(ctx context.Context, trackingCondition TrackingCondition) (TrackingCondition, error) {
	_, err := r.collection.InsertOne(ctx, bson.M{
		"tracking":   bson.D{{Key: "$ref", Value: TrackingCollectionName}, {Key: "$id", Value: trackingCondition.TrackingID}},
		"condition":  trackingCondition.Condition,
		"user":       bson.D{{Key: "$ref", Value: UserCollectionName}, {Key: "$id", Value: trackingCondition.UserID}},
		"active":     true,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	})
	if err != nil {
		return TrackingCondition{}, err
	}
	return trackingCondition, nil
}

func (r *MongoTrackingConditionRepository) FindOneByFilter(ctx context.Context, filter bson.M) (TrackingCondition, error) {
	var trackingCondition TrackingCondition
	err := r.collection.FindOne(ctx, filter).Decode(&trackingCondition)
	if err != nil {
		return TrackingCondition{}, err
	}
	return trackingCondition, nil
}

func (r *MongoTrackingConditionRepository) FindAllByFilter(ctx context.Context, filter bson.M) ([]TrackingCondition, error) {
	var trackingConditions []TrackingCondition
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return []TrackingCondition{}, err
	}
	if err = cursor.All(ctx, &trackingConditions); err != nil {
		return []TrackingCondition{}, err
	}
	return trackingConditions, nil
}

func (r *MongoTrackingConditionRepository) Remove(ctx context.Context, id primitive.ObjectID) (bool, error) {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": bson.M{"active": false,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *MongoTrackingConditionRepository) Update(ctx context.Context, id primitive.ObjectID, trackingCondition bson.M) (bool, error) {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{
		"$set": trackingCondition,
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *MongoTrackingConditionRepository) RemoveByFilter(ctx context.Context, filter bson.M) (bool, error) {
	_, err := r.collection.UpdateOne(ctx, filter, bson.M{
		"$set": bson.M{"active": false,
			"updated_at": time.Now()},
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *MongoTrackingConditionRepository) FindAllByFilterWithUser(ctx context.Context, filter bson.M) ([]TrackingCondition, error) {
	var trackingConditions []TrackingCondition
	// not return password
	projectStage := bson.D{{Key: "$project", Value: bson.D{{Key: "user_info.email", Value: 1},
		{Key: "tracking_info.shopee_url", Value: 1},
	}}}
	cursor, err := r.collection.Aggregate(ctx, mongo.Pipeline{
		bson.D{{Key: "$match", Value: filter}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         UserCollectionName,
			"localField":   "user.$id",
			"foreignField": "_id",
			"as":           "user_info",
		}}},
		bson.D{{Key: "$lookup", Value: bson.M{
			"from":         TrackingCollectionName,
			"localField":   "tracking.$id",
			"foreignField": "_id",
			"as":           "tracking_info",
		}}},
		projectStage,
	})
	if err != nil {
		return []TrackingCondition{}, err
	}
	if err = cursor.All(ctx, &trackingConditions); err != nil {
		return []TrackingCondition{}, err
	}
	return trackingConditions, nil
}

type TrackingConditionService struct {
	repo TrackingConditionRepository
}

func NewTrackingConditionService(repo TrackingConditionRepository) *TrackingConditionService {
	return &TrackingConditionService{repo}
}

func (s *TrackingConditionService) Insert(ctx context.Context, trackingCondition TrackingCondition) (TrackingCondition, error) {
	return s.repo.Insert(ctx, trackingCondition)
}

func (s *TrackingConditionService) FindOneByFilter(ctx context.Context, filter bson.M) (TrackingCondition, error) {
	return s.repo.FindOneByFilter(ctx, filter)
}

func (s *TrackingConditionService) FindAllByFilter(ctx context.Context, filter bson.M) ([]TrackingCondition, error) {
	return s.repo.FindAllByFilter(ctx, filter)
}

func (s *TrackingConditionService) Remove(ctx context.Context, id primitive.ObjectID) (bool, error) {
	return s.repo.Remove(ctx, id)
}

func (s *TrackingConditionService) Update(ctx context.Context, id primitive.ObjectID, trackingCondition bson.M) (bool, error) {
	return s.repo.Update(ctx, id, trackingCondition)
}

func (s *TrackingConditionService) RemoveByFilter(ctx context.Context, filter bson.M) (bool, error) {
	return s.repo.RemoveByFilter(ctx, filter)
}

func (s *TrackingConditionService) FindAllByFilterWithUser(ctx context.Context, filter bson.M) ([]TrackingCondition, error) {
	return s.repo.FindAllByFilterWithUser(ctx, filter)
}
