package dto

import (
	"golang-rest-api-template/pkg/models"
	"time"
)

// CreateBookRequest 创建图书请求DTO
type CreateBookRequest struct {
	Title  string `json:"title" binding:"required,min=1,max=200"`
	Author string `json:"author" binding:"required,min=1,max=100"`
}

// UpdateBookRequest 更新图书请求DTO
type UpdateBookRequest struct {
	Title  *string `json:"title,omitempty" binding:"omitempty,min=1,max=200"`
	Author *string `json:"author,omitempty" binding:"omitempty,min=1,max=100"`
}

// BookResponse 图书响应DTO
type BookResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListBooksRequest 列表查询请求DTO
type ListBooksRequest struct {
	Offset int `form:"offset,default=0" binding:"min=0"`
	Limit  int `form:"limit,default=10" binding:"min=1,max=100"`
}

// ToCreateModel 转换为创建模型
func (r CreateBookRequest) ToCreateModel() models.CreateBook {
	return models.CreateBook{
		Title:  r.Title,
		Author: r.Author,
	}
}

// ToUpdateModel 转换为更新模型
func (r UpdateBookRequest) ToUpdateModel() models.UpdateBook {
	result := models.UpdateBook{}
	if r.Title != nil {
		result.Title = *r.Title
	}
	if r.Author != nil {
		result.Author = *r.Author
	}
	return result
}

// FromModel 从模型转换为响应DTO
func (BookResponse) FromModel(book models.Book) BookResponse {
	return BookResponse{
		ID:        book.ID,
		Title:     book.Title,
		Author:    book.Author,
		CreatedAt: book.CreatedAt,
		UpdatedAt: book.UpdatedAt,
	}
}

// FromModels 从模型列表转换为响应DTO列表
func (BookResponse) FromModels(books []models.Book) []BookResponse {
	responses := make([]BookResponse, 0, len(books))
	for _, book := range books {
		responses = append(responses, BookResponse{}.FromModel(book))
	}
	return responses
}
