package handlers

import (
	"context"
	"fmt"
	"golang-rest-api-template/pkg/apperrors"
	"golang-rest-api-template/pkg/dto"
	"golang-rest-api-template/pkg/response"
	"golang-rest-api-template/pkg/services"
	"golang-rest-api-template/pkg/validators"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户HTTP处理器
type UserHandler struct {
	service   services.UserService
	validator *validators.UserValidator
}

// NewUserHandler 创建用户处理器
func NewUserHandler(service services.UserService) *UserHandler {
	return &UserHandler{
		service:   service,
		validator: validators.NewUserValidator(),
	}
}

// @BasePath /api/v1

// Login godoc
// @Summary Authenticate a user
// @Schemes
// @Description Authenticates a user using email and password, returns a JWT token if successful
// @Tags user
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param   user     body    dto.LoginRequest     true        "User login object"
// @Success 200 {object} dto.LoginResponse "JWT Token and user info"
// @Failure 400 {object} response.ErrorEnvelope "Bad Request"
// @Failure 401 {object} response.ErrorEnvelope "Unauthorized"
// @Failure 500 {object} response.ErrorEnvelope "Internal Server Error"
// @Router /login [post]
func (h *UserHandler) Login(c *gin.Context) {
	// 1. 解析请求体
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	// 2. 验证请求
	if err := h.validator.ValidateLogin(&req); err != nil {
		c.Error(err)
		return
	}

	// 3. 调用服务
	token, refreshToken, user, err := h.service.Login(c.Request.Context(), req.ToLoginModel())
	if err != nil {
		// 根据service错误类型映射到适当的HTTP错误
		switch err {
		case services.ErrInvalid:
			c.Error(apperrors.ErrBadRequest)
		case services.ErrUnauthorized:
			c.Error(apperrors.ErrUnauthorized)
		default:
			c.Error(apperrors.ErrInternal)
		}
		return
	}

	// 4. 返回响应
	loginResp := dto.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}

	// 5. 异步更新最后登录时间
	go func() {
		// 创建一个新的context，避免原请求取消影响更新操作
		updateCtx := context.Background()
		if err := h.service.UpdateLastLogin(updateCtx, user.Email); err != nil {
			// 记录错误但不影响登录响应
			// 可以考虑使用结构化日志替换fmt.Printf
			fmt.Printf("Failed to update last login time for user %s: %v\n", user.Email, err)
		}
	}()

	response.Success(c, loginResp)
}

// Register godoc
// @Summary Register a new user
// @Schemes http
// @Description Registers a new user with the given email and password
// @Tags user
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param   user     body    dto.RegisterRequest     true        "User registration object"
// @Success 201 {object} dto.RegisterResponse "Successfully registered"
// @Failure 400 {object} response.ErrorEnvelope "Bad Request"
// @Failure 409 {object} response.ErrorEnvelope "User already exists"
// @Failure 500 {object} response.ErrorEnvelope "Internal Server Error"
// @Router /register [post]
func (h *UserHandler) Register(c *gin.Context) {
	// 1. 解析请求体
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	// 2. 验证请求
	if err := h.validator.ValidateRegister(&req); err != nil {
		c.Error(err)
		return
	}

	// 3. 调用服务
	user, err := h.service.Register(c.Request.Context(), req.ToRegisterModel())
	if err != nil {
		switch err {
		case services.ErrInvalid:
			c.Error(apperrors.ErrBadRequest)
		case services.ErrConflict:
			c.Error(apperrors.ErrConflict)
		default:
			c.Error(apperrors.ErrInternal)
		}
		return
	}

	// 4. 返回响应
	registerResp := dto.RegisterResponse{
		User: dto.UserResponse{}.FromUser(user),
	}
	response.SuccessWithMessage(c, registerResp, "Registration successful")
}

// ThirdPartyLogin godoc
// @Summary Third-party login (Apple ID, Google, etc.)
// @Description Authenticates or registers a user using third-party providers
// @Tags user
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param   user     body    dto.ThirdPartyLoginRequest     true        "Third-party login object"
// @Success 200 {object} dto.LoginResponse "JWT Token and user info"
// @Failure 400 {object} response.ErrorEnvelope "Bad Request"
// @Failure 500 {object} response.ErrorEnvelope "Internal Server Error"
// @Router /auth/third-party [post]
func (h *UserHandler) ThirdPartyLogin(c *gin.Context) {
	// 1. 解析请求体
	var req dto.ThirdPartyLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	// 2. 验证请求
	if err := h.validator.ValidateThirdPartyLogin(&req); err != nil {
		c.Error(err)
		return
	}

	// 3. 调用服务
	token, refreshToken, user, err := h.service.ThirdPartyLogin(c.Request.Context(), req.ToThirdPartyLoginModel())
	if err != nil {
		if err == services.ErrInvalid {
			c.Error(apperrors.ErrBadRequest)
		} else {
			c.Error(apperrors.ErrInternal)
		}
		return
	}

	// 4. 返回响应
	loginResp := dto.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
	}

	// 5. 异步更新最后登录时间
	go func() {
		// 创建一个新的context，避免原请求取消影响更新操作
		updateCtx := context.Background()
		if err := h.service.UpdateLastLogin(updateCtx, user.Email); err != nil {
			// 记录错误但不影响登录响应
			fmt.Printf("Failed to update last login time for user %s: %v\n", user.Email, err)
		}
	}()

	response.Success(c, loginResp)
}

// ListUsers godoc
// @Summary Get all users with pagination
// @Description Get a list of all users with optional pagination (excludes deleted users)
// @Tags user
// @Security ApiKeyAuth
// @Produce json
// @Param offset query int false "Offset for pagination" default(0)
// @Param limit query int false "Limit for pagination" default(10)
// @Success 200 {array} dto.UserResponse "Successfully retrieved list of users"
// @Failure 400 {object} response.ErrorEnvelope "Bad Request"
// @Failure 500 {object} response.ErrorEnvelope "Internal Server Error"
// @Router /users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	// 1. 解析查询参数
	var req dto.ListUsersRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "Invalid query parameters")
		return
	}

	// 2. 验证请求
	if err := h.validator.ValidateListUsers(&req); err != nil {
		c.Error(err)
		return
	}

	// 3. 调用服务
	users, err := h.service.List(c.Request.Context(), req.Offset, req.Limit)
	if err != nil {
		if err == services.ErrInvalid {
			c.Error(apperrors.ErrBadRequest)
		} else {
			c.Error(apperrors.ErrInternal)
		}
		return
	}

	// 4. 返回响应
	response.Success(c, dto.UserResponse{}.FromUserResponses(users))
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get the profile of the authenticated user
// @Tags user
// @Security ApiKeyAuth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} dto.UserResponse "User profile"
// @Failure 401 {object} response.ErrorEnvelope "Unauthorized"
// @Failure 404 {object} response.ErrorEnvelope "User not found"
// @Failure 500 {object} response.ErrorEnvelope "Internal Server Error"
// @Router /profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	// 1. 获取认证用户信息
	email, err := h.getAuthenticatedUserEmail(c)
	if err != nil {
		c.Error(err)
		return
	}

	// 2. 调用服务
	userResp, err := h.service.GetProfile(c.Request.Context(), email)
	if err != nil {
		if err == services.ErrNotFound {
			c.Error(apperrors.ErrNotFound)
		} else {
			c.Error(apperrors.ErrInternal)
		}
		return
	}

	// 3. 返回响应
	response.Success(c, dto.UserResponse{}.FromUserResponse(userResp))
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the profile of the authenticated user
// @Tags user
// @Security ApiKeyAuth
// @Security BearerAuth
// @Accept  json
// @Produce  json
// @Param   user     body    dto.UpdateUserRequest     true        "User update object"
// @Success 200 {object} dto.UserResponse "Updated user profile"
// @Failure 400 {object} response.ErrorEnvelope "Bad Request"
// @Failure 401 {object} response.ErrorEnvelope "Unauthorized"
// @Failure 404 {object} response.ErrorEnvelope "User not found"
// @Failure 500 {object} response.ErrorEnvelope "Internal Server Error"
// @Router /profile [put]
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// 1. 获取认证用户信息
	email, err := h.getAuthenticatedUserEmail(c)
	if err != nil {
		c.Error(err)
		return
	}

	// 2. 解析请求体
	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	// 3. 验证请求
	if err := h.validator.ValidateUpdateUser(&req); err != nil {
		c.Error(err)
		return
	}

	// 4. 调用服务
	userResp, err := h.service.UpdateProfile(c.Request.Context(), email, req.ToUpdateModel())
	if err != nil {
		if err == services.ErrNotFound {
			c.Error(apperrors.ErrNotFound)
		} else {
			c.Error(apperrors.ErrInternal)
		}
		return
	}

	// 5. 返回响应
	response.SuccessWithMessage(c, dto.UserResponse{}.FromUserResponse(userResp), "Profile updated successfully")
}

// SoftDelete godoc
// @Summary Soft delete user account
// @Description Soft delete the authenticated user's account (sets deleted_at timestamp)
// @Tags user
// @Security ApiKeyAuth
// @Security BearerAuth
// @Produce json
// @Success 200 {string} string "Account deleted successfully"
// @Failure 401 {object} response.ErrorEnvelope "Unauthorized"
// @Failure 404 {object} response.ErrorEnvelope "User not found"
// @Failure 500 {object} response.ErrorEnvelope "Internal Server Error"
// @Router /profile [delete]
func (h *UserHandler) SoftDelete(c *gin.Context) {
	// 1. 获取认证用户信息
	email, err := h.getAuthenticatedUserEmail(c)
	if err != nil {
		c.Error(err)
		return
	}

	// 2. 调用服务
	if err := h.service.SoftDelete(c.Request.Context(), email); err != nil {
		if err == services.ErrNotFound {
			c.Error(apperrors.ErrNotFound)
		} else {
			c.Error(apperrors.ErrInternal)
		}
		return
	}

	// 3. 返回响应
	response.SuccessWithMessage(c, nil, "Account deleted successfully")
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Use a valid refresh token to generate new access and refresh tokens
// @Tags auth
// @Security ApiKeyAuth
// @Accept  json
// @Produce  json
// @Param   refresh     body    dto.RefreshTokenRequest     true        "Refresh token object"
// @Success 200 {object} dto.RefreshTokenResponse "New tokens generated successfully"
// @Failure 400 {object} response.ErrorEnvelope "Bad Request"
// @Failure 401 {object} response.ErrorEnvelope "Unauthorized - Invalid refresh token"
// @Failure 500 {object} response.ErrorEnvelope "Internal Server Error"
// @Router /auth/refresh-token [post]
func (h *UserHandler) RefreshToken(c *gin.Context) {
	// 1. 解析请求体
	var req dto.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request format")
		return
	}

	// 2. 调用服务
	newToken, newRefreshToken, err := h.service.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		switch err {
		case services.ErrInvalid:
			c.Error(apperrors.ErrBadRequest)
		case services.ErrUnauthorized:
			c.Error(apperrors.ErrUnauthorized)
		default:
			c.Error(apperrors.ErrInternal)
		}
		return
	}

	// 3. 返回响应
	refreshResp := dto.RefreshTokenResponse{
		Token:        newToken,
		RefreshToken: newRefreshToken,
	}
	response.SuccessWithMessage(c, refreshResp, "Tokens refreshed successfully")
}

// 辅助方法

// getAuthenticatedUserEmail 获取认证用户的邮箱
func (h *UserHandler) getAuthenticatedUserEmail(c *gin.Context) (string, error) {
	email, exists := c.Get("username")
	if !exists {
		return "", apperrors.ErrUnauthorized
	}

	emailStr, ok := email.(string)
	if !ok || emailStr == "" {
		return "", apperrors.ErrUnauthorized
	}

	// 验证邮箱格式
	if err := h.validator.ValidateEmail(emailStr); err != nil {
		return "", apperrors.ErrUnauthorized
	}

	return emailStr, nil
}
