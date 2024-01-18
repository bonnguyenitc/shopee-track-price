package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoDBClient *mongo.Client
var MongoDB *mongo.Database

func NewMongoDB(dbName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI(os.Getenv("MONGO_URI"))
	MongoDBClient, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}
	MongoDB = MongoDBClient.Database(dbName)
	return nil
}

func IndexedForDocument(collection *mongo.Collection, index map[string]interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    index,
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	log.Println("Indexed for collection", collection.Name())
	return nil
}

func SetupIndexed() error {
	// setup index for users collection
	userCollection := MongoDB.Collection(UserCollectionName)
	err := IndexedForDocument(userCollection, map[string]interface{}{
		"email": 1,
	})
	if err != nil {
		return err
	}
	// setup index for products collection
	productCollection := MongoDB.Collection(ProductCollectionName)
	err = IndexedForDocument(productCollection, map[string]interface{}{
		"id_shopee": 1,
	})
	if err != nil {
		return err
	}
	// setup index for shops collection
	shopCollection := MongoDB.Collection(ShopCollectionName)
	err = IndexedForDocument(shopCollection, map[string]interface{}{
		"shop_id": 1,
	})
	if err != nil {
		return err
	}
	return nil
}

type DataWithPagination[T any] struct {
	Data        []T `json:"data"`
	TotalItems  int `json:"total_items"`
	TotalPages  int `json:"total_pages"`
	CurrentPage int `json:"current_page"`
	Limit       int `json:"limit"`
}
