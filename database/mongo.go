package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoDB *mongo.Database

func NewMongoDB(dbName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}
	MongoDB = client.Database(dbName)
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
