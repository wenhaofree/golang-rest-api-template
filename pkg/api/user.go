package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang-rest-api-template/pkg/auth"
	"golang-rest-api-template/pkg/cache"
	"golang-rest-api-template/pkg/database"
	"golang-rest-api-template/pkg/models"
	"golang-rest-api-template/pkg/response"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	LoginHandler(c *gin.Context)
	RegisterHandler(c *gin.Context)
	FindUsers(c *gin.Context)
}

// userRepository holds shared resources like database and Redis client
type userRepository struct {
	DB          database.Database
	RedisClient cache.Cache
	Ctx         *context.Context
}

func NewUserRepository(db database.Database, redisClient cache.Cache, ctx *context.Context) *userRepository {
	return &userRepository{
		DB:          db,
		RedisClient: redisClient,
		Ctx:         ctx,
	}
}

// @BasePath /api/v1

// LoginHandler godoc
// @Summary Authenticate a user
// @Schemes
// @Description Authenticates a user using username and password, returns a JWT token if successful
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
	var incomingUser models.User
	var dbUser models.User

	// Get JSON body
	if err := c.ShouldBindJSON(&incomingUser); err != nil {
		response.BadRequest(c, "Bad Request")
		return
	}

	// Fetch the user from the database
	if err := r.DB.Where("username = ?", incomingUser.Username).First(&dbUser).Error(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Unauthorized(c, "Invalid username or password")
		} else {
			response.InternalServerError(c, "Internal Server Error")
		}
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(incomingUser.Password)); err != nil {
		response.Unauthorized(c, "Invalid username or password")
		return
	}

	// Generate JWT token
	token, err := auth.GenerateToken(dbUser.Username)
	if err != nil {
		response.InternalServerError(c, "Error generating token")
		return
	}

	response.Success(c, gin.H{"token": token})
}

// RegisterHandler godoc
// @Summary Register a new user
// @Schemes http
// @Description Registers a new user with the given username and password
// @Tags user
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param   user     body    models.LoginUser     true        "User registration object"
// @Success 201 {string} string	"Successfully registered"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /register [post]
func (r *userRepository) RegisterHandler(c *gin.Context) {
	var user models.LoginUser

	if err := c.ShouldBindJSON(&user); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		response.InternalServerError(c, "Could not hash password")
		return
	}

	// Create new user
	newUser := models.User{Username: user.Username, Password: hashedPassword}

	// Save the user to the database
	if err := r.DB.Create(&newUser).Error; err != nil {
		response.InternalServerError(c, fmt.Sprintf("Could not save user: %v", err))
		return
	}

	response.SuccessWithMessage(c, gin.H{"id": newUser.ID, "username": newUser.Username}, "Registration successful")
}

// FindUsers godoc
// @Summary Get all users with pagination
// @Description Get a list of all users with optional pagination
// @Tags user
// @Security ApiKeyAuth
// @Produce json
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(10)
// @Success 200 {array} models.User "Successfully retrieved list of users"
// @Router /users [get]
func (r *userRepository) FindUsers(c *gin.Context) {
	var users []models.User

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
	cacheKey := "users_offset_" + offsetQuery + "_limit_" + limitQuery

	// Try fetching the data from Redis first
	cachedUsers, err := r.RedisClient.Get(*r.Ctx, cacheKey).Result()
	if err == nil {
		err := json.Unmarshal([]byte(cachedUsers), &users)
		if err != nil {
			response.InternalServerError(c, "Failed to unmarshal cached data")
			return
		}
		// Remove password from cached data before returning
		for i := range users {
			users[i].Password = ""
		}
		response.Success(c, users)
		return
	}

	// If cache missed, fetch data from the database
	r.DB.Offset(offset).Limit(limit).Find(&users)

	// Remove passwords from response for security
	for i := range users {
		users[i].Password = ""
	}

	// Serialize users object and store it in Redis
	serializedUsers, err := json.Marshal(users)
	if err != nil {
		response.InternalServerError(c, "Failed to marshal data")
		return
	}
	err = r.RedisClient.Set(*r.Ctx, cacheKey, serializedUsers, time.Minute).Err()
	if err != nil {
		response.InternalServerError(c, "Failed to set cache")
		return
	}

	response.Success(c, users)
}
