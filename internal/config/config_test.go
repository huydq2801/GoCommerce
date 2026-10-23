package config

import (
	"net/url"
	"strings"
	"testing"
)

func baseEnv(t *testing.T) {
	t.Setenv("TRUSTED_PROXIES", "")
	t.Helper()
	for _, key := range []string{"SERVER_PORT", "DB_PORT", "DB_SSLMODE", "RATE_LIMITER_REQUEST_SEC", "RATE_LIMITER_REQUEST_BURST", "SHUTDOWN_TIMEOUT"} {
		t.Setenv(key, "")
	}
	t.Setenv("API_KEY", strings.Repeat("a", 32))
	t.Setenv("JWT_SECRET", strings.Repeat("s", 32))
	t.Setenv("DB_PASSWORD", "test password@:/")
}
func TestLoadConfig(t *testing.T) {
	baseEnv(t)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ServerAddress != ":8080" {
		t.Fatal(cfg.ServerAddress)
	}
	dsn, err := url.Parse(cfg.DSN())
	if err != nil {
		t.Fatal(err)
	}
	password, _ := dsn.User.Password()
	if password != "test password@:/" {
		t.Fatal("password was not URL encoded")
	}
}
func TestInvalidConfig(t *testing.T) {
	for _, tt := range []struct{ key, value string }{{"SERVER_PORT", "0"}, {"SERVER_PORT", "abc"}, {"DB_PORT", "65536"}, {"API_KEY", "short"}, {"JWT_SECRET", "short"}, {"DB_PASSWORD", ""}, {"RATE_LIMITER_REQUEST_SEC", "-1"}, {"RATE_LIMITER_REQUEST_BURST", "0"}, {"SHUTDOWN_TIMEOUT", "-1s"}, {"DB_SSLMODE", "wrong"}} {
		t.Run(tt.key+tt.value, func(t *testing.T) {
			baseEnv(t)
			t.Setenv(tt.key, tt.value)
			if _, err := Load(); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
}
