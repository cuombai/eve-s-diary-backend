package database

import (
    "context"
    "log"
    "os"
    "time"

    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

var (
    // DB is the global client instance
    DB *mongo.Client
    // dbName is the name of your Atlas database
    dbName = "evediary"
)

// Init initializes the MongoDB connection once
func Init() {
    uri := os.Getenv("MONGODB_URI")
    if uri == "" {
        log.Fatal("MONGODB_URI not set")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
    if err != nil {
        log.Fatal("Mongo connect error:", err)
    }

    if err := client.Ping(ctx, nil); err != nil {
        log.Fatal("MongoDB connection failed:", err)
    }

    log.Println("Successfully connected to MongoDB Atlas!")
    DB = client
}

// GetDatabase returns the main database
func GetDatabase() *mongo.Database {
    if DB == nil {
        log.Fatal("Database not initialized. Call database.Init() first.")
    }
    return DB.Database(dbName)
}

// GetCollection returns a specific collection
func GetCollection(collectionName string) *mongo.Collection {
    return GetDatabase().Collection(collectionName)
}
