package utils

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ErrorCode string

const (
	ErrCodeBadRequest      ErrorCode = "BAD_REQUEST"
	ErrCodeNotFound        ErrorCode = "NOT_FOUND"
	ErrCodeUnauthorized    ErrorCode = "UNAUTHORIZED"
	ErrCodeInternal        ErrorCode = "INTERNAL_SERVER_ERROR"
	ErrCodeConflict        ErrorCode = "CONFLICT"
	ErrCodeNotImplemented  ErrorCode = "NOT_IMPLEMENTED"
	ErrCodeTooManyRequests ErrorCode = "TOO_MANY_REQUESTS"
)

type AppError struct {
	code    ErrorCode
	Message string
	Err     error
}

func (a *AppError) Error() string                   { return a.Message }
func (a *AppError) Unwrap() error                   { return a.Err }
func NewError(message string, code ErrorCode) error { return &AppError{code: code, Message: message} }
func WrapError(message string, code ErrorCode, err error) error {
	return &AppError{code: code, Message: message, Err: err}
}

// DatabaseError maps storage failures without exposing SQL, constraint names, or credentials.
func DatabaseError(err error) error {
	var typed *AppError
	if errors.As(err, &typed) {
		return err
	}
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return NewError("Resource not found", ErrCodeNotFound)
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return WrapError("A resource with these unique fields already exists", ErrCodeConflict, err)
		case "23503", "23001":
			return WrapError("Referenced resource is missing or still in use", ErrCodeConflict, err)
		case "23514", "22001", "22003":
			return WrapError("Input violates a data constraint", ErrCodeBadRequest, err)
		}
	}
	return WrapError("Internal server error", ErrCodeInternal, err)
}
func ResponseError(c *gin.Context, err error) {
	code, message := ErrCodeInternal, "Internal server error"
	var appErr *AppError
	if errors.As(err, &appErr) {
		code, message = appErr.code, appErr.Message
	}
	c.JSON(httpStatusFromCode(code), gin.H{"error": gin.H{"code": code, "message": message}, "request_id": c.GetString("request_id")})
}
func ResponseSuccess(c *gin.Context, status int, data any) { c.JSON(status, gin.H{"data": data}) }
func ResponseStatusCode(c *gin.Context, status int)        { c.Status(status) }
func ResponseValidator(c *gin.Context, data any) {
	c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": ErrCodeBadRequest, "message": "Invalid request", "details": data}, "request_id": c.GetString("request_id")})
}
func httpStatusFromCode(code ErrorCode) int {
	switch code {
	case ErrCodeBadRequest:
		return 400
	case ErrCodeNotFound:
		return 404
	case ErrCodeUnauthorized:
		return 401
	case ErrCodeConflict:
		return 409
	case ErrCodeTooManyRequests:
		return 429
	case ErrCodeNotImplemented:
		return 501
	default:
		return 500
	}
}

type Page struct {
	Page   int    `form:"page" binding:"omitempty,min=1,max=1000000"`
	Limit  int    `form:"limit" binding:"omitempty,min=1,max=100"`
	Search string `form:"search" binding:"max=100"`
}

func (p *Page) Defaults() {
	if p.Page == 0 {
		p.Page = 1
	}
	if p.Limit == 0 {
		p.Limit = 20
	}
}
func (p Page) Offset() int32 { return int32((p.Page - 1) * p.Limit) }
func ResponsePage(c *gin.Context, data any, p Page, total int64) {
	c.JSON(200, gin.H{"data": data, "meta": gin.H{"page": p.Page, "limit": p.Limit, "total": total, "total_pages": (total + int64(p.Limit) - 1) / int64(p.Limit)}})
}
