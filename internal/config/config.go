package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type DataBaseConfig struct{ Host, Port, User, Password, DBName, SSLMode string }
type Config struct {
	TrustedProxies []string

	JWTSecret string

	ServerAddress   string
	DB              DataBaseConfig
	APIKey          string
	RatePerSecond   int
	RateBurst       int
	ShutdownTimeout time.Duration
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
func Load() (*Config, error) {
	port := env("SERVER_PORT", "8080")
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
		return nil, fmt.Errorf("SERVER_PORT must be between 1 and 65535")
	}
	cfg := &Config{ServerAddress: ":" + port, APIKey: os.Getenv("API_KEY"), ShutdownTimeout: 10 * time.Second,
		DB: DataBaseConfig{Host: env("DB_HOST", "localhost"), Port: env("DB_PORT", "5432"), User: env("DB_USER", "postgres"), Password: os.Getenv("DB_PASSWORD"), DBName: env("DB_NAME", "GoCommerce"), SSLMode: env("DB_SSLMODE", "disable")}}
	cfg.JWTSecret = os.Getenv("JWT_SECRET")
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must contain at least 32 characters")
	}
	if len(cfg.APIKey) < 32 {
		return nil, fmt.Errorf("API_KEY must contain at least 32 characters (temporary admin credential)")
	}
	if cfg.DB.Password == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}
	if n, err := strconv.Atoi(cfg.DB.Port); err != nil || n < 1 || n > 65535 {
		return nil, fmt.Errorf("DB_PORT must be between 1 and 65535")
	}
	var err error
	cfg.RatePerSecond, err = strconv.Atoi(env("RATE_LIMITER_REQUEST_SEC", "5"))
	if err != nil || cfg.RatePerSecond < 1 || cfg.RatePerSecond > 10000 {
		return nil, fmt.Errorf("RATE_LIMITER_REQUEST_SEC must be between 1 and 10000")
	}
	cfg.RateBurst, err = strconv.Atoi(env("RATE_LIMITER_REQUEST_BURST", "10"))
	if err != nil || cfg.RateBurst < 1 || cfg.RateBurst > 10000 {
		return nil, fmt.Errorf("RATE_LIMITER_REQUEST_BURST must be between 1 and 10000")
	}
	cfg.ShutdownTimeout, err = time.ParseDuration(env("SHUTDOWN_TIMEOUT", "10s"))
	if err != nil || cfg.ShutdownTimeout <= 0 {
		return nil, fmt.Errorf("SHUTDOWN_TIMEOUT must be a positive duration")
	}
	if !strings.Contains(" disable allow prefer require verify-ca verify-full ", " "+cfg.DB.SSLMode+" ") {
		return nil, fmt.Errorf("invalid DB_SSLMODE")
	}
	if raw := strings.TrimSpace(os.Getenv("TRUSTED_PROXIES")); raw != "" {
		for _, proxy := range strings.Split(raw, ",") {
			proxy = strings.TrimSpace(proxy)
			_, _, cidrErr := net.ParseCIDR(proxy)
			if net.ParseIP(proxy) == nil && cidrErr != nil {
				return nil, fmt.Errorf("invalid TRUSTED_PROXIES address")
			}
			if proxy == "0.0.0.0/0" || proxy == "::/0" {
				return nil, fmt.Errorf("TRUSTED_PROXIES must not trust the entire internet")
			}
			cfg.TrustedProxies = append(cfg.TrustedProxies, proxy)
		}
	}
	return cfg, nil
}
func (c Config) DSN() string {
	u := url.URL{Scheme: "postgresql", User: url.UserPassword(c.DB.User, c.DB.Password), Host: c.DB.Host + ":" + c.DB.Port, Path: "/" + c.DB.DBName}
	q := u.Query()
	q.Set("sslmode", c.DB.SSLMode)
	u.RawQuery = q.Encode()
	return u.String()
}
