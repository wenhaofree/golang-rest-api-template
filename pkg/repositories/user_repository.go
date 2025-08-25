package repositories

import (
	"context"
	"golang-rest-api-template/pkg/database"
	"golang-rest-api-template/pkg/models"

	"gorm.io/gorm"
)

// UserRepository 用户数据访问接口
type UserRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByProviderUserID(ctx context.Context, providerUserID string, authProvider models.AuthProviderEnum) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	SoftDelete(ctx context.Context, user *models.User) error

	// 查询操作
	List(ctx context.Context, offset, limit int) ([]models.User, error)
	FindActiveByEmail(ctx context.Context, email string) (*models.User, error)
	FindActiveByProviderUserID(ctx context.Context, providerUserID string, authProvider models.AuthProviderEnum) (*models.User, error)

	// 业务查询
	EmailExists(ctx context.Context, email string) (bool, error)
	ProviderUserExists(ctx context.Context, providerUserID string, authProvider models.AuthProviderEnum) (bool, error)
}

// userRepository 用户数据访问实现
type userRepository struct {
	db database.Database
}

// NewUserRepository 创建用户数据访问层
func NewUserRepository(db database.Database) UserRepository {
	return &userRepository{db: db}
}

// Create 创建用户
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return err
	}
	return nil
}

// FindByID 根据ID查找用户
func (r *userRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", id).First(&user).Error(); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ? AND deleted_at IS NULL", email).First(&user).Error(); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByProviderUserID 根据第三方用户ID查找用户
func (r *userRepository) FindByProviderUserID(ctx context.Context, providerUserID string, authProvider models.AuthProviderEnum) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Where("provider_user_id = ? AND auth_provider = ? AND deleted_at IS NULL", providerUserID, authProvider).
		First(&user).Error(); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// Update 更新用户
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// SoftDelete 软删除用户
func (r *userRepository) SoftDelete(ctx context.Context, user *models.User) error {
	user.SoftDelete()
	return r.db.WithContext(ctx).Model(user).Updates(models.User{DeletedAt: user.DeletedAt}).Error
}

// List 获取用户列表
func (r *userRepository) List(ctx context.Context, offset, limit int) ([]models.User, error) {
	var users []models.User
	if err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// FindActiveByEmail 查找活跃用户（根据邮箱）
func (r *userRepository) FindActiveByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Where("email = ? AND deleted_at IS NULL AND is_active = ?", email, true).
		First(&user).Error(); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindActiveByProviderUserID 查找活跃用户（根据第三方用户ID）
func (r *userRepository) FindActiveByProviderUserID(ctx context.Context, providerUserID string, authProvider models.AuthProviderEnum) (*models.User, error) {
	var user models.User
	if err := r.db.WithContext(ctx).
		Where("provider_user_id = ? AND auth_provider = ? AND deleted_at IS NULL AND is_active = ?", providerUserID, authProvider, true).
		First(&user).Error(); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// EmailExists 检查邮箱是否存在
func (r *userRepository) EmailExists(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("email = ? AND deleted_at IS NULL", email).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// ProviderUserExists 检查第三方用户是否存在
func (r *userRepository) ProviderUserExists(ctx context.Context, providerUserID string, authProvider models.AuthProviderEnum) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.User{}).
		Where("provider_user_id = ? AND auth_provider = ? AND deleted_at IS NULL", providerUserID, authProvider).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
