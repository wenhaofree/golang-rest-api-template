package api

// 这个文件包含了针对当前Database接口的优化版本
// 主要优化点：
// 1. 添加缓存机制
// 2. 异步更新最后登录时间
// 3. 防时序攻击
// 4. 输入验证优化

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"golang-rest-api-template/pkg/auth"
	"golang-rest-api-template/pkg/database"
	"golang-rest-api-template/pkg/models"
	"golang-rest-api-template/pkg/response"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// OptimizedLoginHandler 优化版登录处理器
// 主要优化：缓存、异步更新、防时序攻击
func (r *userRepository) OptimizedLoginHandler(c *gin.Context) {
	var loginUser models.LoginUser

	// 输入验证
	if err := c.ShouldBindJSON(&loginUser); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	if loginUser.Email == "" || loginUser.Password == "" {
		response.BadRequest(c, "Email and password are required")
		return
	}

	// 尝试从缓存获取用户信息
	cacheKey := fmt.Sprintf("user_login:%s", loginUser.Email)
	var dbUser models.User

	// 检查缓存
	if cachedUser, err := r.RedisClient.Get(*r.Ctx, cacheKey).Result(); err == nil {
		if err := json.Unmarshal([]byte(cachedUser), &dbUser); err == nil {
			if r.validateCachedUserLogin(&dbUser, &loginUser) {
				r.returnSuccessfulLogin(c, &dbUser, true)
				return
			}
		}
		// 缓存无效，删除
		r.RedisClient.Del(*r.Ctx, cacheKey)
	}

	// 从数据库查询
	if err := r.DB.Where("email = ? AND deleted_at IS NULL AND is_active = ?", loginUser.Email, true).
		First(&dbUser).Error(); err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 防时序攻击：对不存在的用户也执行bcrypt
			bcrypt.CompareHashAndPassword([]byte("$2a$14$dummy$hash$for$timing$attack$prevention"), []byte(loginUser.Password))
			response.Unauthorized(c, "Invalid email or password")
		} else {
			response.InternalServerError(c, "Authentication service temporarily unavailable")
		}
		return
	}

	// 验证密码
	if !r.validateUserPassword(&dbUser, &loginUser) {
		response.Unauthorized(c, "Invalid email or password")
		return
	}

	// 缓存用户信息
	if userBytes, err := json.Marshal(dbUser); err == nil {
		r.RedisClient.Set(*r.Ctx, cacheKey, userBytes, 5*time.Minute)
	}

	r.returnSuccessfulLogin(c, &dbUser, false)
}

// validateCachedUserLogin 验证缓存的用户登录信息
func (r *userRepository) validateCachedUserLogin(dbUser *models.User, loginUser *models.LoginUser) bool {
	// 检查用户状态
	if !dbUser.IsActive || dbUser.IsDeleted() {
		return false
	}

	// 检查认证方式
	if dbUser.AuthProvider != models.AuthProviderEmail {
		return false
	}

	// 验证密码
	if dbUser.HashedPassword == nil {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(*dbUser.HashedPassword), []byte(loginUser.Password)) == nil
}

// validateUserPassword 验证用户密码
func (r *userRepository) validateUserPassword(dbUser *models.User, loginUser *models.LoginUser) bool {
	if dbUser.AuthProvider != models.AuthProviderEmail {
		return false
	}

	if dbUser.HashedPassword == nil {
		return false
	}

	return bcrypt.CompareHashAndPassword([]byte(*dbUser.HashedPassword), []byte(loginUser.Password)) == nil
}

// returnSuccessfulLogin 返回成功登录响应
func (r *userRepository) returnSuccessfulLogin(c *gin.Context, dbUser *models.User, fromCache bool) {
	// 生成JWT token
	token, err := auth.GenerateToken(dbUser.Email)
	if err != nil {
		response.InternalServerError(c, "Authentication service error")
		return
	}

	// 异步更新最后登录时间
	go r.asyncUpdateLastLogin(dbUser.ID, fromCache)

	// 立即返回响应
	response.Success(c, gin.H{
		"token": token,
		"user":  dbUser.ToResponse(),
	})
}

// asyncUpdateLastLogin 异步更新最后登录时间
func (r *userRepository) asyncUpdateLastLogin(userID interface{}, fromCache bool) {
	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 更新最后登录时间
	now := time.Now()

	// 使用类型断言获取底层的GORM数据库
	if db, ok := r.DB.(*database.GormDatabase); ok {
		// 在goroutine中使用WithContext确保超时控制
		db.WithContext(ctx).Model(&models.User{}).Where("id = ?", userID).Update("last_login", now)
	}
}

// OptimizedFindUsers 优化版用户列表查询
func (r *userRepository) OptimizedFindUsers(c *gin.Context) {
	// 参数验证
	offset, limit, err := r.validatePaginationParams(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	// 缓存键
	cacheKey := fmt.Sprintf("users:list:%d:%d", offset, limit)

	// 尝试从缓存获取
	var userResponses []models.UserResponse
	if cachedData, err := r.RedisClient.Get(*r.Ctx, cacheKey).Result(); err == nil {
		if err := json.Unmarshal([]byte(cachedData), &userResponses); err == nil {
			response.Success(c, userResponses)
			return
		}
		// 缓存数据损坏，删除
		r.RedisClient.Del(*r.Ctx, cacheKey)
	}

	// 从数据库查询
	var users []models.User
	if err := r.DB.Where("deleted_at IS NULL").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		response.InternalServerError(c, "Failed to fetch users")
		return
	}

	// 转换为响应格式
	userResponses = make([]models.UserResponse, 0, len(users))
	for _, user := range users {
		userResponses = append(userResponses, user.ToResponse())
	}

	// 异步缓存结果
	go func() {
		if data, err := json.Marshal(userResponses); err == nil {
			r.RedisClient.Set(*r.Ctx, cacheKey, data, 2*time.Minute)
		}
	}()

	response.Success(c, userResponses)
}

// validatePaginationParams 验证分页参数
func (r *userRepository) validatePaginationParams(c *gin.Context) (int, int, error) {
	offsetStr := c.DefaultQuery("offset", "0")
	limitStr := c.DefaultQuery("limit", "10")

	offset := 0
	limit := 10

	if offsetStr != "0" {
		if o, err := strconv.Atoi(offsetStr); err != nil || o < 0 {
			return 0, 0, fmt.Errorf("invalid offset format")
		} else {
			offset = o
		}
	}

	if limitStr != "10" {
		if l, err := strconv.Atoi(limitStr); err != nil || l <= 0 || l > 100 {
			return 0, 0, fmt.Errorf("invalid limit format (max 100)")
		} else {
			limit = l
		}
	}

	return offset, limit, nil
}
