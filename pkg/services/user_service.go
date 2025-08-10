package services

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-rest-api-template/pkg/auth"
	"golang-rest-api-template/pkg/cache"
	"golang-rest-api-template/pkg/database"
	"golang-rest-api-template/pkg/models"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Login(ctx context.Context, req models.LoginUser) (string, models.User, error)
	Register(ctx context.Context, req models.RegisterUser) (models.User, error)
	List(ctx context.Context, offset, limit int) ([]models.UserResponse, error)
	ThirdPartyLogin(ctx context.Context, req models.ThirdPartyLoginUser) (string, models.User, error)
	GetProfile(ctx context.Context, email string) (models.UserResponse, error)
	UpdateProfile(ctx context.Context, email string, update models.UpdateUser) (models.UserResponse, error)
	SoftDelete(ctx context.Context, email string) error
}

type userService struct {
	db    database.Database
	cache cache.Cache
}

func NewUserService(db database.Database, cache cache.Cache) UserService {
	return &userService{db: db, cache: cache}
}

func (s *userService) Login(ctx context.Context, req models.LoginUser) (string, models.User, error) {
	if req.Email == "" || req.Password == "" {
		return "", models.User{}, ErrInvalid
	}
	var u models.User
	cacheKey := fmt.Sprintf("user_login:%s", req.Email)
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
			_ = json.Unmarshal([]byte(cached), &u)
			if u.HashedPassword != nil && u.IsActive && !u.IsDeleted() && u.AuthProvider == models.AuthProviderEmail {
				if bcrypt.CompareHashAndPassword([]byte(*u.HashedPassword), []byte(req.Password)) == nil {
					t, err := auth.GenerateToken(u.Email)
					return t, u, err
				}
			}
		}
	}
	if err := s.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL AND is_active = ?", req.Email, true).First(&u).Error(); err != nil {
		// 防时序攻击
		_ = bcrypt.CompareHashAndPassword([]byte(auth.GetDummyHash()), []byte(req.Password))
		return "", models.User{}, ErrUnauthorized
	}
	if u.HashedPassword == nil || bcrypt.CompareHashAndPassword([]byte(*u.HashedPassword), []byte(req.Password)) != nil {
		return "", models.User{}, ErrUnauthorized
	}
	if s.cache != nil {
		go func() {
			if b, err := json.Marshal(u); err == nil {
				_ = s.cache.Set(ctx, cacheKey, b, 5*time.Minute).Err()
			}
		}()
	}
	t, err := auth.GenerateToken(u.Email)
	return t, u, err
}

func (s *userService) Register(ctx context.Context, req models.RegisterUser) (models.User, error) {
	if req.Email == "" || req.Password == "" {
		return models.User{}, ErrInvalid
	}
	var existing models.User
	if err := s.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", req.Email).First(&existing).Error(); err == nil {
		return models.User{}, ErrConflict
	}
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return models.User{}, ErrInternal
	}
	platform := req.Platform
	if platform == "" {
		platform = models.PlatformWeb
	}
	provider := req.AuthProvider
	if provider == "" {
		provider = models.AuthProviderEmail
	}
	u := models.User{Email: req.Email, FullName: &req.FullName, HashedPassword: &hash, Platform: platform, AuthProvider: provider, IsActive: true}
	if err := s.db.WithContext(ctx).Create(&u).Error; err != nil {
		return models.User{}, ErrInternal
	}
	return u, nil
}

func (s *userService) List(ctx context.Context, offset, limit int) ([]models.UserResponse, error) {
	if offset < 0 || limit <= 0 || limit > 100 {
		return nil, ErrInvalid
	}
	cacheKey := fmt.Sprintf("users:offset:%d:limit:%d", offset, limit)
	var usersResp []models.UserResponse
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
			if err := json.Unmarshal([]byte(cached), &usersResp); err == nil {
				return usersResp, nil
			}
			_ = s.cache.Del(ctx, cacheKey)
		}
	}
	var users []models.User
	if err := s.db.WithContext(ctx).Where("deleted_at IS NULL").Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, ErrInternal
	}
	usersResp = make([]models.UserResponse, 0, len(users))
	for _, u := range users {
		usersResp = append(usersResp, u.ToResponse())
	}
	if s.cache != nil {
		go func() {
			if b, err := json.Marshal(usersResp); err == nil {
				_ = s.cache.Set(ctx, cacheKey, b, 2*time.Minute).Err()
			}
		}()
	}
	return usersResp, nil
}

func (s *userService) ThirdPartyLogin(ctx context.Context, req models.ThirdPartyLoginUser) (string, models.User, error) {
	// 允许非首次登录不传 email；首次创建/绑定必须有 email
	if req.ProviderUserID == "" || req.AuthProvider == "" {
		return "", models.User{}, ErrInvalid
	}
	cacheKey := fmt.Sprintf("third_party_user:%s:%s", req.AuthProvider, req.ProviderUserID)
	var u models.User
	// 1) 缓存命中（非首次登录）
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
			if err := json.Unmarshal([]byte(cached), &u); err == nil && u.IsActive && !u.IsDeleted() {
				t, err := auth.GenerateToken(u.Email)
				return t, u, err
			}
			_ = s.cache.Del(ctx, cacheKey)
		}
	}
	// 2) 数据库命中（非首次登录）
	if err := s.db.WithContext(ctx).
		Where("provider_user_id = ? AND auth_provider = ? AND deleted_at IS NULL AND is_active = ?", req.ProviderUserID, req.AuthProvider, true).
		First(&u).Error(); err == nil {
		// 命中直接返回
		t, err := auth.GenerateToken(u.Email)
		return t, u, err
	}
	// 3) 首次登录（需要 email 才能创建/绑定）
	if req.Email == "" {
		return "", models.User{}, ErrInvalid
	}
	// 尝试创建新用户
	u = models.User{Email: req.Email, FullName: &req.FullName, Platform: req.Platform, AuthProvider: req.AuthProvider, ProviderUserID: &req.ProviderUserID, AvatarURL: &req.AvatarURL, IsActive: true}
	if err := s.db.WithContext(ctx).Create(&u).Error; err != nil {
		// 创建失败，尝试按 email 绑定已有账号（需要检查活跃状态）
		var existing models.User
		if err := s.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", req.Email).First(&existing).Error(); err == nil {
			if !existing.IsActive || existing.IsDeleted() {
				return "", models.User{}, ErrUnauthorized
			}
			update := models.User{ProviderUserID: &req.ProviderUserID, AuthProvider: req.AuthProvider}
			if req.AvatarURL != "" {
				update.AvatarURL = &req.AvatarURL
			}
			if err := s.db.WithContext(ctx).Model(&existing).Updates(update).Error; err != nil {
				return "", models.User{}, ErrInternal
			}
			u = existing
		} else {
			return "", models.User{}, ErrInternal
		}
	}
	// 缓存并返回
	if s.cache != nil {
		if b, err := json.Marshal(u); err == nil {
			_ = s.cache.Set(ctx, cacheKey, b, 10*time.Minute).Err()
		}
	}
	t, err := auth.GenerateToken(u.Email)
	return t, u, err
}

func (s *userService) GetProfile(ctx context.Context, email string) (models.UserResponse, error) {
	if email == "" {
		return models.UserResponse{}, ErrInvalid
	}
	var u models.User
	if err := s.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&u).Error(); err != nil {
		return models.UserResponse{}, ErrNotFound
	}
	return u.ToResponse(), nil
}

func (s *userService) UpdateProfile(ctx context.Context, email string, update models.UpdateUser) (models.UserResponse, error) {
	if email == "" {
		return models.UserResponse{}, ErrInvalid
	}
	var u models.User
	if err := s.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&u).Error(); err != nil {
		return models.UserResponse{}, ErrNotFound
	}
	if err := s.db.WithContext(ctx).Model(&u).Updates(update).Error; err != nil {
		return models.UserResponse{}, ErrInternal
	}
	if err := s.db.WithContext(ctx).Where("email = ?", email).First(&u).Error(); err != nil {
		return models.UserResponse{}, ErrInternal
	}
	return u.ToResponse(), nil
}

func (s *userService) SoftDelete(ctx context.Context, email string) error {
	if email == "" {
		return ErrInvalid
	}
	var u models.User
	if err := s.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&u).Error(); err != nil {
		return ErrNotFound
	}
	u.SoftDelete()
	if err := s.db.WithContext(ctx).Model(&u).Updates(models.User{DeletedAt: u.DeletedAt}).Error; err != nil {
		return ErrInternal
	}
	return nil
}
