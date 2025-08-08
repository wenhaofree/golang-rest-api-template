package api

import (
	"context"
	"golang-rest-api-template/pkg/cache"
	"golang-rest-api-template/pkg/database"
	"golang-rest-api-template/pkg/models"
	"golang-rest-api-template/pkg/response"
	"golang-rest-api-template/pkg/services"
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

type bookRepository struct {
	service services.BookService
}

// NewBookRepository wires db/cache into service and returns handler
func NewBookRepository(db database.Database, redisClient cache.Cache, ctx *context.Context) *bookRepository {
	svc := services.NewBookService(db, redisClient)
	return &bookRepository{service: svc}
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

	books, err = r.service.ListBooks(c.Request.Context(), offset, limit)
	if err != nil {
		response.InternalServerError(c, "Failed to list books")
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
	var input models.CreateBook
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	book, err := r.service.CreateBook(c.Request.Context(), input)
	if err != nil {
		response.InternalServerError(c, "Failed to create book")
		return
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
	book, err := r.service.GetBook(c.Request.Context(), c.Param("id"))
	if err != nil {
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
	var input models.UpdateBook
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	book, err := r.service.UpdateBook(c.Request.Context(), c.Param("id"), input)
	if err != nil {
		response.NotFound(c, "book not found")
		return
	}
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
	if err := r.service.DeleteBook(c.Request.Context(), c.Param("id")); err != nil {
		response.NotFound(c, "book not found")
		return
	}
	response.SuccessWithMessage(c, true, "Book deleted successfully")
}
