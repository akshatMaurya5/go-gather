package utils

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getConnection() (*mongo.Client, error) {
	// MongoDB connection URI
	uri := "mongodb+srv://akshat:akshat@sales-assist.en35s.mongodb.net/?retryWrites=true&w=majority&appName=sales-assist"

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

// Uncomment the main function if you want to run this file directly for testing
