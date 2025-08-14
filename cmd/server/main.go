package main

import (
	"context"
	"golang-rest-api-template/pkg/api"
	"golang-rest-api-template/pkg/cache"
	"golang-rest-api-template/pkg/config"
	"golang-rest-api-template/pkg/database"
	"golang-rest-api-template/pkg/logging"
	"log"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/gin-gonic/gin"
)

// @title           Swagger Example API
// @version         1.0
// @description     This is a sample server celler server.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8001
// @BasePath  /api/v1

// @securityDefinitions.apikey JwtAuth
// @in header
// @name Authorization

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化各种服务（统一使用 cfg）
	redisClient := cache.NewRedisClient(cfg)
	db := database.NewDatabase(cfg)
	dbWrapper := &database.GormDatabase{DB: db}

	// 根据配置决定是否启用MongoDB
	var mongoCollection *mongo.Collection
	if cfg.MongoEnabled {
		mongoCollection = database.SetupMongoDB(cfg)
		if mongoCollection != nil {
			log.Println("MongoDB logging enabled")
		} else {
			log.Println("MongoDB connection failed, logging will only use structured logger")
		}
	} else {
		log.Println("MongoDB logging disabled by configuration")
	}

	ctx := context.Background()
	logger, _, err := logging.NewLoggerWithColor(cfg.LogLevel, gin.Mode() != gin.ReleaseMode, cfg.LogColorized)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()

	//gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.DebugMode)

	r := api.NewRouter(logger, mongoCollection, dbWrapper, redisClient, &ctx, cfg.RequestTimeoutMs, cfg.RateLimitRequests, cfg.RateLimitWindow)

	log.Printf("Server starting on port %s...", cfg.Port)
	log.Printf("MongoDB logging enabled: %v", cfg.MongoEnabled)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
