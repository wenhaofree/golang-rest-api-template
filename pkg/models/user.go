package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// PlatformEnum 定义平台类型枚举
type PlatformEnum string

const (
	PlatformWeb    PlatformEnum = "web"
	PlatformMobile PlatformEnum = "mobile"
	PlatformApp    PlatformEnum = "app"
)

// AuthProviderEnum 定义认证提供商枚举
type AuthProviderEnum string

const (
	AuthProviderEmail    AuthProviderEnum = "email"
	AuthProviderAppleID  AuthProviderEnum = "apple_id"
	AuthProviderGoogle   AuthProviderEnum = "google"
	AuthProviderFacebook AuthProviderEnum = "facebook"
	AuthProviderGithub   AuthProviderEnum = "github"
	AuthProviderWechat   AuthProviderEnum = "wechat"
)

// LoginUser 登录请求结构体
type LoginUser struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// ThirdPartyLoginUser 第三方登录请求结构体
// 说明：Apple/部分平台在非首次登录可能不再返回 email，因此 email 在非首次允许省略
// 首次创建/绑定账号时仍需提供 email（由服务层校验）
type ThirdPartyLoginUser struct {
	AuthProvider   AuthProviderEnum `json:"auth_provider" binding:"required"`
	ProviderUserID string           `json:"provider_user_id" binding:"required"`
	Email          string           `json:"email" binding:"omitempty,email"`
	FullName       string           `json:"full_name"`
	AvatarURL      string           `json:"avatar_url"`
	Platform       PlatformEnum     `json:"platform" binding:"required"`
	ProviderToken  string           `json:"provider_token,omitempty"` // 第三方平台的访问令牌，用于验证
}

// RegisterUser 注册请求结构体
type RegisterUser struct {
	Email        string           `json:"email" binding:"required,email"`
	Password     string           `json:"password" binding:"required,min=6"`
	FullName     string           `json:"full_name" binding:"required"`
	Platform     PlatformEnum     `json:"platform"`
	AuthProvider AuthProviderEnum `json:"auth_provider"`
}

// User 用户模型，对应数据库表结构
type User struct {
	ID             uuid.UUID        `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email          string           `json:"email" gorm:"type:varchar(255);not null;uniqueIndex"`
	IsActive       bool             `json:"is_active" gorm:"not null;default:true"`
	IsSuperuser    bool             `json:"is_superuser" gorm:"not null;default:false"`
	FullName       *string          `json:"full_name" gorm:"type:varchar(255)"`
	HashedPassword *string          `json:"-" gorm:"type:varchar"` // 不在JSON中返回密码
	Platform       PlatformEnum     `json:"platform" gorm:"type:varchar(20);not null;default:'web'"`
	AuthProvider   AuthProviderEnum `json:"auth_provider" gorm:"type:varchar(20);not null;default:'email'"`
	ProviderUserID *string          `json:"provider_user_id" gorm:"type:varchar(255)"` // 第三方平台的用户ID
	AvatarURL      *string          `json:"avatar_url" gorm:"type:varchar(500)"`       // 头像URL
	LastLogin      *time.Time       `json:"last_login" gorm:"type:timestamp(6)"`
	// 时间字段统一放在一起
	CreatedAt time.Time  `json:"created_at" gorm:"type:timestamp(6);not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"type:timestamp(6);not null;default:CURRENT_TIMESTAMP"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"type:timestamp(6)"` // 逻辑删除时间
}

// UserResponse 用户响应结构体，用于API返回
type UserResponse struct {
	ID           uuid.UUID        `json:"id"`
	Email        string           `json:"email"`
	IsActive     bool             `json:"is_active"`
	IsSuperuser  bool             `json:"is_superuser"`
	FullName     *string          `json:"full_name"`
	Platform     PlatformEnum     `json:"platform"`
	AuthProvider AuthProviderEnum `json:"auth_provider"`
	AvatarURL    *string          `json:"avatar_url"`
	LastLogin    *time.Time       `json:"last_login"`
	// 时间字段统一放在一起
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateUser 更新用户信息结构体
type UpdateUser struct {
	FullName  *string `json:"full_name"`
	AvatarURL *string `json:"avatar_url"`
	IsActive  *bool   `json:"is_active"`
}

// BeforeCreate GORM钩子，在创建前设置UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// ToResponse 转换为响应结构体
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:           u.ID,
		Email:        u.Email,
		IsActive:     u.IsActive,
		IsSuperuser:  u.IsSuperuser,
		FullName:     u.FullName,
		Platform:     u.Platform,
		AuthProvider: u.AuthProvider,
		AvatarURL:    u.AvatarURL,
		LastLogin:    u.LastLogin,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
	}
}

// IsDeleted 检查用户是否被逻辑删除
func (u *User) IsDeleted() bool {
	return u.DeletedAt != nil
}

// SoftDelete 执行逻辑删除
func (u *User) SoftDelete() {
	now := time.Now()
	u.DeletedAt = &now
}
