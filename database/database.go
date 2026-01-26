package database

import (
	"context"
	"log"
	"time"
	"os"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectDB connects to MongoDB and returns a client instance
func ConnectDB() *mongo.Client {
	uri := os.Getenv("MONGODB_URI")
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel() // Important to prevent context leak

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Ping the primary
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Successfully connected to MongoDB!")
	return client
}

// Client instance
var DB *mongo.Client = ConnectDB()

// GetCollection gets a collection from the database
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database("evesdiary").Collection(collectionName)
	return collection
}

func GetDatabase() *mongo.Database {
    return DB.Database("evesdiary") // replace with your actual DB name
}
