package dto

import (
	"golang-rest-api-template/pkg/models"
	"time"

	"github.com/google/uuid"
)

// LoginRequest 登录请求DTO
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// RegisterRequest 注册请求DTO
type RegisterRequest struct {
	Email        string                  `json:"email" binding:"required,email"`
	Password     string                  `json:"password" binding:"required,min=6,max=128"`
	FullName     string                  `json:"full_name" binding:"required,min=2,max=100"`
	Platform     models.PlatformEnum     `json:"platform,omitempty"`
	AuthProvider models.AuthProviderEnum `json:"auth_provider,omitempty"`
}

// ThirdPartyLoginRequest 第三方登录请求DTO
type ThirdPartyLoginRequest struct {
	AuthProvider   models.AuthProviderEnum `json:"auth_provider" binding:"required"`
	ProviderUserID string                  `json:"provider_user_id" binding:"required"`
	Email          string                  `json:"email" binding:"omitempty,email"`
	FullName       string                  `json:"full_name,omitempty"`
	AvatarURL      string                  `json:"avatar_url,omitempty"`
	Platform       models.PlatformEnum     `json:"platform" binding:"required"`
	ProviderToken  string                  `json:"provider_token,omitempty"`
}

// UpdateUserRequest 更新用户请求DTO
type UpdateUserRequest struct {
	FullName  *string `json:"full_name,omitempty" binding:"omitempty,min=2,max=100"`
	AvatarURL *string `json:"avatar_url,omitempty" binding:"omitempty,url"`
	IsActive  *bool   `json:"is_active,omitempty"`
}

// ListUsersRequest 用户列表请求DTO
type ListUsersRequest struct {
	Offset int `form:"offset,default=0" binding:"min=0"`
	Limit  int `form:"limit,default=10" binding:"min=1,max=100"`
}

// UserResponse 用户响应DTO
type UserResponse struct {
	ID           uuid.UUID               `json:"id"`
	Email        string                  `json:"email"`
	IsActive     bool                    `json:"is_active"`
	IsSuperuser  bool                    `json:"is_superuser"`
	FullName     *string                 `json:"full_name"`
	Platform     models.PlatformEnum     `json:"platform"`
	AuthProvider models.AuthProviderEnum `json:"auth_provider"`
	AvatarURL    *string                 `json:"avatar_url"`
	LastLogin    *time.Time              `json:"last_login"`
	CreatedAt    time.Time               `json:"created_at"`
	UpdatedAt    time.Time               `json:"updated_at"`
}

// LoginResponse 登录响应DTO
type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// RegisterResponse 注册响应DTO
type RegisterResponse struct {
	User UserResponse `json:"user"`
}

// RefreshTokenRequest 刷新令牌请求DTO
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// RefreshTokenResponse 刷新令牌响应DTO
type RefreshTokenResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

// 转换方法

// ToLoginModel 转换为登录模型
func (r LoginRequest) ToLoginModel() models.LoginUser {
	return models.LoginUser{
		Email:    r.Email,
		Password: r.Password,
	}
}

// ToRegisterModel 转换为注册模型
func (r RegisterRequest) ToRegisterModel() models.RegisterUser {
	platform := r.Platform
	if platform == "" {
		platform = models.PlatformWeb
	}

	provider := r.AuthProvider
	if provider == "" {
		provider = models.AuthProviderEmail
	}

	return models.RegisterUser{
		Email:        r.Email,
		Password:     r.Password,
		FullName:     r.FullName,
		Platform:     platform,
		AuthProvider: provider,
	}
}

// ToThirdPartyLoginModel 转换为第三方登录模型
func (r ThirdPartyLoginRequest) ToThirdPartyLoginModel() models.ThirdPartyLoginUser {
	return models.ThirdPartyLoginUser{
		AuthProvider:   r.AuthProvider,
		ProviderUserID: r.ProviderUserID,
		Email:          r.Email,
		FullName:       r.FullName,
		AvatarURL:      r.AvatarURL,
		Platform:       r.Platform,
		ProviderToken:  r.ProviderToken,
	}
}

// ToUpdateModel 转换为更新模型
func (r UpdateUserRequest) ToUpdateModel() models.UpdateUser {
	return models.UpdateUser{
		FullName:  r.FullName,
		AvatarURL: r.AvatarURL,
		IsActive:  r.IsActive,
	}
}

// FromUserResponse 从UserResponse模型转换
func (UserResponse) FromUserResponse(userResp models.UserResponse) UserResponse {
	return UserResponse{
		ID:           userResp.ID,
		Email:        userResp.Email,
		IsActive:     userResp.IsActive,
		IsSuperuser:  userResp.IsSuperuser,
		FullName:     userResp.FullName,
		Platform:     userResp.Platform,
		AuthProvider: userResp.AuthProvider,
		AvatarURL:    userResp.AvatarURL,
		LastLogin:    userResp.LastLogin,
		CreatedAt:    userResp.CreatedAt,
		UpdatedAt:    userResp.UpdatedAt,
	}
}

// FromUser 从User模型转换
func (UserResponse) FromUser(user models.User) UserResponse {
	return UserResponse{
		ID:           user.ID,
		Email:        user.Email,
		IsActive:     user.IsActive,
		IsSuperuser:  user.IsSuperuser,
		FullName:     user.FullName,
		Platform:     user.Platform,
		AuthProvider: user.AuthProvider,
		AvatarURL:    user.AvatarURL,
		LastLogin:    user.LastLogin,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
	}
}

// FromUserResponses 从UserResponse列表转换
func (UserResponse) FromUserResponses(userResps []models.UserResponse) []UserResponse {
	responses := make([]UserResponse, 0, len(userResps))
	for _, userResp := range userResps {
		responses = append(responses, UserResponse{}.FromUserResponse(userResp))
	}
	return responses
}
