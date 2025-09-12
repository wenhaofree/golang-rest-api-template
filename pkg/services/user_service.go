package services

import (
	"context"
	"encoding/json"
	"fmt"
	"golang-rest-api-template/pkg/auth"
	"golang-rest-api-template/pkg/cache"
	"golang-rest-api-template/pkg/models"
	"golang-rest-api-template/pkg/repositories"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Login(ctx context.Context, req models.LoginUser) (string, string, models.User, error)
	Register(ctx context.Context, req models.RegisterUser) (models.User, error)
	List(ctx context.Context, offset, limit int) ([]models.UserResponse, error)
	ThirdPartyLogin(ctx context.Context, req models.ThirdPartyLoginUser) (string, string, models.User, error)
	GetProfile(ctx context.Context, email string) (models.UserResponse, error)
	UpdateProfile(ctx context.Context, email string, update models.UpdateUser) (models.UserResponse, error)
	SoftDelete(ctx context.Context, email string) error
	RefreshToken(ctx context.Context, refreshToken string) (string, string, error)
	UpdateLastLogin(ctx context.Context, email string) error
}

type userService struct {
	repo  repositories.UserRepository
	cache cache.Cache
}

func NewUserService(repo repositories.UserRepository, cache cache.Cache) UserService {
	return &userService{repo: repo, cache: cache}
}

func (s *userService) Login(ctx context.Context, req models.LoginUser) (string, string, models.User, error) {
	if req.Email == "" || req.Password == "" {
		return "", "", models.User{}, ErrInvalid
	}

	cacheKey := fmt.Sprintf("user_login:%s", req.Email)
	var u *models.User

	// 尝试从缓存获取（优化：减少数据库查询）
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
			var cachedUser models.User
			if err := json.Unmarshal([]byte(cached), &cachedUser); err == nil {
				if cachedUser.HashedPassword != nil && cachedUser.IsActive && !cachedUser.IsDeleted() && cachedUser.AuthProvider == models.AuthProviderEmail {
					// 快速密码验证
					if bcrypt.CompareHashAndPassword([]byte(*cachedUser.HashedPassword), []byte(req.Password)) == nil {
						// 并发生成令牌（优化性能）
						tokenChan := make(chan string, 1)
						refreshTokenChan := make(chan struct {
							token string
							id    string
						}, 1)
						errChan := make(chan error, 2)

						// 并发生成访问令牌
						go func() {
							token, err := auth.GenerateToken(cachedUser.Email)
							if err != nil {
								errChan <- err
								return
							}
							tokenChan <- token
						}()

						// 并发生成刷新令牌
						go func() {
							refreshToken, tokenID, err := auth.GenerateRefreshToken(cachedUser.Email)
							if err != nil {
								errChan <- err
								return
							}
							refreshTokenChan <- struct {
								token string
								id    string
							}{refreshToken, tokenID}
						}()

						// 等待令牌生成完成
						var token, refreshToken, tokenID string
						for i := 0; i < 2; i++ {
							select {
							case token = <-tokenChan:
							case refreshData := <-refreshTokenChan:
								refreshToken = refreshData.token
								tokenID = refreshData.id
							case <-errChan:
								return "", "", models.User{}, ErrInternal
							case <-ctx.Done():
								return "", "", models.User{}, ErrInternal
							}
						}

						// 异步存储刷新令牌（不阻塞响应）
						go func() {
							refreshCacheKey := fmt.Sprintf("refresh_token:%s:%s", cachedUser.Email, tokenID)
							_ = s.cache.Set(context.Background(), refreshCacheKey, "valid", 7*24*time.Hour).Err()
						}()

						return token, refreshToken, cachedUser, nil
					}
				}
			}
		}
	}

	// 从数据库查询（缓存未命中）
	u, err := s.repo.FindActiveByEmail(ctx, req.Email)
	if err != nil {
		return "", "", models.User{}, ErrInternal
	}
	if u == nil {
		// 防时序攻击
		_ = bcrypt.CompareHashAndPassword([]byte(auth.GetDummyHash()), []byte(req.Password))
		return "", "", models.User{}, ErrUnauthorized
	}

	// 验证密码
	if u.HashedPassword == nil || bcrypt.CompareHashAndPassword([]byte(*u.HashedPassword), []byte(req.Password)) != nil {
		return "", "", models.User{}, ErrUnauthorized
	}

	// 检查认证提供商
	if u.AuthProvider != models.AuthProviderEmail {
		return "", "", models.User{}, ErrUnauthorized
	}

	// 并发生成令牌和更新缓存（优化：减少总响应时间）
	tokenChan := make(chan string, 1)
	refreshTokenChan := make(chan struct {
		token string
		id    string
	}, 1)
	errChan := make(chan error, 2)

	// 并发生成访问令牌
	go func() {
		token, err := auth.GenerateToken(u.Email)
		if err != nil {
			errChan <- err
			return
		}
		tokenChan <- token
	}()

	// 并发生成刷新令牌
	go func() {
		refreshToken, tokenID, err := auth.GenerateRefreshToken(u.Email)
		if err != nil {
			errChan <- err
			return
		}
		refreshTokenChan <- struct {
			token string
			id    string
		}{refreshToken, tokenID}
	}()

	// 并发更新用户缓存
	go func() {
		if s.cache != nil {
			if b, err := json.Marshal(*u); err == nil {
				_ = s.cache.Set(context.Background(), cacheKey, b, 10*time.Minute).Err() // 延长缓存时间
			}
		}
	}()

	// 等待令牌生成完成
	var token, refreshToken, tokenID string
	for i := 0; i < 2; i++ {
		select {
		case token = <-tokenChan:
		case refreshData := <-refreshTokenChan:
			refreshToken = refreshData.token
			tokenID = refreshData.id
		case <-errChan:
			return "", "", models.User{}, ErrInternal
		case <-ctx.Done():
			return "", "", models.User{}, ErrInternal
		}
	}

	// 异步存储刷新令牌（不阻塞响应）
	go func() {
		if s.cache != nil {
			refreshCacheKey := fmt.Sprintf("refresh_token:%s:%s", u.Email, tokenID)
			_ = s.cache.Set(context.Background(), refreshCacheKey, "valid", 7*24*time.Hour).Err()
		}
	}()

	return token, refreshToken, *u, nil
}

func (s *userService) Register(ctx context.Context, req models.RegisterUser) (models.User, error) {
	if req.Email == "" || req.Password == "" {
		return models.User{}, ErrInvalid
	}

	// 检查邮箱是否已存在
	exists, err := s.repo.EmailExists(ctx, req.Email)
	if err != nil {
		return models.User{}, ErrInternal
	}
	if exists {
		return models.User{}, ErrConflict
	}

	// 加密密码
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return models.User{}, ErrInternal
	}

	// 设置默认值
	platform := req.Platform
	if platform == "" {
		platform = models.PlatformWeb
	}
	provider := req.AuthProvider
	if provider == "" {
		provider = models.AuthProviderEmail
	}

	// 创建用户
	u := models.User{
		Email:          req.Email,
		FullName:       &req.FullName,
		HashedPassword: &hash,
		Platform:       platform,
		AuthProvider:   provider,
		IsActive:       true,
	}

	if err := s.repo.Create(ctx, &u); err != nil {
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

	// 尝试从缓存获取
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
			if err := json.Unmarshal([]byte(cached), &usersResp); err == nil {
				return usersResp, nil
			}
			_ = s.cache.Del(ctx, cacheKey)
		}
	}

	// 从数据库查询
	users, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, ErrInternal
	}

	// 转换为响应DTO
	usersResp = make([]models.UserResponse, 0, len(users))
	for _, u := range users {
		usersResp = append(usersResp, u.ToResponse())
	}

	// 更新缓存
	if s.cache != nil {
		go func() {
			if b, err := json.Marshal(usersResp); err == nil {
				_ = s.cache.Set(ctx, cacheKey, b, 2*time.Minute).Err()
			}
		}()
	}

	return usersResp, nil
}

func (s *userService) ThirdPartyLogin(ctx context.Context, req models.ThirdPartyLoginUser) (string, string, models.User, error) {
	// 允许非首次登录不传 email；首次创建/绑定必须有 email
	if req.ProviderUserID == "" || req.AuthProvider == "" {
		return "", "", models.User{}, ErrInvalid
	}

	cacheKey := fmt.Sprintf("third_party_user:%s:%s", req.AuthProvider, req.ProviderUserID)
	var u *models.User

	// 1) 缓存命中（非首次登录）
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
			var cachedUser models.User
			if err := json.Unmarshal([]byte(cached), &cachedUser); err == nil && cachedUser.IsActive && !cachedUser.IsDeleted() {
				// 生成访问令牌和刷新令牌
				token, err := auth.GenerateToken(cachedUser.Email)
				if err != nil {
					return "", "", models.User{}, ErrInternal
				}
				refreshToken, _, err := auth.GenerateRefreshToken(cachedUser.Email)
				if err != nil {
					return "", "", models.User{}, ErrInternal
				}
				return token, refreshToken, cachedUser, nil
			}
			_ = s.cache.Del(ctx, cacheKey)
		}
	}

	// 2) 数据库命中（非首次登录）
	u, err := s.repo.FindActiveByProviderUserID(ctx, req.ProviderUserID, req.AuthProvider)
	if err != nil {
		return "", "", models.User{}, ErrInternal
	}
	if u != nil {
		// 生成访问令牌和刷新令牌
		token, err := auth.GenerateToken(u.Email)
		if err != nil {
			return "", "", models.User{}, ErrInternal
		}
		refreshToken, tokenID, err := auth.GenerateRefreshToken(u.Email)
		if err != nil {
			return "", "", models.User{}, ErrInternal
		}
		
		// 将刷新令牌存储到缓存中
		if s.cache != nil {
			refreshCacheKey := fmt.Sprintf("refresh_token:%s:%s", u.Email, tokenID)
			if err := s.cache.Set(ctx, refreshCacheKey, "valid", 7*24*time.Hour).Err(); err != nil {
				fmt.Printf("Warning: Failed to store refresh token in cache during third-party login: %v\n", err)
			}
		}
		
		return token, refreshToken, *u, nil
	}

	// 3) 首次登录（需要 email 才能创建/绑定）
	if req.Email == "" {
		return "", "", models.User{}, ErrInvalid
	}

	// 尝试创建新用户
	newUser := models.User{
		Email:          req.Email,
		FullName:       &req.FullName,
		Platform:       req.Platform,
		AuthProvider:   req.AuthProvider,
		ProviderUserID: &req.ProviderUserID,
		AvatarURL:      &req.AvatarURL,
		IsActive:       true,
	}

	if err := s.repo.Create(ctx, &newUser); err != nil {
		// 创建失败，尝试按 email 绑定已有账号（需要检查活跃状态）
		existing, err := s.repo.FindByEmail(ctx, req.Email)
		if err != nil {
			return "", "", models.User{}, ErrInternal
		}
		if existing == nil {
			return "", "", models.User{}, ErrInternal
		}
		if !existing.IsActive || existing.IsDeleted() {
			return "", "", models.User{}, ErrUnauthorized
		}

		// 更新已有用户的第三方信息
		existing.ProviderUserID = &req.ProviderUserID
		existing.AuthProvider = req.AuthProvider
		if req.AvatarURL != "" {
			existing.AvatarURL = &req.AvatarURL
		}

		if err := s.repo.Update(ctx, existing); err != nil {
			return "", "", models.User{}, ErrInternal
		}
		newUser = *existing
	}

	// 缓存并返回
	if s.cache != nil {
		if b, err := json.Marshal(newUser); err == nil {
			_ = s.cache.Set(ctx, cacheKey, b, 10*time.Minute).Err()
		}
	}

	// 生成访问令牌和刷新令牌
	token, err := auth.GenerateToken(newUser.Email)
	if err != nil {
		return "", "", models.User{}, ErrInternal
	}
	refreshToken, tokenID, err := auth.GenerateRefreshToken(newUser.Email)
	if err != nil {
		return "", "", models.User{}, ErrInternal
	}

	// 将刷新令牌存储到缓存中
	if s.cache != nil {
		refreshCacheKey := fmt.Sprintf("refresh_token:%s:%s", newUser.Email, tokenID)
		if err := s.cache.Set(ctx, refreshCacheKey, "valid", 7*24*time.Hour).Err(); err != nil {
			fmt.Printf("Warning: Failed to store refresh token in cache during new user creation: %v\n", err)
		}
	}

	return token, refreshToken, newUser, nil
}

func (s *userService) GetProfile(ctx context.Context, email string) (models.UserResponse, error) {
	if email == "" {
		return models.UserResponse{}, ErrInvalid
	}

	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return models.UserResponse{}, ErrInternal
	}
	if u == nil {
		return models.UserResponse{}, ErrNotFound
	}

	return u.ToResponse(), nil
}

func (s *userService) UpdateProfile(ctx context.Context, email string, update models.UpdateUser) (models.UserResponse, error) {
	if email == "" {
		return models.UserResponse{}, ErrInvalid
	}

	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return models.UserResponse{}, ErrInternal
	}
	if u == nil {
		return models.UserResponse{}, ErrNotFound
	}

	// 更新字段
	if update.FullName != nil {
		u.FullName = update.FullName
	}
	if update.AvatarURL != nil {
		u.AvatarURL = update.AvatarURL
	}
	if update.IsActive != nil {
		u.IsActive = *update.IsActive
	}

	if err := s.repo.Update(ctx, u); err != nil {
		return models.UserResponse{}, ErrInternal
	}

	return u.ToResponse(), nil
}

func (s *userService) SoftDelete(ctx context.Context, email string) error {
	if email == "" {
		return ErrInvalid
	}

	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return ErrInternal
	}
	if u == nil {
		return ErrNotFound
	}

	if err := s.repo.SoftDelete(ctx, u); err != nil {
		return ErrInternal
	}

	return nil
}

// RefreshToken 使用refresh token生成新的访问令牌和刷新令牌
func (s *userService) RefreshToken(ctx context.Context, refreshToken string) (string, string, error) {
	if refreshToken == "" {
		return "", "", ErrInvalid
	}

	// 验证refresh token
	username, tokenID, err := auth.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", "", ErrUnauthorized
	}

	// 检查refresh token是否在缓存中存在且有效（非强制性检查）
	cacheHit := false
	if s.cache != nil {
		refreshCacheKey := fmt.Sprintf("refresh_token:%s:%s", username, tokenID)
		if _, err := s.cache.Get(ctx, refreshCacheKey).Result(); err == nil {
			cacheHit = true
			// 撤销旧的refresh token
			_ = s.cache.Del(ctx, refreshCacheKey).Err()
		}
		// 注意：不在这里返回错误，继续验证用户状态
	}

	// 验证用户是否仍然有效（这是主要的安全检查）
	user, err := s.repo.FindActiveByEmail(ctx, username)
	if err != nil {
		return "", "", ErrInternal
	}
	if user == nil || !user.IsActive || user.IsDeleted() {
		return "", "", ErrUnauthorized
	}

	// 如果没有缓存命中，检查refresh token的时间是否过于久远（额外的安全措施）
	if !cacheHit {
		// 可以在这里添加额外的检查，比如检查上次登录时间等
		// 目前我们信任JWT的过期时间验证
	}

	// 生成新的访问令牌和刷新令牌
	newToken, err := auth.GenerateToken(username)
	if err != nil {
		return "", "", ErrInternal
	}

	newRefreshToken, newTokenID, err := auth.GenerateRefreshToken(username)
	if err != nil {
		return "", "", ErrInternal
	}

	// 将新的刷新令牌存储到缓存中（尽力而为）
	if s.cache != nil {
		newRefreshCacheKey := fmt.Sprintf("refresh_token:%s:%s", username, newTokenID)
		if err := s.cache.Set(ctx, newRefreshCacheKey, "valid", 7*24*time.Hour).Err(); err != nil {
			// 记录错误但不阻止操作
			fmt.Printf("Warning: Failed to store new refresh token in cache: %v\n", err)
		}
	}

	return newToken, newRefreshToken, nil
}

// UpdateLastLogin 更新用户最后登录时间
func (s *userService) UpdateLastLogin(ctx context.Context, email string) error {
	if email == "" {
		return ErrInvalid
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return ErrInternal
	}
	if user == nil {
		return ErrNotFound
	}

	// 更新最后登录时间
	now := time.Now()
	user.LastLogin = &now
	
	if err := s.repo.Update(ctx, user); err != nil {
		return ErrInternal
	}

	return nil
}
