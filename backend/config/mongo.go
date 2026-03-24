package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client

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
	return nil
}
