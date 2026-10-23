package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"gocommerce/internal/db/sqlc"
	"gocommerce/internal/utils"
)

func TestLimiterRejectsEveryExcessRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	logger := zerolog.New(io.Discard)
	r := gin.New()
	r.Use(RateLimiterMiddleware(&logger, 1, 1))
	calls := 0
	r.GET("/", func(c *gin.Context) { calls++; c.Status(204) })
	for i := 0; i < 6; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
		want := 429
		if i == 0 {
			want = 204
		}
		if w.Code != want {
			t.Fatalf("request %d: got %d want %d", i, w.Code, want)
		}
	}
	if calls != 1 {
		t.Fatalf("handler called %d times", calls)
	}
}
func TestLogsExcludeSensitiveRequestData(t *testing.T) {
	var output bytes.Buffer
	logger := zerolog.New(&output)
	r := gin.New()
	r.Use(RequestID(), LoggerMiddleware(&logger), RecoveryMiddleware(&logger))
	r.POST("/", func(c *gin.Context) { c.JSON(201, gin.H{"data": "ok"}) })
	req := httptest.NewRequest("POST", "/?token=query-secret", strings.NewReader(`{"password":"body-secret"}`))
	req.Header.Set("Authorization", "Bearer bearer-secret")
	req.Header.Set("X-API-Key", "key-secret")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var event map[string]any
	if err := json.Unmarshal(output.Bytes(), &event); err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"query-secret", "body-secret", "bearer-secret", "key-secret"} {
		if strings.Contains(output.String(), secret) {
			t.Fatalf("log exposed %s", secret)
		}
	}
	if event["request_id"] == "" || w.Header().Get("X-Request-ID") == "" {
		t.Fatal("missing correlation ID")
	}
}
func TestRecoveryDoesNotExposePanic(t *testing.T) {
	var output bytes.Buffer
	logger := zerolog.New(&output)
	r := gin.New()
	r.Use(RequestID(), LoggerMiddleware(&logger), RecoveryMiddleware(&logger))
	r.GET("/", func(c *gin.Context) { panic("private-panic") })
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != 500 || strings.Contains(w.Body.String()+output.String(), "private-panic") {
		t.Fatalf("unsafe recovery: %d %s", w.Code, w.Body)
	}
}

type lookupStub struct{ user sqlc.User }

func (s lookupStub) GetUser(context.Context, uuid.UUID) (sqlc.User, error) { return s.user, nil }
func TestJWTAuthenticationAndAdmin(t *testing.T) {
	secret := strings.Repeat("s", 32)
	user := sqlc.User{Uuid: uuid.New(), UserStatus: 1, UserLevel: 3}
	auth := NewAuthenticator(secret, lookupStub{user: user})
	r := gin.New()
	r.GET("/private", auth.RequireUser(), func(c *gin.Context) { c.Status(204) })
	r.GET("/admin", auth.RequireAdmin(strings.Repeat("a", 32)), func(c *gin.Context) { c.Status(204) })
	sign := func(exp time.Time, key string) string {
		claims := jwt.RegisteredClaims{Issuer: "gocommerce", Audience: jwt.ClaimStrings{"gocommerce-api"}, Subject: user.Uuid.String(), ExpiresAt: jwt.NewNumericDate(exp)}
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(key))
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	tests := []struct {
		path, token string
		want        int
	}{{"/private", "", 401}, {"/private", sign(time.Now().Add(time.Hour), secret), 204}, {"/private", sign(time.Now().Add(-time.Hour), secret), 401}, {"/private", sign(time.Now().Add(time.Hour), "wrong-key"), 401}, {"/admin", sign(time.Now().Add(time.Hour), secret), 403}}
	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		if tt.token != "" {
			req.Header.Set("Authorization", "Bearer "+tt.token)
		}
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tt.want {
			t.Errorf("%s status=%d want %d", tt.path, w.Code, tt.want)
		}
	}
}

func TestOversizedBodyRejectedBeforeHandlerWork(t *testing.T) {
	logger := zerolog.New(io.Discard)
	r := gin.New()
	r.Use(RequestID(), RecoveryMiddleware(&logger))
	r.POST("/", func(c *gin.Context) {
		var input struct {
			Name string `json:"name"`
		}
		if !utils.DecodeJSON(c, &input) {
			return
		}
		c.Status(204)
	})
	// Valid JSON followed by oversized whitespace must still produce 413, not 400.
	req := httptest.NewRequest("POST", "/", strings.NewReader(`{"name":"ok"}`+strings.Repeat(" ", 1<<20)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 413 {
		t.Fatalf("status=%d want 413 body=%s", w.Code, w.Body)
	}
}
