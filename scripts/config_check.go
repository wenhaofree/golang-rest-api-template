package main

import (
	"fmt"
	"golang-rest-api-template/pkg/config"
	"os"
)

func main() {
	cfg := config.LoadConfig()
	
	mongoEnabledEnv := os.Getenv("MONGO_ENABLED")
	fmt.Printf("Environment MONGO_ENABLED: '%s'\n", mongoEnabledEnv)
	fmt.Printf("Parsed MongoEnabled: %t\n", cfg.MongoEnabled)
	
	if cfg.MongoEnabled {
		fmt.Printf("MongoDB URI: %s\n", cfg.MongoURI)
		fmt.Printf("MongoDB Database: %s\n", cfg.MongoDB)
		fmt.Printf("MongoDB Collection: %s\n", cfg.MongoLogCollection)
	} else {
		fmt.Println("MongoDB is disabled - logging will use structured logger only")
	}
}