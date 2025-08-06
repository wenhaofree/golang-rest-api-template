package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUserModel(t *testing.T) {
	t.Run("Test User ToResponse", func(t *testing.T) {
		userID := uuid.New()
		fullName := "张三"
		avatarURL := "https://example.com/avatar.jpg"
		now := time.Now()
		updatedTime := now.Add(time.Hour)

		user := User{
			ID:           userID,
			Email:        "test@example.com",
			IsActive:     true,
			IsSuperuser:  false,
			FullName:     &fullName,
			Platform:     PlatformWeb,
			AuthProvider: AuthProviderEmail,
			AvatarURL:    &avatarURL,
			LastLogin:    &now,
			CreatedAt:    now,
			UpdatedAt:    updatedTime,
		}

		response := user.ToResponse()

		assert.Equal(t, userID, response.ID)
		assert.Equal(t, "test@example.com", response.Email)
		assert.Equal(t, true, response.IsActive)
		assert.Equal(t, false, response.IsSuperuser)
		assert.Equal(t, &fullName, response.FullName)
		assert.Equal(t, PlatformWeb, response.Platform)
		assert.Equal(t, AuthProviderEmail, response.AuthProvider)
		assert.Equal(t, &avatarURL, response.AvatarURL)
		assert.Equal(t, &now, response.LastLogin)
		assert.Equal(t, now, response.CreatedAt)
		assert.Equal(t, updatedTime, response.UpdatedAt)
	})

	t.Run("Test User IsDeleted", func(t *testing.T) {
		user := User{}
		assert.False(t, user.IsDeleted())

		now := time.Now()
		user.DeletedAt = &now
		assert.True(t, user.IsDeleted())
	})

	t.Run("Test User SoftDelete", func(t *testing.T) {
		user := User{}
		assert.Nil(t, user.DeletedAt)

		user.SoftDelete()
		assert.NotNil(t, user.DeletedAt)
		assert.True(t, user.IsDeleted())
	})
}

func TestEnums(t *testing.T) {
	t.Run("Test PlatformEnum values", func(t *testing.T) {
		assert.Equal(t, PlatformEnum("web"), PlatformWeb)
		assert.Equal(t, PlatformEnum("mobile"), PlatformMobile)
		assert.Equal(t, PlatformEnum("app"), PlatformApp)
	})

	t.Run("Test AuthProviderEnum values", func(t *testing.T) {
		assert.Equal(t, AuthProviderEnum("email"), AuthProviderEmail)
		assert.Equal(t, AuthProviderEnum("apple_id"), AuthProviderAppleID)
		assert.Equal(t, AuthProviderEnum("google"), AuthProviderGoogle)
		assert.Equal(t, AuthProviderEnum("facebook"), AuthProviderFacebook)
		assert.Equal(t, AuthProviderEnum("github"), AuthProviderGithub)
		assert.Equal(t, AuthProviderEnum("wechat"), AuthProviderWechat)
	})
}

func TestUserStructs(t *testing.T) {
	t.Run("Test LoginUser struct", func(t *testing.T) {
		loginUser := LoginUser{
			Email:    "test@example.com",
			Password: "password123",
		}

		assert.Equal(t, "test@example.com", loginUser.Email)
		assert.Equal(t, "password123", loginUser.Password)
	})

	t.Run("Test ThirdPartyLoginUser struct", func(t *testing.T) {
		thirdPartyUser := ThirdPartyLoginUser{
			AuthProvider:   AuthProviderAppleID,
			ProviderUserID: "001234.567890abcdef.1234",
			Email:          "user@privaterelay.appleid.com",
			FullName:       "张三",
			AvatarURL:      "https://example.com/avatar.jpg",
			Platform:       PlatformMobile,
			ProviderToken:  "apple_token_here",
		}

		assert.Equal(t, AuthProviderAppleID, thirdPartyUser.AuthProvider)
		assert.Equal(t, "001234.567890abcdef.1234", thirdPartyUser.ProviderUserID)
		assert.Equal(t, "user@privaterelay.appleid.com", thirdPartyUser.Email)
		assert.Equal(t, "张三", thirdPartyUser.FullName)
		assert.Equal(t, "https://example.com/avatar.jpg", thirdPartyUser.AvatarURL)
		assert.Equal(t, PlatformMobile, thirdPartyUser.Platform)
		assert.Equal(t, "apple_token_here", thirdPartyUser.ProviderToken)
	})

	t.Run("Test RegisterUser struct", func(t *testing.T) {
		registerUser := RegisterUser{
			Email:        "test@example.com",
			Password:     "password123",
			FullName:     "张三",
			Platform:     PlatformWeb,
			AuthProvider: AuthProviderEmail,
		}

		assert.Equal(t, "test@example.com", registerUser.Email)
		assert.Equal(t, "password123", registerUser.Password)
		assert.Equal(t, "张三", registerUser.FullName)
		assert.Equal(t, PlatformWeb, registerUser.Platform)
		assert.Equal(t, AuthProviderEmail, registerUser.AuthProvider)
	})

	t.Run("Test UpdateUser struct", func(t *testing.T) {
		fullName := "李四"
		avatarURL := "https://example.com/new-avatar.jpg"
		isActive := false

		updateUser := UpdateUser{
			FullName:  &fullName,
			AvatarURL: &avatarURL,
			IsActive:  &isActive,
		}

		assert.Equal(t, &fullName, updateUser.FullName)
		assert.Equal(t, &avatarURL, updateUser.AvatarURL)
		assert.Equal(t, &isActive, updateUser.IsActive)
	})
}
