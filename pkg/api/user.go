package api

import (
	"context"
	"golang-rest-api-template/pkg/apperrors"
	"golang-rest-api-template/pkg/cache"
	"golang-rest-api-template/pkg/database"
	"golang-rest-api-template/pkg/models"
	"golang-rest-api-template/pkg/response"
	"golang-rest-api-template/pkg/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserRepository interface {
	LoginHandler(c *gin.Context)
	RegisterHandler(c *gin.Context)
	ThirdPartyLoginHandler(c *gin.Context)
	FindUsers(c *gin.Context)
	GetUserProfile(c *gin.Context)
	UpdateUserProfile(c *gin.Context)
	SoftDeleteUser(c *gin.Context)
}

// userRepository holds shared resources like database and Redis client
type userRepository struct {
	service services.UserService
}

func NewUserRepository(db database.Database, redisClient cache.Cache, ctx *context.Context) *userRepository {
	return &userRepository{service: services.NewUserService(db, redisClient)}
}

// @BasePath /api/v1

// LoginHandler godoc
// @Summary Authenticate a user
// @Schemes
// @Description Authenticates a user using email and password, returns a JWT token if successful
// @Tags user
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param   user     body    models.LoginUser     true        "User login object"
// @Success 200 {string} string "JWT Token"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 500 {string} string "Internal Server Error"
// @Router /login [post]
func (r *userRepository) LoginHandler(c *gin.Context) {
	var req models.LoginUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.Wrap("BAD_REQUEST", "Invalid request", 400, err))
		return
	}
	token, user, err := r.service.Login(c.Request.Context(), req)
	if err == nil {
		response.Success(c, gin.H{"token": token, "user": user.ToResponse()})
		return
	}
	switch err {
	case services.ErrInvalid:
		c.Error(apperrors.ErrBadRequest)
	case services.ErrUnauthorized:
		c.Error(apperrors.ErrUnauthorized)
	default:
		c.Error(apperrors.ErrInternal)
	}
}

// RegisterHandler godoc
// @Summary Register a new user
// @Schemes http
// @Description Registers a new user with the given email and password
// @Tags user
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param   user     body    models.RegisterUser     true        "User registration object"
// @Success 201 {string} string	"Successfully registered"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /register [post]
func (r *userRepository) RegisterHandler(c *gin.Context) {
	var req models.RegisterUser
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.Wrap("BAD_REQUEST", "Invalid request", 400, err))
		return
	}
	user, err := r.service.Register(c.Request.Context(), req)
	if err == nil {
		response.SuccessWithMessage(c, gin.H{"user": user.ToResponse()}, "Registration successful")
		return
	}
	switch err {
	case services.ErrInvalid:
		c.Error(apperrors.ErrBadRequest)
	case services.ErrConflict:
		c.Error(apperrors.ErrConflict)
	default:
		c.Error(apperrors.ErrInternal)
	}
}

// FindUsers godoc
// @Summary Get all users with pagination
// @Description Get a list of all users with optional pagination (excludes deleted users)
// @Tags user
// @Security ApiKeyAuth
// @Produce json
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(10)
// @Success 200 {array} models.UserResponse "Successfully retrieved list of users"
// @Router /users [get]
func (r *userRepository) FindUsers(c *gin.Context) {
	offsetQuery := c.DefaultQuery("offset", "0")
	limitQuery := c.DefaultQuery("limit", "10")
	offset, err := strconv.Atoi(offsetQuery)
	if err != nil {
		response.BadRequest(c, "Invalid offset format")
		return
	}
	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		response.BadRequest(c, "Invalid limit format (max 100)")
		return
	}
	users, err := r.service.List(c.Request.Context(), offset, limit)
	if err != nil {
		if err == services.ErrInvalid {
			response.BadRequest(c, "Invalid pagination")
		} else {
			response.InternalServerError(c, "Failed to fetch users")
		}
		return
	}
	response.Success(c, users)
}

// ThirdPartyLoginHandler godoc
// @Summary Third-party login (Apple ID, Google, etc.)
// @Description Authenticates or registers a user using third-party providers
// @Tags user
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param   user     body    models.ThirdPartyLoginUser     true        "Third-party login object"
// @Success 200 {object} map[string]interface{} "JWT Token and user info"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /auth/third-party [post]
func (r *userRepository) ThirdPartyLoginHandler(c *gin.Context) {
	var req models.ThirdPartyLoginUser
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}
	token, user, err := r.service.ThirdPartyLogin(c.Request.Context(), req)
	if err == nil {
		response.Success(c, gin.H{"token": token, "user": user.ToResponse()})
		return
	}
	if err == services.ErrInvalid {
		c.Error(apperrors.ErrBadRequest)
	} else {
		response.InternalServerError(c, "Authentication service temporarily unavailable")
	}
}

// GetUserProfile godoc
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags user
// @Security ApiKeyAuth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.UserResponse "User profile"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "User not found"
// @Router /profile [get]
func (r *userRepository) GetUserProfile(c *gin.Context) {
	email, exists := c.Get("username")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	emailStr, ok := email.(string)
	if !ok || emailStr == "" {
		response.Unauthorized(c, "Invalid user information in token")
		return
	}
	resp, err := r.service.GetProfile(c.Request.Context(), emailStr)
	if err != nil {
		if err == services.ErrNotFound {
			response.NotFound(c, "User not found")
		} else {
			c.Error(apperrors.ErrInternal)
		}
		return
	}
	response.Success(c, resp)
}

// UpdateUserProfile godoc
// @Summary Update user profile
// @Description Update the profile of the authenticated user
// @Tags user
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept  json
// @Produce  json
// @Param   user     body    models.UpdateUser     true        "User update object"
// @Success 200 {object} models.UserResponse "Updated user profile"
// @Failure 400 {string} string "Bad Request"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "User not found"
// @Router /profile [put]
func (r *userRepository) UpdateUserProfile(c *gin.Context) {
	email, exists := c.Get("username")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	emailStr, ok := email.(string)
	if !ok || emailStr == "" {
		response.Unauthorized(c, "Invalid user information in token")
		return
	}
	var update models.UpdateUser
	if err := c.ShouldBindJSON(&update); err != nil {
		c.Error(apperrors.Wrap("BAD_REQUEST", "Invalid request", 400, err))
		return
	}
	resp, err := r.service.UpdateProfile(c.Request.Context(), emailStr, update)
	if err != nil {
		if err == services.ErrNotFound {
			response.NotFound(c, "User not found")
		} else {
			c.Error(apperrors.ErrInternal)
		}
		return
	}
	response.Success(c, resp)
}

// SoftDeleteUser godoc
// @Summary Soft delete user account
// @Description Soft delete the authenticated user's account (sets deleted_at timestamp)
// @Tags user
// @Security ApiKeyAuth
// @Security BearerAuth
// @Produce json
// @Success 200 {string} string "Account deleted successfully"
// @Failure 401 {string} string "Unauthorized"
// @Failure 404 {string} string "User not found"
// @Router /profile [delete]
func (r *userRepository) SoftDeleteUser(c *gin.Context) {
	email, exists := c.Get("username")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}
	emailStr, ok := email.(string)
	if !ok || emailStr == "" {
		response.Unauthorized(c, "Invalid user information in token")
		return
	}
	if err := r.service.SoftDelete(c.Request.Context(), emailStr); err != nil {
		if err == services.ErrNotFound {
			response.NotFound(c, "User not found")
		} else {
			c.Error(apperrors.ErrInternal)
		}
		return
	}
	response.SuccessWithMessage(c, nil, "Account deleted successfully")
}
