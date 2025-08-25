package api

import (
	"context"
	"golang-rest-api-template/pkg/cache"
	"golang-rest-api-template/pkg/database"
	"golang-rest-api-template/pkg/handlers"
	"golang-rest-api-template/pkg/middleware"
	"golang-rest-api-template/pkg/repositories"
	"golang-rest-api-template/pkg/services"
	"time"

	docs "golang-rest-api-template/docs"

	"github.com/gin-gonic/gin"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"

	"golang.org/x/time/rate"
)

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

func NewRouter(logger *zap.Logger, mongoCollection *mongo.Collection, db database.Database, redisClient cache.Cache, ctx *context.Context, requestTimeoutMs int, rateLimitRequests int, rateLimitWindow int) *gin.Engine {
	// 创建 Repository 层
	bookRepo := repositories.NewBookRepository(db)
	userRepo := repositories.NewUserRepository(db)

	// 创建服务层
	bookService := services.NewBookService(bookRepo, redisClient)
	userService := services.NewUserService(userRepo, redisClient)

	// 创建Handler层
	bookHandler := handlers.NewBookHandler(bookService)
	userHandler := handlers.NewUserHandler(userService)

	r := gin.New()
	// 统一错误处理中间件要尽早注册
	r.Use(middleware.ErrorHandler())
	// 注入每请求 logger 与 request_id
	r.Use(middleware.RequestContextLogger(logger))
	r.Use(MongoStatusMiddleware(mongoCollection)) // 设置MongoDB状态到上下文

	// 全局请求超时（结合数据库 statement_timeout，更上游兜底）
	r.Use(middleware.RequestTimeout(requestTimeoutMs))

	// 性能监控中间件
	r.Use(middleware.PerformanceMonitor(logger))

	// 请求大小限制 (1MB)
	r.Use(middleware.RequestSizeLimit(1 << 20))

	//r.Use(gin.Logger())
	r.Use(middleware.Logger(logger, mongoCollection)) // mongoCollection可能为nil，中间件会处理
	if gin.Mode() == gin.ReleaseMode {
		r.Use(middleware.Security())
		r.Use(middleware.Xss())
	}
	r.Use(middleware.Cors())
	// 使用配置化的速率限制
	rateLimitInterval := time.Duration(rateLimitWindow) * time.Second
	r.Use(middleware.RateLimiter(rate.Every(rateLimitInterval), rateLimitRequests))

	docs.SwaggerInfo.BasePath = "/api/v1"
	v1 := r.Group("/api/v1")
	{
		// 健康检查
		v1.GET("/", bookHandler.Healthcheck)

		// 图书管理路由 - 使用新的Handler层
		books := v1.Group("/books")
		{
			books.GET("", middleware.APIKeyAuth(), bookHandler.ListBooks)
			books.POST("", middleware.APIKeyAuth(), middleware.JWTAuth(), bookHandler.CreateBook)
			books.GET("/:id", middleware.APIKeyAuth(), bookHandler.GetBook)
			books.PUT("/:id", middleware.APIKeyAuth(), bookHandler.UpdateBook)
			books.DELETE("/:id", middleware.APIKeyAuth(), bookHandler.DeleteBook)
		}

		// 用户管理路由 - 使用新的Handler层
		users := v1.Group("/users")
		{
			users.GET("", middleware.APIKeyAuth(), userHandler.ListUsers)
		}

		// 认证路由 - 使用新的Handler层
		auth := v1.Group("/auth")
		{
			auth.POST("/login", middleware.APIKeyAuth(), userHandler.Login)
			auth.POST("/register", middleware.APIKeyAuth(), userHandler.Register)
			auth.POST("/third-party", middleware.APIKeyAuth(), userHandler.ThirdPartyLogin)
		}

		// 用户个人资料路由 - 使用新的Handler层（需要JWT认证）
		profile := v1.Group("/profile")
		profile.Use(middleware.APIKeyAuth(), middleware.JWTAuth())
		{
			profile.GET("", userHandler.GetProfile)
			profile.PUT("", userHandler.UpdateProfile)
			profile.DELETE("", userHandler.SoftDelete)
		}
	}
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return r
}
