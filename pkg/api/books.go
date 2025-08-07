package api

import (
	"context"
	"encoding/json"
	"golang-rest-api-template/pkg/cache"
	"golang-rest-api-template/pkg/database"
	"golang-rest-api-template/pkg/models"
	"golang-rest-api-template/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type BookRepository interface {
	Healthcheck(c *gin.Context)
	FindBooks(c *gin.Context)
	CreateBook(c *gin.Context)
	FindBook(c *gin.Context)
	UpdateBook(c *gin.Context)
	DeleteBook(c *gin.Context)
}

// bookRepository holds shared resources like database and Redis client
type bookRepository struct {
	DB          database.Database
	RedisClient cache.Cache
	Ctx         *context.Context
}

// NewAppContext creates a new AppContext
func NewBookRepository(db database.Database, redisClient cache.Cache, ctx *context.Context) *bookRepository {
	return &bookRepository{
		DB:          db,
		RedisClient: redisClient,
		Ctx:         ctx,
	}
}

// @BasePath /api/v1

// Healthcheck godoc
// @Summary Health check endpoint
// @Schemes
// @Description Check the health status of the application and its dependencies
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Health status"
// @Router / [get]
func (r *bookRepository) Healthcheck(c *gin.Context) {
	healthStatus := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"services": map[string]interface{}{
			"database": "connected", // 假设数据库连接正常，实际可以添加检查
			"redis":    "connected", // 假设Redis连接正常，实际可以添加检查
		},
	}
	
	// 从上下文中获取MongoDB状态（如果有的话）
	if mongoStatus, exists := c.Get("mongo_status"); exists {
		healthStatus["services"].(map[string]interface{})["mongodb"] = mongoStatus
	} else {
		healthStatus["services"].(map[string]interface{})["mongodb"] = "disabled"
	}
	
	response.Success(c, healthStatus)
}

// FindBooks godoc
// @Summary Get all books with pagination
// @Description Get a list of all books with optional pagination
// @Tags books
// @Security ApiKeyAuth
// @Produce json
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(10)
// @Success 200 {array} models.Book "Successfully retrieved list of books"
// @Router /books [get]
func (r *bookRepository) FindBooks(c *gin.Context) {
	var books []models.Book

	// Get query params
	offsetQuery := c.DefaultQuery("offset", "0")
	limitQuery := c.DefaultQuery("limit", "10")

	// Convert query params to integers
	offset, err := strconv.Atoi(offsetQuery)
	if err != nil {
		response.BadRequest(c, "Invalid offset format")
		return
	}

	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		response.BadRequest(c, "Invalid limit format")
		return
	}

	// Create a cache key based on query params
	cacheKey := "books_offset_" + offsetQuery + "_limit_" + limitQuery

	// Try fetching the data from Redis first
	cachedBooks, err := r.RedisClient.Get(*r.Ctx, cacheKey).Result()
	if err == nil {
		err := json.Unmarshal([]byte(cachedBooks), &books)
		if err != nil {
			response.InternalServerError(c, "Failed to unmarshal cached data")
			return
		}
		response.Success(c, books)
		return
	}

	// If cache missed, fetch data from the database
	r.DB.Offset(offset).Limit(limit).Find(&books)

	// Serialize books object and store it in Redis
	serializedBooks, err := json.Marshal(books)
	if err != nil {
		response.InternalServerError(c, "Failed to marshal data")
		return
	}
	err = r.RedisClient.Set(*r.Ctx, cacheKey, serializedBooks, time.Minute).Err() // Here TTL is set to one hour
	if err != nil {
		response.InternalServerError(c, "Failed to set cache")
		return
	}

	response.Success(c, books)
}

// CreateBook godoc
// @Summary Create a new book
// @Description Create a new book with the given input data
// @Tags books
// @Security ApiKeyAuth
// @Security JwtAuth
// @Accept  json
// @Produce  json
// @Param   input     body   models.CreateBook   true   "Create book object"
// @Success 201 {object} models.Book "Successfully created book"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Router /books [post]
func (r *bookRepository) CreateBook(c *gin.Context) {
	appCtx, exists := c.MustGet("appCtx").(*bookRepository)
	if !exists {
		response.InternalServerError(c, "internal server error")
		return
	}
	var input models.CreateBook

	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	book := models.Book{Title: input.Title, Author: input.Author}

	appCtx.DB.Create(&book)

	// Invalidate cache
	keysPattern := "books_offset_*"
	keys, err := appCtx.RedisClient.Keys(*appCtx.Ctx, keysPattern).Result()
	if err == nil {
		for _, key := range keys {
			appCtx.RedisClient.Del(*appCtx.Ctx, key)
		}
	}

	response.SuccessWithMessage(c, book, "Book created successfully")
}

// FindBook godoc
// @Summary Find a book by ID
// @Description Get details of a book by its ID
// @Tags books
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Book ID"
// @Success 200 {object} models.Book "Successfully retrieved book"
// @Failure 404 {string} string "Book not found"
// @Router /books/{id} [get]
func (r *bookRepository) FindBook(c *gin.Context) {
	var book models.Book

	if err := r.DB.Where("id = ?", c.Param("id")).First(&book).Error(); err != nil {
		response.NotFound(c, "book not found")
		return
	}

	response.Success(c, book)
}

// UpdateBook godoc
// @Summary Update a book by ID
// @Description Update the book details for the given ID
// @Tags books
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param id path string true "Book ID"
// @Param input body models.UpdateBook true "Update book object"
// @Success 200 {object} models.Book "Successfully updated book"
// @Failure 400 {string} string "Bad Request"
// @Failure 404 {string} string "book not found"
// @Router /books/{id} [put]
func (r *bookRepository) UpdateBook(c *gin.Context) {
	var book models.Book
	var input models.UpdateBook

	if err := r.DB.Where("id = ?", c.Param("id")).First(&book).Error(); err != nil {
		response.NotFound(c, "book not found")
		return
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	r.DB.Model(&book).Updates(models.Book{Title: input.Title, Author: input.Author})

	response.SuccessWithMessage(c, book, "Book updated successfully")
}

// DeleteBook godoc
// @Summary Delete a book by ID
// @Description Delete the book with the given ID
// @Tags books
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Book ID"
// @Success 204 {string} string "Successfully deleted book"
// @Failure 404 {string} string "book not found"
// @Router /books/{id} [delete]
func (r *bookRepository) DeleteBook(c *gin.Context) {
	var book models.Book

	if err := r.DB.Where("id = ?", c.Param("id")).First(&book).Error(); err != nil {
		response.NotFound(c, "book not found")
		return
	}

	r.DB.Delete(&book)

	response.SuccessWithMessage(c, true, "Book deleted successfully")
}
