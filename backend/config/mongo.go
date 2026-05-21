package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

var MongoDBName = "medieval_store"

func ConnectMongo() error {
	uri := os.Getenv("MONGO_URI")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)

	//Failed to connect to server
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v\n", err)
		return err
	}

	//Ping to verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("Failed to ping MongoDB: %v\n", err)
		return err
	}

	MongoClient = client
	log.Println("Successfully connected to MongoDB!")

	// Ensure (product_id, user_id) is unique on ratings — at most one rating per
	// user per product. CreateOne is idempotent: if the index already exists it
	// no-ops; if duplicate documents already exist, it errors here so we log it
	// rather than silently letting the invariant break.
	idxCtx, idxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer idxCancel()
	_, idxErr := client.Database(MongoDBName).Collection("ratings").Indexes().CreateOne(idxCtx, mongo.IndexModel{
		Keys:    bson.D{{Key: "product_id", Value: 1}, {Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_product_user"),
	})
	if idxErr != nil {
		log.Printf("Warning: failed to ensure unique ratings index: %v\n", idxErr)
	}

	// Ensure category names are unique so the catalog can't end up with duplicate
	// "Weapons" entries. Same idempotent CreateOne pattern as the ratings index above.
	catIdxCtx, catIdxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer catIdxCancel()
	_, catIdxErr := client.Database(MongoDBName).Collection("categories").Indexes().CreateOne(catIdxCtx, mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("uniq_category_name"),
	})
	if catIdxErr != nil {
		log.Printf("Warning: failed to ensure unique categories index: %v\n", catIdxErr)
	}

	return nil
}
