package validators

import (
	"fmt"
	"golang-rest-api-template/pkg/apperrors"
	"golang-rest-api-template/pkg/dto"
	"golang-rest-api-template/pkg/models"
	"net/url"
	"regexp"
	"strings"
)

// UserValidator 用户验证器
type UserValidator struct{}

// NewUserValidator 创建用户验证器
func NewUserValidator() *UserValidator {
	return &UserValidator{}
}

// 正则表达式
var (
	// 邮箱正则
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	// 姓名正则（支持中英文、空格、点号）
	nameRegex = regexp.MustCompile(`^[a-zA-Z\p{Han}\s.'-]{2,100}$`)
)

// ValidateLogin 验证登录请求
func (v *UserValidator) ValidateLogin(req *dto.LoginRequest) error {
	// 邮箱验证
	if err := v.validateEmail(req.Email); err != nil {
		return err
	}

	// 密码验证
	if err := v.validatePassword(req.Password, false); err != nil {
		return err
	}

	return nil
}

// ValidateRegister 验证注册请求
func (v *UserValidator) ValidateRegister(req *dto.RegisterRequest) error {
	// 邮箱验证
	if err := v.validateEmail(req.Email); err != nil {
		return err
	}

	// 密码验证（注册时需要强密码）
	if err := v.validatePassword(req.Password, true); err != nil {
		return err
	}

	// 姓名验证
	if err := v.validateFullName(req.FullName); err != nil {
		return err
	}

	// 平台验证
	if err := v.validatePlatform(req.Platform); err != nil {
		return err
	}

	// 认证提供商验证
	if err := v.validateAuthProvider(req.AuthProvider); err != nil {
		return err
	}

	return nil
}

// ValidateThirdPartyLogin 验证第三方登录请求
func (v *UserValidator) ValidateThirdPartyLogin(req *dto.ThirdPartyLoginRequest) error {
	// 认证提供商验证
	if err := v.validateAuthProvider(req.AuthProvider); err != nil {
		return err
	}

	// 提供商用户ID验证
	if err := v.validateProviderUserID(req.ProviderUserID); err != nil {
		return err
	}

	// 邮箱验证（可选）
	if req.Email != "" {
		if err := v.validateEmail(req.Email); err != nil {
			return err
		}
	}

	// 姓名验证（可选）
	if req.FullName != "" {
		if err := v.validateFullName(req.FullName); err != nil {
			return err
		}
	}

	// 头像URL验证（可选）
	if req.AvatarURL != "" {
		if err := v.validateAvatarURL(req.AvatarURL); err != nil {
			return err
		}
	}

	// 平台验证
	if err := v.validatePlatform(req.Platform); err != nil {
		return err
	}

	return nil
}

// ValidateUpdateUser 验证更新用户请求
func (v *UserValidator) ValidateUpdateUser(req *dto.UpdateUserRequest) error {
	// 至少需要一个字段
	if req.FullName == nil && req.AvatarURL == nil && req.IsActive == nil {
		return apperrors.New("EMPTY_UPDATE", "At least one field must be provided for update", 400)
	}

	// 姓名验证（如果提供）
	if req.FullName != nil {
		if err := v.validateFullName(*req.FullName); err != nil {
			return err
		}
	}

	// 头像URL验证（如果提供）
	if req.AvatarURL != nil {
		if err := v.validateAvatarURL(*req.AvatarURL); err != nil {
			return err
		}
	}

	return nil
}

// ValidateListUsers 验证用户列表请求
func (v *UserValidator) ValidateListUsers(req *dto.ListUsersRequest) error {
	if req.Offset < 0 {
		return apperrors.New("INVALID_OFFSET", "Offset must be non-negative", 400)
	}

	if req.Limit <= 0 || req.Limit > 100 {
		return apperrors.New("INVALID_LIMIT", "Limit must be between 1 and 100", 400)
	}

	return nil
}

// 私有验证方法

// validateEmail 验证邮箱
func (v *UserValidator) validateEmail(email string) error {
	email = strings.TrimSpace(email)

	if email == "" {
		return apperrors.New("EMPTY_EMAIL", "Email cannot be empty", 400)
	}

	if len(email) > 255 {
		return apperrors.New("EMAIL_TOO_LONG", "Email cannot exceed 255 characters", 400)
	}

	if !emailRegex.MatchString(email) {
		return apperrors.New("INVALID_EMAIL", "Invalid email format", 400)
	}

	// 检查是否包含危险字符
	if strings.ContainsAny(email, "<>\"'&") {
		return apperrors.New("INVALID_EMAIL_CHARS", "Email contains invalid characters", 400)
	}

	return nil
}

// validatePassword 验证密码
func (v *UserValidator) validatePassword(password string, requireStrong bool) error {
	if password == "" {
		return apperrors.New("EMPTY_PASSWORD", "Password cannot be empty", 400)
	}

	if len(password) < 6 {
		return apperrors.New("PASSWORD_TOO_SHORT", "Password must be at least 6 characters", 400)
	}

	if len(password) > 128 {
		return apperrors.New("PASSWORD_TOO_LONG", "Password cannot exceed 128 characters", 400)
	}

	// 注册时需要强密码验证
	if requireStrong {
		hasLetter := false
		hasDigit := false

		for _, char := range password {
			if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') {
				hasLetter = true
			} else if char >= '0' && char <= '9' {
				hasDigit = true
			}
			if hasLetter && hasDigit {
				break
			}
		}

		if !hasLetter || !hasDigit {
			return apperrors.New("WEAK_PASSWORD", "Password must contain at least one letter and one number", 400)
		}
	}

	return nil
}

// validateFullName 验证姓名
func (v *UserValidator) validateFullName(fullName string) error {
	fullName = strings.TrimSpace(fullName)

	if fullName == "" {
		return apperrors.New("EMPTY_FULL_NAME", "Full name cannot be empty", 400)
	}

	if len(fullName) < 2 {
		return apperrors.New("FULL_NAME_TOO_SHORT", "Full name must be at least 2 characters", 400)
	}

	if len(fullName) > 100 {
		return apperrors.New("FULL_NAME_TOO_LONG", "Full name cannot exceed 100 characters", 400)
	}

	// if !nameRegex.MatchString(fullName) {
	// 	return apperrors.New("INVALID_FULL_NAME", "Full name contains invalid characters", 400)
	// }

	return nil
}

// validatePlatform 验证平台
func (v *UserValidator) validatePlatform(platform models.PlatformEnum) error {
	// 如果为空，使用默认值
	if platform == "" {
		return nil
	}

	validPlatforms := []models.PlatformEnum{
		models.PlatformWeb,
		models.PlatformMobile,
		models.PlatformApp,
	}

	for _, valid := range validPlatforms {
		if platform == valid {
			return nil
		}
	}

	return apperrors.New("INVALID_PLATFORM", fmt.Sprintf("Invalid platform: %s", platform), 400)
}

// validateAuthProvider 验证认证提供商
func (v *UserValidator) validateAuthProvider(provider models.AuthProviderEnum) error {
	if provider == "" {
		return apperrors.New("EMPTY_AUTH_PROVIDER", "Auth provider cannot be empty", 400)
	}

	validProviders := []models.AuthProviderEnum{
		models.AuthProviderEmail,
		models.AuthProviderAppleID,
		models.AuthProviderGoogle,
		models.AuthProviderFacebook,
		models.AuthProviderGithub,
		models.AuthProviderWechat,
	}

	for _, valid := range validProviders {
		if provider == valid {
			return nil
		}
	}

	return apperrors.New("INVALID_AUTH_PROVIDER", fmt.Sprintf("Invalid auth provider: %s", provider), 400)
}

// validateProviderUserID 验证提供商用户ID
func (v *UserValidator) validateProviderUserID(providerUserID string) error {
	providerUserID = strings.TrimSpace(providerUserID)

	if providerUserID == "" {
		return apperrors.New("EMPTY_PROVIDER_USER_ID", "Provider user ID cannot be empty", 400)
	}

	if len(providerUserID) > 255 {
		return apperrors.New("PROVIDER_USER_ID_TOO_LONG", "Provider user ID cannot exceed 255 characters", 400)
	}

	return nil
}

// validateAvatarURL 验证头像URL
func (v *UserValidator) validateAvatarURL(avatarURL string) error {
	avatarURL = strings.TrimSpace(avatarURL)

	if avatarURL == "" {
		return nil // 允许空值
	}

	if len(avatarURL) > 500 {
		return apperrors.New("AVATAR_URL_TOO_LONG", "Avatar URL cannot exceed 500 characters", 400)
	}

	// 验证URL格式
	if _, err := url.Parse(avatarURL); err != nil {
		return apperrors.New("INVALID_AVATAR_URL", "Invalid avatar URL format", 400)
	}

	// 验证是否为HTTPS（安全考虑）
	if !strings.HasPrefix(avatarURL, "https://") {
		return apperrors.New("INSECURE_AVATAR_URL", "Avatar URL must use HTTPS", 400)
	}

	return nil
}

// ValidateEmail 验证邮箱（公共方法）
func (v *UserValidator) ValidateEmail(email string) error {
	return v.validateEmail(email)
}
