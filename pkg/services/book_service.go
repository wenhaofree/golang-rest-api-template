package services

import (
	"context"
	"encoding/json"
	"golang-rest-api-template/pkg/cache"
	"golang-rest-api-template/pkg/database"
	"golang-rest-api-template/pkg/models"
	"strconv"
	"time"
)

type BookService interface {
	ListBooks(ctx context.Context, offset, limit int) ([]models.Book, error)
	CreateBook(ctx context.Context, input models.CreateBook) (models.Book, error)
	GetBook(ctx context.Context, id string) (models.Book, error)
	UpdateBook(ctx context.Context, id string, input models.UpdateBook) (models.Book, error)
	DeleteBook(ctx context.Context, id string) error
}

type bookService struct {
	db    database.Database
	cache cache.Cache
}

func NewBookService(db database.Database, cache cache.Cache) BookService {
	return &bookService{db: db, cache: cache}
}

func (s *bookService) ListBooks(ctx context.Context, offset, limit int) ([]models.Book, error) {
	var books []models.Book
	cacheKey := "books_offset_" + itoa(offset) + "_limit_" + itoa(limit)

	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey).Result(); err == nil {
			if err := json.Unmarshal([]byte(cached), &books); err == nil {
				return books, nil
			}
		}
	}

	s.db.Offset(offset).Limit(limit).Find(&books)
	if s.cache != nil {
		if b, err := json.Marshal(books); err == nil {
			_ = s.cache.Set(ctx, cacheKey, b, time.Minute).Err()
		}
	}
	return books, nil
}

func (s *bookService) CreateBook(ctx context.Context, input models.CreateBook) (models.Book, error) {
	book := models.Book{Title: input.Title, Author: input.Author}
	s.db.Create(&book)
	// invalidate list caches
	if s.cache != nil {
		if keys, err := s.cache.Keys(ctx, "books_offset_*").Result(); err == nil {
			if len(keys) > 0 {
				_, _ = s.cache.Del(ctx, keys...).Result()
			}
		}
	}
	return book, nil
}

func (s *bookService) GetBook(ctx context.Context, id string) (models.Book, error) {
	var book models.Book
	if err := s.db.Where("id = ?", id).First(&book).Error(); err != nil {
		return models.Book{}, ErrNotFound
	}
	return book, nil
}

func (s *bookService) UpdateBook(ctx context.Context, id string, input models.UpdateBook) (models.Book, error) {
	var book models.Book
	if err := s.db.Where("id = ?", id).First(&book).Error(); err != nil {
		return models.Book{}, ErrNotFound
	}
	s.db.Model(&book).Updates(models.Book{Title: input.Title, Author: input.Author})
	return book, nil
}

func (s *bookService) DeleteBook(ctx context.Context, id string) error {
	var book models.Book
	if err := s.db.Where("id = ?", id).First(&book).Error(); err != nil {
		return ErrNotFound
	}
	s.db.Delete(&book)
	return nil
}

// small, allocation-free int to string for small values
func itoa(i int) string {
	return fmtInt(int64(i))
}

func fmtInt(i int64) string {
	// fast path, no fmt import
	return strconv.FormatInt(i, 10)
}
