package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const ProductCollectionName = "products"

type Product struct {
	ID                     primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	IDShopee               int64              `json:"id_shopee,omitempty" bson:"id_shopee,omitempty"`
	ShopName               string             `json:"shop_name,omitempty" bson:"shop_name,omitempty"`
	ShopRating             float64            `json:"shop_rating,omitempty" bson:"shop_rating,omitempty"`
	Name                   string             `json:"name,omitempty" bson:"name,omitempty"`
	ShopID                 primitive.ObjectID `json:"shop_id,omitempty" bson:"shop_id,omitempty"`
	Stock                  int32              `json:"stock,omitempty" bson:"stock,omitempty"`
	Sold                   int32              `json:"sold,omitempty" bson:"sold,omitempty"`
	HistoricalSold         int32              `json:"historical_sold,omitempty" bson:"historical_sold,omitempty"`
	LikedCount             int32              `json:"liked_count,omitempty" bson:"liked_count,omitempty"`
	CmtCount               int32              `json:"cmt_count,omitempty" bson:"cmt_count,omitempty"`
	Price                  int64              `json:"price,omitempty" bson:"price,omitempty"`
	PriceMin               int64              `json:"price_min,omitempty" bson:"price_min,omitempty"`
	PriceMax               int64              `json:"price_max,omitempty" bson:"price_max,omitempty"`
	PriceMinBeforeDiscount int64              `json:"price_min_before_discount,omitempty" bson:"price_min_before_discount,omitempty"`
	PriceMaxBeforeDiscount int64              `json:"price_max_before_discount,omitempty" bson:"price_max_before_discount,omitempty"`
	PriceBeforeDiscount    int64              `json:"price_before_discount,omitempty" bson:"price_before_discount,omitempty"`
	RawDiscount            float32            `json:"raw_discount,omitempty" bson:"raw_discount,omitempty"`
	Images                 []string           `json:"images,omitempty" bson:"images,omitempty"`
	CreatedAt              time.Time          `bson:"created_at,omitempty"`
	UpdatedAt              time.Time          `bson:"updated_at,omitempty"`
}

type ProductRepository interface {
	Insert(ctx context.Context, product Product) (any, error)
	FindAll(ctx context.Context) ([]Product, error)
	FindByIdShopee(ctx context.Context, id int64) (Product, error)
	FindByName(ctx context.Context, name string) ([]Product, error)
	Remove(ctx context.Context, id primitive.ObjectID) (bool, error)
	Update(ctx context.Context, id string, product Product) (Product, error)
}

type MongoProductRepository struct {
	collection *mongo.Collection
}

func NewMongoProductRepository(collection *mongo.Collection) *MongoProductRepository {
	return &MongoProductRepository{collection}
}

func (r *MongoProductRepository) Insert(ctx context.Context, product Product) (any, error) {
	result, err := r.collection.InsertOne(ctx, bson.M{
		"id_shopee":  product.IDShopee,
		"name":       product.Name,
		"shop_id":    product.ShopID,
		"images":     product.Images,
		"created_at": time.Now(),
		"updated_at": time.Now(),
	})
	if err != nil {
		return nil, err
	} else {
		return result.InsertedID.(primitive.ObjectID), nil
	}
}

func (r *MongoProductRepository) FindAll(ctx context.Context) ([]Product, error) {
	cursor, err := r.collection.Find(ctx, nil)
	if err != nil {
		return nil, err
	} else {
		var products []Product
		for cursor.Next(ctx) {
			var product Product
			cursor.Decode(&product)
			products = append(products, product)
		}
		return products, nil
	}
}

func (r *MongoProductRepository) FindByIdShopee(ctx context.Context, id int64) (Product, error) {
	var result Product
	err := r.collection.FindOne(ctx, bson.M{
		"id_shopee": id,
	}).Decode(&result)
	if err != nil {
		return Product{}, err
	}
	return result, nil
}

func (r *MongoProductRepository) FindByName(ctx context.Context, name string) ([]Product, error) {
	cursor, err := r.collection.Find(ctx, nil)
	if err != nil {
		return nil, err
	} else {
		var products []Product
		for cursor.Next(ctx) {
			var product Product
			cursor.Decode(&product)
			products = append(products, product)
		}
		return products, nil
	}
}

func (r *MongoProductRepository) Remove(ctx context.Context, id primitive.ObjectID) (bool, error) {
	_, err := r.collection.DeleteOne(ctx, nil)
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func (r *MongoProductRepository) Update(ctx context.Context, id string, product Product) (Product, error) {
	userIDObj, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Product{}, err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{
		"_id": userIDObj,
	}, bson.M{
		"$set": bson.M{
			"name":       product.Name,
			"images":     product.Images,
			"updated_at": time.Now(),
		},
	})
	if err != nil {
		return Product{}, err
	} else {
		return product, nil
	}
}

type ProductService struct {
	repo ProductRepository
}

func NewProductService(repo ProductRepository) *ProductService {
	return &ProductService{repo}
}

func (s *ProductService) Insert(ctx context.Context, product Product) (any, error) {
	return s.repo.Insert(ctx, product)
}

func (s *ProductService) FindAll(ctx context.Context) ([]Product, error) {
	return s.repo.FindAll(ctx)
}

func (s *ProductService) FindByIdShopee(ctx context.Context, id int64) (Product, error) {
	return s.repo.FindByIdShopee(ctx, id)
}

func (s *ProductService) FindByName(ctx context.Context, name string) ([]Product, error) {
	return s.repo.FindByName(ctx, name)
}

func (s *ProductService) Remove(ctx context.Context, id primitive.ObjectID) (bool, error) {
	return s.repo.Remove(ctx, id)
}

func (s *ProductService) Update(ctx context.Context, id string, product Product) (Product, error) {
	return s.repo.Update(ctx, id, product)
}
