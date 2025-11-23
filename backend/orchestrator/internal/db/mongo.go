package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// InitDB initializes MongoDB connection
func InitDB() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 👉 Change URI if you're using MongoDB Atlas
	uri := "mongodb://localhost:27017"
	clientOptions := options.Client().ApplyURI(uri)

	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal("❌ Failed to connect MongoDB:", err)
	}

	// Ping the database to confirm the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("❌ MongoDB connection failed:", err)
	}

	log.Println("✅ Connected to MongoDB successfully!")
}

// GetCollection returns a MongoDB collection
func GetCollection(name string) *mongo.Collection {
  return client.Database("auto_orchestrator").Collection(name)
}
