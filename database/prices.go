package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const PriceCollectionName = "prices"

type Price struct {
	ID                     primitive.ObjectID `json:"_id" bson:"_id"`
	ProductID              primitive.ObjectID `json:"product_id,omitempty" bson:"product_id,omitempty"`
	Stock                  float64            `json:"stock,omitempty" bson:"stock,omitempty"`
	Sold                   float64            `json:"sold,omitempty" bson:"sold,omitempty"`
	HistoricalSold         float64            `json:"historical_sold,omitempty" bson:"historical_sold,omitempty"`
	LikedCount             float64            `json:"liked_count,omitempty" bson:"liked_count,omitempty"`
	CmtCount               float64            `json:"cmt_count,omitempty" bson:"cmt_count,omitempty"`
	Price                  float64            `json:"price,omitempty" bson:"price,omitempty"`
	PriceMin               float64            `json:"price_min,omitempty" bson:"price_min,omitempty"`
	PriceMax               float64            `json:"price_max,omitempty" bson:"price_max,omitempty"`
	PriceMinBeforeDiscount float64            `json:"price_min_before_discount,omitempty" bson:"price_min_before_discount,omitempty"`
	PriceMaxBeforeDiscount float64            `json:"price_max_before_discount,omitempty" bson:"price_max_before_discount,omitempty"`
	PriceBeforeDiscount    float64            `json:"price_before_discount,omitempty" bson:"price_before_discount,omitempty"`
	RawDiscount            float64            `json:"raw_discount,omitempty" bson:"raw_discount,omitempty"`
	CreatedAt              time.Time          `bson:"created_at,omitempty"`
	UpdatedAt              time.Time          `bson:"updated_at,omitempty"`
}

type PriceRepository interface {
	Insert(ctx context.Context, price Price) (Price, error)
	Update(ctx context.Context, id string, price Price) (Price, error)
	FindByProductID(ctx context.Context, productID primitive.ObjectID) ([]Price, error)
	FindOneByFilter(ctx context.Context, filter bson.M) (Price, error)
}

type MongoPriceRepository struct {
	collection *mongo.Collection
}

func NewMongoPriceRepository(collection *mongo.Collection) *MongoPriceRepository {
	return &MongoPriceRepository{collection}
}

func (r *MongoPriceRepository) Insert(ctx context.Context, price Price) (Price, error) {
	_, err := r.collection.InsertOne(ctx, bson.M{
		"product_id":                price.ProductID,
		"stock":                     price.Stock,
		"sold":                      price.Sold,
		"historical_sold":           price.HistoricalSold,
		"liked_count":               price.LikedCount,
		"cmt_count":                 price.CmtCount,
		"price":                     price.Price,
		"price_min":                 price.PriceMin,
		"price_max":                 price.PriceMax,
		"price_min_before_discount": price.PriceMinBeforeDiscount,
		"price_max_before_discount": price.PriceMaxBeforeDiscount,
		"price_before_discount":     price.PriceBeforeDiscount,
		"raw_discount":              price.RawDiscount,
		"created_at":                time.Now(),
		"updated_at":                time.Now(),
	})
	if err != nil {
		return Price{}, err
	}
	return price, nil
}

func (r *MongoPriceRepository) Update(ctx context.Context, id string, price Price) (Price, error) {
	idPrimitive, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return Price{}, err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{
		"_id": idPrimitive,
	}, bson.M{
		"$set": bson.M{
			"stock":                     price.Stock,
			"sold":                      price.Sold,
			"historical_sold":           price.HistoricalSold,
			"liked_count":               price.LikedCount,
			"cmt_count":                 price.CmtCount,
			"price":                     price.Price,
			"price_min":                 price.PriceMin,
			"price_max":                 price.PriceMax,
			"price_min_before_discount": price.PriceMinBeforeDiscount,
			"price_max_before_discount": price.PriceMaxBeforeDiscount,
			"price_before_discount":     price.PriceBeforeDiscount,
			"raw_discount":              price.RawDiscount,
			"updated_at":                time.Now(),
		},
	})
	if err != nil {
		return Price{}, err
	}
	return price, nil
}

func (r *MongoPriceRepository) FindByProductID(ctx context.Context, productID primitive.ObjectID) ([]Price, error) {
	var prices []Price
	cursor, err := r.collection.Find(ctx, bson.M{"product_id": productID})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &prices); err != nil {
		return nil, err
	}
	return prices, nil
}

func (r *MongoPriceRepository) FindOneByFilter(ctx context.Context, filter bson.M) (Price, error) {
	var price Price
	err := r.collection.FindOne(ctx, filter).Decode(&price)
	if err != nil {
		return Price{}, err
	}
	return price, nil
}

type PriceService struct {
	repo PriceRepository
}

func NewPriceService(repo PriceRepository) *PriceService {
	return &PriceService{repo}
}

func (s *PriceService) Insert(ctx context.Context, price Price) (Price, error) {
	return s.repo.Insert(ctx, price)
}

func (s *PriceService) Update(ctx context.Context, id string, price Price) (Price, error) {
	return s.repo.Update(ctx, id, price)
}

func (s *PriceService) FindByProductID(ctx context.Context, productID primitive.ObjectID) ([]Price, error) {
	return s.repo.FindByProductID(ctx, productID)
}

func (s *PriceService) FindOneByFilter(ctx context.Context, filter bson.M) (Price, error) {
	return s.repo.FindOneByFilter(ctx, filter)
}
