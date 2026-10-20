package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
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

func (a *AppError) Error() string {
	return a.Message
}
func NewError(message string, code ErrorCode) error {
	return &AppError{
		Message: message,
		code:    code,
	}
}

func WrapError(message string, code ErrorCode, err error) error {
	return &AppError{
		Message: message,
		code:    code,
		Err:     err,
	}
}

func ResponseError(ctx *gin.Context, err error) {
	if appErr, ok := err.(*AppError); ok {
		status := httpStatusFromCode(appErr.code)
		response := gin.H{
			"error":   appErr.code,
			"message": appErr.Message,
		}
		if appErr.Err != nil {
			response["detail"] = appErr.Err.Error()
		}
		ctx.JSON(status, response)
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{
		"error": err.Error(),
		"code":  ErrCodeInternal,
	})
}
func ResponseSuccess(ctx *gin.Context, status int, data any) {
	ctx.JSON(status, gin.H{
		"data":   data,
		"status": "success",
	})
}
func ResponseStatusCode(ctx *gin.Context, status int) {
	ctx.Status(status)
}
func ResponseValidator(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusBadRequest, data)
}
func httpStatusFromCode(code ErrorCode) int {
	switch code {
	case ErrCodeBadRequest:
		return http.StatusBadRequest
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case ErrCodeInternal:
		return http.StatusInternalServerError
	case ErrCodeConflict:
		return http.StatusConflict
	case ErrCodeNotImplemented:
		return http.StatusNotImplemented
	case ErrCodeTooManyRequests:
		return http.StatusTooManyRequests
	}
	return http.StatusInternalServerError
}
