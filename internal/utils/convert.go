package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	matchFirstCap = regexp.MustCompile(`(.)([A-Z][a-z]+)`)
	matchAllCap   = regexp.MustCompile(`([a-z0-9])([A-Z])`)
)

func CamelToSnake(str string) string {
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func FormatFileSize(size int64) string {
	switch {
	case size < 1024:
		return fmt.Sprintf("%d B", size)
	case size < 1024*1024:
		return fmt.Sprintf("%.2f KB", float64(size)/1024)
	case size < 1024*1024*1024:
		return fmt.Sprintf("%.2f MB", float64(size)/(1024*1024))
	default:
		return fmt.Sprintf("%.2f GB", float64(size)/(1024*1024*1024))
	}
}
func NormalizeString(text string) string {
	return strings.ToLower(strings.TrimSpace(text))
}
