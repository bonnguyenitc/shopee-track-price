package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const ShopCollectionName = "shops"

type Shop struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	ShopID     int64              `json:"shop_id,omitempty" bson:"shop_id,omitempty"`
	Name       string             `json:"name,omitempty" bson:"name,omitempty"`
	ShopRating float64            `json:"shop_rating,omitempty" bson:"shop_rating,omitempty"`
	CreatedAt  time.Time          `bson:"created_at,omitempty"`
	UpdatedAt  time.Time          `bson:"updated_at,omitempty"`
}

type ShopRepository interface {
	Insert(ctx context.Context, shop Shop) (Shop, error)
	FindAll(ctx context.Context) ([]Shop, error)
	FindById(ctx context.Context, id string) (Shop, error)
	FindByShopShopeeId(ctx context.Context, id int64) (Shop, error)
	FindByName(ctx context.Context, name string) (Shop, error)
	Remove(ctx context.Context, id string) (bool, error)
	Update(ctx context.Context, id string, shop Shop) (Shop, error)
}

type MongoShopRepository struct {
	collection *mongo.Collection
}

func NewMongoShopRepository(collection *mongo.Collection) *MongoShopRepository {
	return &MongoShopRepository{collection}
}

func (r *MongoShopRepository) Insert(ctx context.Context, shop Shop) (Shop, error) {
	result, err := r.collection.InsertOne(ctx, bson.M{
		"shop_id":     shop.ShopID,
		"name":        shop.Name,
		"shop_rating": shop.ShopRating,
		"created_at":  time.Now(),
		"updated_at":  time.Now(),
	})
	if err != nil {
		return Shop{}, err
	}
	return Shop{
		ID: result.InsertedID.(primitive.ObjectID),
	}, nil
}

func (r *MongoShopRepository) FindAll(ctx context.Context) ([]Shop, error) {
	var shops []Shop
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &shops); err != nil {
		return nil, err
	}
	return shops, nil
}

func (r *MongoShopRepository) FindById(ctx context.Context, id string) (Shop, error) {
	var shop Shop
	err := r.collection.FindOne(ctx, id).Decode(&shop)
	if err != nil {
		return Shop{}, err
	}
	return shop, nil
}

func (r *MongoShopRepository) FindByName(ctx context.Context, name string) (Shop, error) {
	var shop Shop
	err := r.collection.FindOne(ctx, name).Decode(&shop)
	if err != nil {
		return Shop{}, err
	}
	return shop, nil
}

func (r *MongoShopRepository) Remove(ctx context.Context, id string) (bool, error) {
	_, err := r.collection.DeleteOne(ctx, nil)
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (r *MongoShopRepository) Update(ctx context.Context, id string, shop Shop) (Shop, error) {
	userIDObj, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Shop{}, err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": userIDObj}, bson.M{"$set": shop})
	if err != nil {
		return Shop{}, err
	}
	return shop, nil
}

func (r *MongoShopRepository) FindByShopShopeeId(ctx context.Context, id int64) (Shop, error) {
	var shop Shop
	err := r.collection.FindOne(ctx, bson.M{"shop_id": id}).Decode(&shop)
	if err != nil {
		return Shop{}, err
	}
	return shop, nil
}

type ShopService struct {
	repo ShopRepository
}

func NewShopService(repo ShopRepository) *ShopService {
	return &ShopService{repo}
}

func (s *ShopService) Insert(ctx context.Context, shop Shop) (Shop, error) {
	return s.repo.Insert(ctx, shop)
}

func (s *ShopService) FindAll(ctx context.Context) ([]Shop, error) {
	return s.repo.FindAll(ctx)
}

func (s *ShopService) FindByShopShopeeId(ctx context.Context, id int64) (Shop, error) {
	return s.repo.FindByShopShopeeId(ctx, id)
}
