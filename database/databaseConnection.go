package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBInstance() *mongo.Client {
	// More robust environment handling
	// Try to load .env, but don't fail if it doesn't exist
	isRailway := os.Getenv("RAILWAY_ENVIRONMENT") != "" || os.Getenv("RAILWAY_STATIC_URL") != ""

	if !isRailway {
		err := godotenv.Load(".env")
		if err != nil {
			log.Println("Warning: .env file not found, using system environment variables")
		}
	} else {
		log.Println("Running in Railway environment, using Railway environment variables")
	}

	// Get MongoDB URL from environment
	MongoDB := os.Getenv("MONGODB_URL")
	if MongoDB == "" {
		log.Fatal("Error: MONGODB_URL environment variable not set")
	}

	// Create MongoDB client
	client, err := mongo.NewClient(options.Client().ApplyURI(MongoDB))
	if err != nil {
		log.Fatal(err)
	}

	// Connect to MongoDB with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Verify connection
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB: ", err)
	}

	fmt.Println("Connected to MongoDB")
	return client
}

var Client *mongo.Client = DBInstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	// Get database name from environment or use default
	dbName := os.Getenv("MONGODB_DATABASE")
	if dbName == "" {
		dbName = "cluster0" // fallback to your default
	}

	var collection *mongo.Collection = client.Database(dbName).Collection(collectionName)
	return collection
}
