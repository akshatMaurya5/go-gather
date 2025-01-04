package db

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getConnection() (*mongo.Client, error) {
	// MongoDB connection URI

	err := godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	uri := os.Getenv("URI")

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	fmt.Println("Connected to the database")

	return client, nil
}
func TestDbConnection() error {
	client, err := getConnection()
	if err != nil {
		return err
	}

	// Send a ping to confirm a successful connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
	return nil
}

func GetCollection(collectionName string) (*mongo.Collection, error) {
	dbConnection, err := getConnection()
	if err != nil {
		return nil, err
	}

	err = godotenv.Load(".env")

	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	dbName := os.Getenv("DB_NAME")

	collection := dbConnection.Database(dbName).Collection(collectionName)
	return collection, nil
}

func GetAllCollections() ([]string, error) {
	dbConnection, err := getConnection()
	if err != nil {
		return nil, err
	}

	// Get the database reference
	database := dbConnection.Database("initial")

	// Call ListCollectionNames on the database
	collections, err := database.ListCollectionNames(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	return collections, nil
}
