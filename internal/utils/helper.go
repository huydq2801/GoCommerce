package utils

import "os"

func GetEnv(key string, def string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return def
}
