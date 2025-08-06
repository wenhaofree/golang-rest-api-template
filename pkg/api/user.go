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
	ThirdPartyLoginHandler(c *gin.Context)
	FindUsers(c *gin.Context)
	GetUserProfile(c *gin.Context)
	UpdateUserProfile(c *gin.Context)
	SoftDeleteUser(c *gin.Context)
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
	var loginUser models.LoginUser
	var dbUser models.User

	// Get JSON body
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		response.BadRequest(c, "Bad Request")
		return
	}

	// Fetch the user from the database (check for non-deleted users)
	if err := r.DB.Where("email = ? AND deleted_at IS NULL", loginUser.Email).First(&dbUser).Error(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Unauthorized(c, "Invalid email or password")
		} else {
			response.InternalServerError(c, "Internal Server Error")
		}
		return
	}

	// Check if user is active
	if !dbUser.IsActive {
		response.Unauthorized(c, "Account is deactivated")
		return
	}

	// Verify password (only for email auth provider)
	if dbUser.AuthProvider == models.AuthProviderEmail {
		if dbUser.HashedPassword == nil {
			response.Unauthorized(c, "Invalid email or password")
			return
		}
		if err := bcrypt.CompareHashAndPassword([]byte(*dbUser.HashedPassword), []byte(loginUser.Password)); err != nil {
			response.Unauthorized(c, "Invalid email or password")
			return
		}
	} else {
		response.BadRequest(c, "Please use third-party login for this account")
		return
	}

	// Update last login time
	now := time.Now()
	dbUser.LastLogin = &now
	r.DB.Model(&dbUser).Updates(models.User{LastLogin: &now})

	// Generate JWT token
	token, err := auth.GenerateToken(dbUser.Email)
	if err != nil {
		response.InternalServerError(c, "Error generating token")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user":  dbUser.ToResponse(),
	})
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
	var registerUser models.RegisterUser

	if err := c.ShouldBindJSON(&registerUser); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := r.DB.Where("email = ? AND deleted_at IS NULL", registerUser.Email).First(&existingUser).Error(); err == nil {
		response.BadRequest(c, "User with this email already exists")
		return
	}

	// Hash the password
	hashedPassword, err := auth.HashPassword(registerUser.Password)
	if err != nil {
		response.InternalServerError(c, "Could not hash password")
		return
	}

	// Set default values if not provided
	platform := registerUser.Platform
	if platform == "" {
		platform = models.PlatformWeb
	}
	authProvider := registerUser.AuthProvider
	if authProvider == "" {
		authProvider = models.AuthProviderEmail
	}

	// Create new user
	newUser := models.User{
		Email:          registerUser.Email,
		FullName:       &registerUser.FullName,
		HashedPassword: &hashedPassword,
		Platform:       platform,
		AuthProvider:   authProvider,
		IsActive:       true,
		IsSuperuser:    false,
	}

	// Save the user to the database
	if err := r.DB.Create(&newUser).Error; err != nil {
		response.InternalServerError(c, fmt.Sprintf("Could not save user: %v", err))
		return
	}

	response.SuccessWithMessage(c, gin.H{
		"user": newUser.ToResponse(),
	}, "Registration successful")
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
	var users []models.User
	var userResponses []models.UserResponse

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
		err := json.Unmarshal([]byte(cachedUsers), &userResponses)
		if err != nil {
			response.InternalServerError(c, "Failed to unmarshal cached data")
			return
		}
		response.Success(c, userResponses)
		return
	}

	// If cache missed, fetch data from the database (exclude deleted users)
	r.DB.Where("deleted_at IS NULL").Offset(offset).Limit(limit).Find(&users)

	// Convert to response format
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	// Serialize users object and store it in Redis
	serializedUsers, err := json.Marshal(userResponses)
	if err != nil {
		response.InternalServerError(c, "Failed to marshal data")
		return
	}
	err = r.RedisClient.Set(*r.Ctx, cacheKey, serializedUsers, time.Minute).Err()
	if err != nil {
		response.InternalServerError(c, "Failed to set cache")
		return
	}

	response.Success(c, userResponses)
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
	var thirdPartyUser models.ThirdPartyLoginUser

	if err := c.ShouldBindJSON(&thirdPartyUser); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// TODO: 这里应该验证第三方平台的token
	// 例如：验证Apple ID token、Google token等
	// if !validateThirdPartyToken(thirdPartyUser.AuthProvider, thirdPartyUser.ProviderToken) {
	//     response.Unauthorized(c, "Invalid provider token")
	//     return
	// }

	var dbUser models.User

	// 尝试通过provider_user_id和auth_provider查找用户
	err := r.DB.Where("provider_user_id = ? AND auth_provider = ? AND deleted_at IS NULL",
		thirdPartyUser.ProviderUserID, thirdPartyUser.AuthProvider).First(&dbUser).Error()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在，创建新用户
			newUser := models.User{
				Email:          thirdPartyUser.Email,
				FullName:       &thirdPartyUser.FullName,
				Platform:       thirdPartyUser.Platform,
				AuthProvider:   thirdPartyUser.AuthProvider,
				ProviderUserID: &thirdPartyUser.ProviderUserID,
				AvatarURL:      &thirdPartyUser.AvatarURL,
				IsActive:       true,
				IsSuperuser:    false,
			}

			if err := r.DB.Create(&newUser).Error; err != nil {
				response.InternalServerError(c, fmt.Sprintf("Could not create user: %v", err))
				return
			}
			dbUser = newUser
		} else {
			response.InternalServerError(c, "Database error")
			return
		}
	}

	// 检查用户是否激活
	if !dbUser.IsActive {
		response.Unauthorized(c, "Account is deactivated")
		return
	}

	// 更新最后登录时间
	now := time.Now()
	dbUser.LastLogin = &now
	r.DB.Model(&dbUser).Updates(models.User{LastLogin: &now})

	// 生成JWT token
	token, err := auth.GenerateToken(dbUser.Email)
	if err != nil {
		response.InternalServerError(c, "Error generating token")
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user":  dbUser.ToResponse(),
	})
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
	// 从JWT中获取用户email
	email, exists := c.Get("username") // 注意：这里实际上是email
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var user models.User
	if err := r.DB.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "User not found")
		} else {
			response.InternalServerError(c, "Database error")
		}
		return
	}

	response.Success(c, user.ToResponse())
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
	// 从JWT中获取用户email
	email, exists := c.Get("username")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var updateUser models.UpdateUser
	if err := c.ShouldBindJSON(&updateUser); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var user models.User
	if err := r.DB.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "User not found")
		} else {
			response.InternalServerError(c, "Database error")
		}
		return
	}

	// 更新用户信息
	if err := r.DB.Model(&user).Updates(updateUser).Error; err != nil {
		response.InternalServerError(c, "Failed to update user")
		return
	}

	// 重新获取更新后的用户信息
	if err := r.DB.Where("email = ?", email).First(&user).Error(); err != nil {
		response.InternalServerError(c, "Failed to fetch updated user")
		return
	}

	response.Success(c, user.ToResponse())
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
	// 从JWT中获取用户email
	email, exists := c.Get("username")
	if !exists {
		response.Unauthorized(c, "User not authenticated")
		return
	}

	var user models.User
	if err := r.DB.Where("email = ? AND deleted_at IS NULL", email).First(&user).Error(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, "User not found")
		} else {
			response.InternalServerError(c, "Database error")
		}
		return
	}

	// 执行逻辑删除
	user.SoftDelete()
	if err := r.DB.Model(&user).Updates(models.User{DeletedAt: user.DeletedAt}).Error; err != nil {
		response.InternalServerError(c, "Failed to delete user")
		return
	}

	response.SuccessWithMessage(c, nil, "Account deleted successfully")
}
