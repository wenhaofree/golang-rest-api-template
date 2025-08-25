package repositories

import (
	"context"
	"golang-rest-api-template/pkg/database"
	"golang-rest-api-template/pkg/models"

	"gorm.io/gorm"
)

// BookRepository 图书数据访问接口
type BookRepository interface {
	// 基础CRUD操作
	Create(ctx context.Context, book *models.Book) error
	FindByID(ctx context.Context, id string) (*models.Book, error)
	Update(ctx context.Context, book *models.Book) error
	Delete(ctx context.Context, book *models.Book) error

	// 查询操作
	List(ctx context.Context, offset, limit int) ([]models.Book, error)
	FindByTitle(ctx context.Context, title string) (*models.Book, error)
	FindByAuthor(ctx context.Context, author string) ([]models.Book, error)

	// 业务查询
	TitleExists(ctx context.Context, title string) (bool, error)
	Count(ctx context.Context) (int64, error)
}

// bookRepository 图书数据访问实现
type bookRepository struct {
	db database.Database
}

// NewBookRepository 创建图书数据访问层
func NewBookRepository(db database.Database) BookRepository {
	return &bookRepository{db: db}
}

// Create 创建图书
func (r *bookRepository) Create(ctx context.Context, book *models.Book) error {
	return r.db.WithContext(ctx).Create(book).Error
}

// FindByID 根据ID查找图书
func (r *bookRepository) FindByID(ctx context.Context, id string) (*models.Book, error) {
	var book models.Book
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&book).Error(); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &book, nil
}

// Update 更新图书
func (r *bookRepository) Update(ctx context.Context, book *models.Book) error {
	return r.db.WithContext(ctx).Save(book).Error
}

// Delete 删除图书
func (r *bookRepository) Delete(ctx context.Context, book *models.Book) error {
	return r.db.WithContext(ctx).Delete(book).Error
}

// List 获取图书列表
func (r *bookRepository) List(ctx context.Context, offset, limit int) ([]models.Book, error) {
	var books []models.Book
	if err := r.db.WithContext(ctx).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&books).Error; err != nil {
		return nil, err
	}
	return books, nil
}

// FindByTitle 根据标题查找图书
func (r *bookRepository) FindByTitle(ctx context.Context, title string) (*models.Book, error) {
	var book models.Book
	if err := r.db.WithContext(ctx).Where("title = ?", title).First(&book).Error(); err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &book, nil
}

// FindByAuthor 根据作者查找图书
func (r *bookRepository) FindByAuthor(ctx context.Context, author string) ([]models.Book, error) {
	var books []models.Book
	if err := r.db.WithContext(ctx).Where("author = ?", author).Find(&books).Error; err != nil {
		return nil, err
	}
	return books, nil
}

// TitleExists 检查标题是否存在
func (r *bookRepository) TitleExists(ctx context.Context, title string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&models.Book{}).
		Where("title = ?", title).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Count 获取图书总数
func (r *bookRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&models.Book{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
