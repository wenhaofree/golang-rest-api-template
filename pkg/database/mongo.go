package database

import (
	"context"
	"golang-rest-api-template/pkg/config"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// SetupMongoDB 根据配置设置MongoDB连接
func SetupMongoDB(cfg *config.Config) *mongo.Collection {
	if !cfg.MongoEnabled {
		log.Println("MongoDB is disabled by configuration")
		return nil
	}

	clientOptions := options.Client().ApplyURI(cfg.MongoURI)
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Printf("Failed to connect to MongoDB: %v", err)
		return nil
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Printf("Failed to ping MongoDB: %v", err)
		return nil
	}

	log.Println("MongoDB connected successfully")
	return client.Database(cfg.MongoDB).Collection(cfg.MongoLogCollection)
}
