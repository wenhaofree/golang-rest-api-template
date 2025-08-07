package api

import (
	"context"
	"golang-rest-api-template/pkg/cache"
	"golang-rest-api-template/pkg/database"
	"golang-rest-api-template/pkg/middleware"
	"time"

	docs "golang-rest-api-template/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"golang.org/x/time/rate"
)

func ContextMiddleware(bookRepository BookRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("appCtx", bookRepository)
		c.Next()
	}
}

func MongoStatusMiddleware(mongoCollection *mongo.Collection) gin.HandlerFunc {
	return func(c *gin.Context) {
		if mongoCollection != nil {
			c.Set("mongo_status", "connected")
		} else {
			c.Set("mongo_status", "disabled")
		}
		c.Next()
	}
}

func NewRouter(logger *zap.Logger, mongoCollection *mongo.Collection, db database.Database, redisClient cache.Cache, ctx *context.Context) *gin.Engine {
	bookRepository := NewBookRepository(db, redisClient, ctx)
	userRepository := NewUserRepository(db, redisClient, ctx)

	r := gin.Default()
	r.Use(ContextMiddleware(bookRepository))
	r.Use(MongoStatusMiddleware(mongoCollection)) // 设置MongoDB状态到上下文

	//r.Use(gin.Logger())
	r.Use(middleware.Logger(logger, mongoCollection)) // mongoCollection可能为nil，中间件会处理
	if gin.Mode() == gin.ReleaseMode {
		r.Use(middleware.Security())
		r.Use(middleware.Xss())
	}
	r.Use(middleware.Cors())
	r.Use(middleware.RateLimiter(rate.Every(1*time.Minute), 60)) // 60 requests per minute

	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := r.Group("/api/v1")
	{
		v1.GET("/", bookRepository.Healthcheck)
		v1.GET("/books", middleware.APIKeyAuth(), bookRepository.FindBooks)
		v1.POST("/books", middleware.APIKeyAuth(), middleware.JWTAuth(), bookRepository.CreateBook)
		v1.GET("/books/:id", middleware.APIKeyAuth(), bookRepository.FindBook)
		v1.PUT("/books/:id", middleware.APIKeyAuth(), bookRepository.UpdateBook)
		v1.DELETE("/books/:id", middleware.APIKeyAuth(), bookRepository.DeleteBook)

		// 用户管理路由
		v1.GET("/users", middleware.APIKeyAuth(), userRepository.FindUsers)

		// 认证路由
		v1.POST("/login", middleware.APIKeyAuth(), userRepository.LoginHandler)
		v1.POST("/register", middleware.APIKeyAuth(), userRepository.RegisterHandler)
		v1.POST("/auth/third-party", middleware.APIKeyAuth(), userRepository.ThirdPartyLoginHandler)

		// 用户个人资料路由（需要JWT认证）
		v1.GET("/profile", middleware.APIKeyAuth(), middleware.JWTAuth(), userRepository.GetUserProfile)
		v1.PUT("/profile", middleware.APIKeyAuth(), middleware.JWTAuth(), userRepository.UpdateUserProfile)
		v1.DELETE("/profile", middleware.APIKeyAuth(), middleware.JWTAuth(), userRepository.SoftDeleteUser)
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return r
}
