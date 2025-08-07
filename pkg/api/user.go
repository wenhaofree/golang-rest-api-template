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
	"github.com/google/uuid"
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
	startTime := time.Now()
	var loginUser models.User

	// 1. JSON解析阶段
	parseStart := time.Now()
	var loginRequest models.LoginUser
	if err := c.ShouldBindJSON(&loginRequest); err != nil {
		fmt.Printf("JSON parsing took: %v\n", time.Since(parseStart))
		response.BadRequest(c, "Invalid request format")
		return
	}
	fmt.Printf("JSON parsing took: %v\n", time.Since(parseStart))

	// 基本输入验证
	if loginRequest.Email == "" || loginRequest.Password == "" {
		response.BadRequest(c, "Email and password are required")
		return
	}

	// 2. 缓存查询阶段 - 优化版本
	cacheStart := time.Now()
	cacheKey := fmt.Sprintf("user_login:%s", loginRequest.Email)

	// 使用异步缓存查询，设置超时
	ctx, cancel := context.WithTimeout(*r.Ctx, 10*time.Millisecond)
	defer cancel()

	cachedUser, cacheErr := r.RedisClient.Get(ctx, cacheKey).Result()
	fmt.Printf("Cache lookup took: %v\n", time.Since(cacheStart))

	if cacheErr == nil {
		// 缓存命中，反序列化用户数据
		deserializeStart := time.Now()
		if err := json.Unmarshal([]byte(cachedUser), &loginUser); err == nil {
			fmt.Printf("Cache deserialization took: %v\n", time.Since(deserializeStart))

			// 验证缓存的用户数据是否仍然有效
			validateStart := time.Now()
			if r.validateCachedUser(&loginUser, &loginRequest) {
				fmt.Printf("Cache validation took: %v\n", time.Since(validateStart))
				fmt.Printf("Total login (cache hit) took: %v\n", time.Since(startTime))
				r.handleSuccessfulLogin(c, &loginUser, true) // true表示来自缓存
				return
			}
			fmt.Printf("Cache validation took: %v\n", time.Since(validateStart))
		}
		// 缓存数据无效，异步删除缓存避免阻塞
		go r.RedisClient.Del(*r.Ctx, cacheKey)
	}

	// 3. 数据库查询阶段
	dbStart := time.Now()
	if err := r.DB.Where("email = ? AND deleted_at IS NULL AND is_active = ?", loginRequest.Email, true).
		First(&loginUser).Error(); err != nil {
		fmt.Printf("Database query took: %v\n", time.Since(dbStart))

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 4. 防时序攻击阶段
			timingStart := time.Now()
			bcrypt.CompareHashAndPassword([]byte(auth.GetDummyHash()), []byte(loginRequest.Password))
			fmt.Printf("Timing attack prevention took: %v\n", time.Since(timingStart))
			fmt.Printf("Total login (user not found) took: %v\n", time.Since(startTime))
			response.Unauthorized(c, "Invalid email or password")
		} else {
			fmt.Printf("Total login (db error) took: %v\n", time.Since(startTime))
			response.InternalServerError(c, "Authentication service temporarily unavailable")
		}
		return
	}
	fmt.Printf("Database query took: %v\n", time.Since(dbStart))

	// 5. 密码验证阶段
	validateStart := time.Now()
	if r.validateUserCredentials(&loginUser, &loginRequest) {
		fmt.Printf("Password validation took: %v\n", time.Since(validateStart))

		// 6. 缓存更新阶段 - 异步优化
		cacheUpdateStart := time.Now()
		go func() {
			if userBytes, err := json.Marshal(loginUser); err == nil {
				r.RedisClient.Set(*r.Ctx, cacheKey, userBytes, 5*time.Minute)
			}
		}()
		fmt.Printf("Cache update took: %v\n", time.Since(cacheUpdateStart))

		fmt.Printf("Total login (success) took: %v\n", time.Since(startTime))
		r.handleSuccessfulLogin(c, &loginUser, false) // false表示来自数据库
	} else {
		fmt.Printf("Password validation took: %v\n", time.Since(validateStart))
		fmt.Printf("Total login (invalid password) took: %v\n", time.Since(startTime))
		response.Unauthorized(c, "Invalid email or password")
	}
}

// validateCachedUser 验证缓存的用户数据
func (r *userRepository) validateCachedUser(dbUser *models.User, loginUser *models.LoginUser) bool {
	// 检查用户是否仍然激活且未被删除
	if !dbUser.IsActive || dbUser.IsDeleted() {
		return false
	}

	// 验证认证提供商
	if dbUser.AuthProvider != models.AuthProviderEmail {
		return false
	}

	// 验证密码
	if dbUser.HashedPassword == nil {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(*dbUser.HashedPassword), []byte(loginUser.Password)) == nil
}

// validateUserCredentials 验证用户凭据
func (r *userRepository) validateUserCredentials(dbUser *models.User, loginUser *models.LoginUser) bool {
	// 验证认证提供商
	if dbUser.AuthProvider != models.AuthProviderEmail {
		return false
	}

	// 验证密码
	if dbUser.HashedPassword == nil {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(*dbUser.HashedPassword), []byte(loginUser.Password)) == nil
}

// handleSuccessfulLogin 处理成功登录的逻辑
func (r *userRepository) handleSuccessfulLogin(c *gin.Context, dbUser *models.User, fromCache bool) {
	// 生成JWT token（这个操作比较快）
	token, err := auth.GenerateToken(dbUser.Email)
	if err != nil {
		response.InternalServerError(c, "Authentication service error")
		return
	}

	// 异步更新最后登录时间，避免阻塞响应
	go r.updateLastLoginAsync(dbUser.ID, fromCache)

	// 立即返回响应
	response.Success(c, gin.H{
		"token": token,
		"user":  dbUser.ToResponse(),
	})
}

// updateLastLoginAsync 异步更新最后登录时间
func (r *userRepository) updateLastLoginAsync(userID interface{}, fromCache bool) {
	now := time.Now()

	// 如果是从缓存获取的用户，检查是否需要更新（避免频繁更新）
	if fromCache {
		// 可以添加更智能的更新策略，比如只有距离上次登录超过一定时间才更新
		// 这里简化处理，每次都更新
	}

	// 只更新last_login字段，减少数据库负载
	if db, ok := r.DB.(*database.GormDatabase); ok {
		db.Model(&models.User{}).Where("id = ?", userID).Update("last_login", now)
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
	// Get and validate query params
	offsetQuery := c.DefaultQuery("offset", "0")
	limitQuery := c.DefaultQuery("limit", "10")

	offset, err := strconv.Atoi(offsetQuery)
	if err != nil || offset < 0 {
		response.BadRequest(c, "Invalid offset format")
		return
	}

	limit, err := strconv.Atoi(limitQuery)
	if err != nil || limit <= 0 || limit > 100 { // 限制最大查询数量
		response.BadRequest(c, "Invalid limit format (max 100)")
		return
	}

	// Create a cache key based on query params
	cacheKey := fmt.Sprintf("users:offset:%d:limit:%d", offset, limit)

	// Try fetching from cache first
	var userResponses []models.UserResponse
	cachedUsers, err := r.RedisClient.Get(*r.Ctx, cacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cachedUsers), &userResponses); err == nil {
			response.Success(c, userResponses)
			return
		}
		// 缓存数据损坏，删除缓存
		r.RedisClient.Del(*r.Ctx, cacheKey)
	}

	// Cache miss, fetch from database with optimized query
	var users []models.User

	// 使用数据库查询获取用户列表
	if err := r.DB.Where("deleted_at IS NULL").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		response.InternalServerError(c, "Failed to fetch users")
		return
	}

	// Convert to response format
	userResponses = make([]models.UserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	// Cache the result asynchronously to avoid blocking the response
	go func() {
		if serializedUsers, err := json.Marshal(userResponses); err == nil {
			r.RedisClient.Set(*r.Ctx, cacheKey, serializedUsers, 2*time.Minute)
		}
	}()

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
		response.BadRequest(c, "Invalid request format")
		return
	}

	// 基本输入验证
	if thirdPartyUser.ProviderUserID == "" || thirdPartyUser.Email == "" {
		response.BadRequest(c, "Provider user ID and email are required")
		return
	}

	// TODO: 这里应该验证第三方平台的token
	// 例如：验证Apple ID token、Google token等
	// if !validateThirdPartyToken(thirdPartyUser.AuthProvider, thirdPartyUser.ProviderToken) {
	//     response.Unauthorized(c, "Invalid provider token")
	//     return
	// }

	// 尝试从缓存获取用户信息
	cacheKey := fmt.Sprintf("third_party_user:%s:%s", thirdPartyUser.AuthProvider, thirdPartyUser.ProviderUserID)
	var dbUser models.User

	cachedUser, err := r.RedisClient.Get(*r.Ctx, cacheKey).Result()
	if err == nil {
		if err := json.Unmarshal([]byte(cachedUser), &dbUser); err == nil && dbUser.IsActive && !dbUser.IsDeleted() {
			// 缓存命中且用户有效
			r.handleSuccessfulLogin(c, &dbUser, true)
			return
		}
		// 缓存数据无效，删除缓存
		r.RedisClient.Del(*r.Ctx, cacheKey)
	}

	// 缓存未命中，从数据库查询
	// 优化查询：添加复合索引查询条件
	err = r.DB.Where("provider_user_id = ? AND auth_provider = ? AND deleted_at IS NULL AND is_active = ?",
		thirdPartyUser.ProviderUserID, thirdPartyUser.AuthProvider, true).First(&dbUser).Error()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 用户不存在，创建新用户
			dbUser = r.createThirdPartyUser(&thirdPartyUser)
			if dbUser.ID == (uuid.UUID{}) { // 检查是否创建失败
				response.InternalServerError(c, "Failed to create user account")
				return
			}
		} else {
			response.InternalServerError(c, "Authentication service temporarily unavailable")
			return
		}
	}

	// 缓存用户信息
	if userBytes, err := json.Marshal(dbUser); err == nil {
		r.RedisClient.Set(*r.Ctx, cacheKey, userBytes, 10*time.Minute)
	}

	r.handleSuccessfulLogin(c, &dbUser, false)
}

// createThirdPartyUser 创建第三方登录用户
func (r *userRepository) createThirdPartyUser(thirdPartyUser *models.ThirdPartyLoginUser) models.User {
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
		// 可能是邮箱冲突，尝试通过邮箱查找现有用户
		var existingUser models.User
		if err := r.DB.Where("email = ? AND deleted_at IS NULL", thirdPartyUser.Email).First(&existingUser).Error(); err == nil {
			// 更新现有用户的第三方登录信息
			updateData := models.User{
				ProviderUserID: &thirdPartyUser.ProviderUserID,
				AuthProvider:   thirdPartyUser.AuthProvider,
			}
			if thirdPartyUser.AvatarURL != "" {
				updateData.AvatarURL = &thirdPartyUser.AvatarURL
			}
			r.DB.Model(&existingUser).Updates(updateData)
			return existingUser
		}
		return models.User{} // 返回空用户表示创建失败
	}

	return newUser
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

	// 调试信息：检查获取到的email
	fmt.Printf("Debug - JWT extracted email: '%s'\n", email)

	// 类型断言确保email是字符串
	emailStr, ok := email.(string)
	if !ok || emailStr == "" {
		fmt.Printf("Debug - Invalid email type or empty: %T, value: %v\n", email, email)
		response.Unauthorized(c, "Invalid user information in token")
		return
	}

	var user models.User
	if err := r.DB.Where("email = ? AND deleted_at IS NULL", emailStr).First(&user).Error(); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			fmt.Printf("Debug - User not found with email: '%s'\n", emailStr)
			response.NotFound(c, "User not found")
		} else {
			fmt.Printf("Debug - Database error: %v\n", err)
			response.InternalServerError(c, "Database error")
		}
		return
	}

	fmt.Printf("Debug - User found: ID=%s, Email=%s\n", user.ID, user.Email)
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
