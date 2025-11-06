package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

func InitMongo() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal("MongoDB Connection Error:", err)
	}
	Client = client
	log.Println("✅ Connected to MongoDB")
}

func GetCollection(name string) *mongo.Collection {
	return Client.Database("autoflow_ai").Collection(name)
}