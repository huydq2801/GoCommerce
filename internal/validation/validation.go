package validation

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"huydq2801.com/golearn/internal/utils"
)

func InitValidator() error {
	v, ok := binding.Validator.Engine().(*validator.Validate)
	if !ok {
		return fmt.Errorf("Failed to get validator engine")
	}
	RegisterCustomValidation(v)
	return nil
}

func HandleValidationError(err error) gin.H {
	if ValidationError, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]string)
		for _, e := range ValidationError {
			root := strings.Split(e.Namespace(), ".")[0]
			rawPath := strings.TrimPrefix(e.Namespace(), root+".")
			parts := strings.Split(rawPath, ".")
			for i, part := range parts {
				if strings.Contains(part, "[") {
					idx := strings.Index(part, "[")
					base := utils.CamelToSnake(part[:idx])
					index := part[idx:]
					parts[i] = base + index
				} else {
					parts[i] = utils.CamelToSnake(part)
				}
			}
			fieldPath := strings.Join(parts, ".")
			switch e.Tag() {
			case "eq":
				errors[fieldPath] = e.Field() + " phải bằng " + e.Param()
			case "eq_ignore_case":
				errors[fieldPath] = e.Field() + " phải bằng " + e.Param() + " không phân biệt chữ hoa chữ thường"
			case "gt":
				errors[fieldPath] = e.Field() + " phải lớn hơn " + e.Param()
			case "gte":
				errors[fieldPath] = e.Field() + " phải lớn hơn hoặc bằng " + e.Param()
			case "lt":
				errors[fieldPath] = e.Field() + " phải nhỏ hơn " + e.Param()
			case "lte":
				errors[fieldPath] = e.Field() + " phải nhỏ hơn hoặc bằng " + e.Param()
			case "ne":
				errors[fieldPath] = e.Field() + " phải khác " + e.Param()
			case "ne_ignore_case":
				errors[fieldPath] = e.Field() + " phải khác " + e.Param() + " không phân biệt chữ hoa chữ thường"
			case "uuid":
				errors[fieldPath] = e.Field() + " phải là một UUID hợp lệ"
			case "slug":
				errors[fieldPath] = e.Field() + " phải là một slug hợp lệ"
			case "required":
				errors[fieldPath] = e.Field() + " là bắt buộc"
			case "min":
				errors[fieldPath] = e.Field() + " phải có độ dài tối thiểu là " + e.Param()
			case "max":
				errors[fieldPath] = e.Field() + " phải có độ dài tối đa là " + e.Param()
			case "oneof":
				errors[fieldPath] = e.Field() + " phải là một trong các giá trị sau: " + e.Param()
			case "omitempty":
				errors[fieldPath] = e.Field() + " là tùy chọn"
			case "search":
				errors[fieldPath] = e.Field() + " chỉ được chứa các ký tự chữ cái, số và khoảng trắng"
			case "email":
				errors[fieldPath] = e.Field() + " phải là một email hợp lệ"
			case "datetime":
				errors[fieldPath] = e.Field() + " phải theo đúng định dạng dd-mm-yyyy"
			case "min_int":
				errors[fieldPath] = e.Field() + " phải có giá trị tối thiểu là " + e.Param()
			case "max_int":
				errors[fieldPath] = e.Field() + " phải có giá trị tối đa là " + e.Param()
			case "file_extention":
				errors[fieldPath] = e.Field() + " phải có đuôi mở rộng là " + e.Param()
			case "email_advanced":
				errors[fieldPath] = e.Field() + " không được chứa domain bị chặn"
			case "password_strong":
				errors[fieldPath] = e.Field() + " không đáp ứng yêu cầu mật khẩu mạnh(ít nhất 8 kí tự, có chữ hoa, chữ thường, số và kí tự đặc biệt)"
			}
		}
		return gin.H{"errors": errors}
	}
	return gin.H{
		"error":  "Yêu cầu không hợp lệ ",
		"detail": err.Error(),
	}
}
