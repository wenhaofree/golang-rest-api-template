package services

import (
	"context"
	"encoding/json"
	"golang-rest-api-template/pkg/cache"
	"golang-rest-api-template/pkg/models"
	"golang-rest-api-template/pkg/repositories"
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
	repo  repositories.BookRepository
	cache cache.Cache
}

func NewBookService(repo repositories.BookRepository, cache cache.Cache) BookService {
	return &bookService{repo: repo, cache: cache}
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

	// 使用Repository查询
	books, err := s.repo.List(ctx, offset, limit)
	if err != nil {
		return nil, ErrInternal
	}

	if s.cache != nil {
		if b, err := json.Marshal(books); err == nil {
			_ = s.cache.Set(ctx, cacheKey, b, time.Minute).Err()
		}
	}
	return books, nil
}

func (s *bookService) CreateBook(ctx context.Context, input models.CreateBook) (models.Book, error) {
	book := models.Book{Title: input.Title, Author: input.Author}
	if err := s.repo.Create(ctx, &book); err != nil {
		return models.Book{}, ErrInternal
	}

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
	book, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return models.Book{}, ErrInternal
	}
	if book == nil {
		return models.Book{}, ErrNotFound
	}
	return *book, nil
}

func (s *bookService) UpdateBook(ctx context.Context, id string, input models.UpdateBook) (models.Book, error) {
	book, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return models.Book{}, ErrInternal
	}
	if book == nil {
		return models.Book{}, ErrNotFound
	}

	// 更新字段
	if input.Title != "" {
		book.Title = input.Title
	}
	if input.Author != "" {
		book.Author = input.Author
	}

	if err := s.repo.Update(ctx, book); err != nil {
		return models.Book{}, ErrInternal
	}

	return *book, nil
}

func (s *bookService) DeleteBook(ctx context.Context, id string) error {
	book, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return ErrInternal
	}
	if book == nil {
		return ErrNotFound
	}

	if err := s.repo.Delete(ctx, book); err != nil {
		return ErrInternal
	}

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
