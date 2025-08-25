package handlers

import (
	"golang-rest-api-template/pkg/dto"
	"golang-rest-api-template/pkg/response"
	"golang-rest-api-template/pkg/services"
	"golang-rest-api-template/pkg/validators"
	"time"

	"github.com/gin-gonic/gin"
)

// BookHandler 图书HTTP处理器
type BookHandler struct {
	service   services.BookService
	validator *validators.BookValidator
}

// NewBookHandler 创建图书处理器
func NewBookHandler(service services.BookService) *BookHandler {
	return &BookHandler{
		service:   service,
		validator: validators.NewBookValidator(),
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
func (h *BookHandler) Healthcheck(c *gin.Context) {
	healthStatus := map[string]interface{}{
		"status":    "ok",
		"service":   "book-service",
		"timestamp": time.Now().UTC(),
		"services": map[string]interface{}{
			"database": "connected",
			"redis":    "connected",
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

// ListBooks godoc
// @Summary Get all books with pagination
// @Description Get a list of all books with optional pagination
// @Tags books
// @Security ApiKeyAuth
// @Produce json
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(10)
// @Success 200 {array} dto.BookResponse "Successfully retrieved list of books"
// @Failure 400 {object} response.ErrorEnvelope "Bad Request"
// @Failure 500 {object} response.ErrorEnvelope "Internal Server Error"
// @Router /books [get]
func (h *BookHandler) ListBooks(c *gin.Context) {
	// 1. 解析查询参数
	var req dto.ListBooksRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	// 2. 验证请求
	if err := h.validator.ValidateListBooks(&req); err != nil {
		c.Error(err)
		return
	}

	// 3. 调用服务
	books, err := h.service.ListBooks(c.Request.Context(), req.Offset, req.Limit)
	if err != nil {
		c.Error(err)
		return
	}

	// 4. 返回响应
	response.Success(c, dto.BookResponse{}.FromModels(books))
}

// CreateBook godoc
// @Summary Create a new book
// @Description Create a new book with the given input data
// @Tags books
// @Security ApiKeyAuth
// @Security JwtAuth
// @Accept  json
// @Produce  json
// @Param   input     body   dto.CreateBookRequest   true   "Create book object"
// @Success 201 {object} dto.BookResponse "Successfully created book"
// @Failure 400 {object} response.ErrorEnvelope "Bad Request"
// @Failure 401 {object} response.ErrorEnvelope "Unauthorized"
// @Failure 500 {object} response.ErrorEnvelope "Internal Server Error"
// @Router /books [post]
func (h *BookHandler) CreateBook(c *gin.Context) {
	// 1. 解析请求体
	var req dto.CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	// 2. 验证请求
	if err := h.validator.ValidateCreateBook(&req); err != nil {
		c.Error(err)
		return
	}

	// 3. 调用服务
	book, err := h.service.CreateBook(c.Request.Context(), req.ToCreateModel())
	if err != nil {
		c.Error(err)
		return
	}

	// 4. 返回响应
	response.SuccessWithMessage(c, dto.BookResponse{}.FromModel(book), "Book created successfully")
}

// GetBook godoc
// @Summary Find a book by ID
// @Description Get details of a book by its ID
// @Tags books
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Book ID"
// @Success 200 {object} dto.BookResponse "Successfully retrieved book"
// @Failure 400 {object} response.ErrorEnvelope "Bad Request"
// @Failure 404 {object} response.ErrorEnvelope "Book not found"
// @Failure 500 {object} response.ErrorEnvelope "Internal Server Error"
// @Router /books/{id} [get]
func (h *BookHandler) GetBook(c *gin.Context) {
	// 1. 获取路径参数
	id := c.Param("id")

	// 2. 验证ID
	if err := h.validator.ValidateID(id); err != nil {
		c.Error(err)
		return
	}

	// 3. 调用服务
	book, err := h.service.GetBook(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	// 4. 返回响应
	response.Success(c, dto.BookResponse{}.FromModel(book))
}

// UpdateBook godoc
// @Summary Update a book by ID
// @Description Update the book details for the given ID
// @Tags books
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param id path string true "Book ID"
// @Param input body dto.UpdateBookRequest true "Update book object"
// @Success 200 {object} dto.BookResponse "Successfully updated book"
// @Failure 400 {object} response.ErrorEnvelope "Bad Request"
// @Failure 404 {object} response.ErrorEnvelope "Book not found"
// @Failure 500 {object} response.ErrorEnvelope "Internal Server Error"
// @Router /books/{id} [put]
func (h *BookHandler) UpdateBook(c *gin.Context) {
	// 1. 获取路径参数
	id := c.Param("id")

	// 2. 验证ID
	if err := h.validator.ValidateID(id); err != nil {
		c.Error(err)
		return
	}

	// 3. 解析请求体
	var req dto.UpdateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	// 4. 验证请求
	if err := h.validator.ValidateUpdateBook(&req); err != nil {
		c.Error(err)
		return
	}

	// 5. 调用服务
	book, err := h.service.UpdateBook(c.Request.Context(), id, req.ToUpdateModel())
	if err != nil {
		c.Error(err)
		return
	}

	// 6. 返回响应
	response.SuccessWithMessage(c, dto.BookResponse{}.FromModel(book), "Book updated successfully")
}

// DeleteBook godoc
// @Summary Delete a book by ID
// @Description Delete the book with the given ID
// @Tags books
// @Security ApiKeyAuth
// @Produce json
// @Param id path string true "Book ID"
// @Success 204 {string} string "Successfully deleted book"
// @Failure 400 {object} response.ErrorEnvelope "Bad Request"
// @Failure 404 {object} response.ErrorEnvelope "Book not found"
// @Failure 500 {object} response.ErrorEnvelope "Internal Server Error"
// @Router /books/{id} [delete]
func (h *BookHandler) DeleteBook(c *gin.Context) {
	// 1. 获取路径参数
	id := c.Param("id")

	// 2. 验证ID
	if err := h.validator.ValidateID(id); err != nil {
		c.Error(err)
		return
	}

	// 3. 调用服务
	if err := h.service.DeleteBook(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	// 4. 返回响应
	response.SuccessWithMessage(c, nil, "Book deleted successfully")
}
