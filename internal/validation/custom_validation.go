package validation

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"huydq2801.com/golearn/internal/utils"
)

func RegisterCustomValidation(v *validator.Validate) {
	var blockedDomain = map[string]bool{
		"mailinator.com":   true,
		"tempmail.com":     true,
		"10minutemail.com": true,
		"blacklist.com":    true,
		"spam.com":         true,
		"random.com":       true,
	}

	v.RegisterValidation("email_advanced", func(fl validator.FieldLevel) bool {
		email := strings.ToLower(fl.Field().String())
		parts := strings.Split(email, "@")
		if len(parts) != 2 {
			return false
		}
		domain := utils.NormalizeString(parts[1])
		return !blockedDomain[domain]
	})
	v.RegisterValidation("password_strong", func(fl validator.FieldLevel) bool {
		password := fl.Field().String()
		if len(password) < 8 {
			return false
		}
		hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
		hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
		hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
		hasSpecial := regexp.MustCompile(`[!@#$%^&*()]`).MatchString(password)
		return hasUpper && hasLower && hasNumber && hasSpecial
	})
	var slugRegex = regexp.MustCompile(`^[a-z0-9]+(?:[-.][a-z0-9]+)*$`)
	v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		return slugRegex.MatchString(fl.Field().String())
	})
	var searchRegex = regexp.MustCompile(`^[a-zA-Z0-9\s]*$`)
	v.RegisterValidation("search", func(fl validator.FieldLevel) bool {
		return searchRegex.MatchString(fl.Field().String())
	})
	v.RegisterValidation("min_int", func(fl validator.FieldLevel) bool {
		minStr := fl.Param()
		minVal, err := strconv.ParseInt(minStr, 10, 64)
		if err != nil {
			return false
		}
		return fl.Field().Int() >= minVal
	})
	v.RegisterValidation("max_int", func(fl validator.FieldLevel) bool {
		maxStr := fl.Param()
		maxVal, err := strconv.ParseInt(maxStr, 10, 64)
		if err != nil {
			return false
		}
		return fl.Field().Int() <= maxVal
	})
	v.RegisterValidation("file_extention", func(fl validator.FieldLevel) bool {
		filename := fl.Field().String()
		allowedStr := fl.Param()
		if allowedStr == "" {
			return false
		}
		allowedExt := strings.Fields(allowedStr)
		ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(filename)), ".")
		for _, allowed := range allowedExt {
			if ext == strings.ToLower(allowed) {
				return true
			}
		}
		return false
	})
}
