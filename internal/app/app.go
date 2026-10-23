package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	"huydq2801.com/golearn/internal/apidocs"
	"huydq2801.com/golearn/internal/config"
	"huydq2801.com/golearn/internal/db/sqlc"
	v1handler "huydq2801.com/golearn/internal/handler/v1"
	"huydq2801.com/golearn/internal/middleware"
	"huydq2801.com/golearn/internal/repository"
	"huydq2801.com/golearn/internal/routes"
	v1routes "huydq2801.com/golearn/internal/routes/v1"
	v1service "huydq2801.com/golearn/internal/service/v1"
	"huydq2801.com/golearn/internal/validation"
)

type Module interface{ Routes() routes.Route }
type Application struct {
	config *config.Config
	routes *gin.Engine
	logger *zerolog.Logger
}
type ModuleContext struct {
	Auth *middleware.Authenticator

	DB     sqlc.Querier
	APIKey string
}

func NewApplication(cfg *config.Config, pool *pgxpool.Pool, logger *zerolog.Logger) (*Application, error) {
	if err := validation.InitValidator(); err != nil {
		return nil, err
	}
	r := gin.New()
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, err
	}
	r.Use(middleware.RequestID(), middleware.LoggerMiddleware(logger), middleware.RecoveryMiddleware(logger))
	r.GET("/health/live", func(c *gin.Context) { c.JSON(200, gin.H{"data": gin.H{"status": "ok"}}) })
	r.GET("/health/ready", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), time.Second)
		defer cancel()
		if err := pool.Ping(ctx); err != nil {
			c.JSON(503, gin.H{"error": gin.H{"code": "NOT_READY", "message": "Database unavailable"}})
			return
		}
		c.JSON(200, gin.H{"data": gin.H{"status": "ok"}})
	})
	r.GET("/openapi.json", func(c *gin.Context) { c.Data(200, "application/json", apidocs.Spec) })
	r.Use(middleware.RateLimiterMiddleware(logger, cfg.RatePerSecond, cfg.RateBurst))
	queries := sqlc.New(pool)
	auth := middleware.NewAuthenticator(cfg.JWTSecret, queries)
	modules := &ModuleContext{DB: queries, APIKey: cfg.APIKey, Auth: auth}
	authHandler := v1handler.NewAuthHandler(v1service.NewAuthService(queries, cfg.JWTSecret))
	userHandler := v1handler.NewUserHandler(v1service.NewUserService(repository.NewSQLUserRepository(queries)))
	r.POST("/api/v1/auth/register", userHandler.CreateUser)
	r.POST("/api/v1/auth/login", authHandler.Login)
	r.GET("/api/v1/users/me", auth.RequireUser(), authHandler.Me)
	shoppingHandler := v1handler.NewShoppingHandler(v1service.NewShoppingService(repository.NewShoppingStore(pool)))
	routes.RegisterRoutes(r, NewUserModule(modules).Routes(), NewCatalogModule(modules).Routes(), v1routes.NewShoppingRoutes(shoppingHandler, auth, cfg.APIKey))
	r.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{"error": gin.H{"code": "NOT_FOUND", "message": "Resource not found"}})
	})
	return &Application{config: cfg, routes: r, logger: logger}, nil
}
func (a *Application) Handler() http.Handler { return a.routes }
func (a *Application) Run(ctx context.Context) error {
	server := &http.Server{Addr: a.config.ServerAddress, Handler: a.routes, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 15 * time.Second, IdleTimeout: 60 * time.Second}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return fmt.Errorf("listen HTTP: %w", err)
	}
	a.logger.Info().Str("address", listener.Addr().String()).Msg("server_started")
	return serve(ctx, server, listener, a.config.ShutdownTimeout, a.logger)
}
func serve(ctx context.Context, server *http.Server, listener net.Listener, timeout time.Duration, logger *zerolog.Logger) error {
	serverErrors := make(chan error, 1)
	go func() { serverErrors <- server.Serve(listener) }()
	select {
	case err := <-serverErrors:
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		logger.Info().Msg("server_stopping")
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		_ = server.Close()
		return fmt.Errorf("shutdown HTTP: %w", err)
	}
	if err := <-serverErrors; err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	logger.Info().Msg("server_stopped")
	return nil
}
