package validators

import (
	"fmt"
	"golang-rest-api-template/pkg/apperrors"
	"golang-rest-api-template/pkg/dto"
	"strings"
)

// BookValidator 图书验证器
type BookValidator struct{}

// NewBookValidator 创建图书验证器
func NewBookValidator() *BookValidator {
	return &BookValidator{}
}

// ValidateCreateBook 验证创建图书请求
func (v *BookValidator) ValidateCreateBook(req *dto.CreateBookRequest) error {
	// 标题验证
	if err := v.validateTitle(req.Title); err != nil {
		return err
	}

	// 作者验证
	if err := v.validateAuthor(req.Author); err != nil {
		return err
	}

	return nil
}

// ValidateUpdateBook 验证更新图书请求
func (v *BookValidator) ValidateUpdateBook(req *dto.UpdateBookRequest) error {
	// 至少需要一个字段
	if req.Title == nil && req.Author == nil {
		return apperrors.New("INVALID_REQUEST", "At least one field must be provided for update", 400)
	}

	// 标题验证（如果提供）
	if req.Title != nil {
		if err := v.validateTitle(*req.Title); err != nil {
			return err
		}
	}

	// 作者验证（如果提供）
	if req.Author != nil {
		if err := v.validateAuthor(*req.Author); err != nil {
			return err
		}
	}

	return nil
}

// ValidateListBooks 验证列表查询请求
func (v *BookValidator) ValidateListBooks(req *dto.ListBooksRequest) error {
	if req.Offset < 0 {
		return apperrors.New("INVALID_OFFSET", "Offset must be non-negative", 400)
	}

	if req.Limit <= 0 || req.Limit > 100 {
		return apperrors.New("INVALID_LIMIT", "Limit must be between 1 and 100", 400)
	}

	return nil
}

// validateTitle 验证标题
func (v *BookValidator) validateTitle(title string) error {
	title = strings.TrimSpace(title)

	if len(title) == 0 {
		return apperrors.New("EMPTY_TITLE", "Title cannot be empty", 400)
	}

	if len(title) < 1 {
		return apperrors.New("TITLE_TOO_SHORT", "Title must be at least 1 character", 400)
	}

	if len(title) > 200 {
		return apperrors.New("TITLE_TOO_LONG", "Title cannot exceed 200 characters", 400)
	}

	// 检查是否包含特殊字符
	if strings.ContainsAny(title, "<>\"'&") {
		return apperrors.New("INVALID_TITLE_CHARS", "Title contains invalid characters", 400)
	}

	return nil
}

// validateAuthor 验证作者
func (v *BookValidator) validateAuthor(author string) error {
	author = strings.TrimSpace(author)

	if len(author) == 0 {
		return apperrors.New("EMPTY_AUTHOR", "Author cannot be empty", 400)
	}

	if len(author) < 1 {
		return apperrors.New("AUTHOR_TOO_SHORT", "Author must be at least 1 character", 400)
	}

	if len(author) > 100 {
		return apperrors.New("AUTHOR_TOO_LONG", "Author cannot exceed 100 characters", 400)
	}

	// 检查是否包含特殊字符
	if strings.ContainsAny(author, "<>\"'&") {
		return apperrors.New("INVALID_AUTHOR_CHARS", "Author contains invalid characters", 400)
	}

	return nil
}

// ValidateID 验证ID格式
func (v *BookValidator) ValidateID(id string) error {
	if id == "" {
		return apperrors.New("EMPTY_ID", "ID cannot be empty", 400)
	}

	// 简单的数字ID验证
	if len(id) > 20 {
		return apperrors.New("INVALID_ID", "Invalid ID format", 400)
	}

	return nil
}

// GetValidationErrorCode 根据错误类型返回标准错误码
func (v *BookValidator) GetValidationErrorCode(err error) string {
	if appErr, ok := apperrors.IsAppError(err); ok {
		return fmt.Sprintf("BOOK_VALIDATION_%s", appErr.Code)
	}
	return "BOOK_VALIDATION_ERROR"
}
