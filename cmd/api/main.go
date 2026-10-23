package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	"huydq2801.com/golearn/internal/app"
	"huydq2801.com/golearn/internal/config"
	"huydq2801.com/golearn/internal/db"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	if err := run(&logger); err != nil {
		logger.Error().Err(err).Msg("application_failed")
		os.Exit(1)
	}
}
func run(logger *zerolog.Logger) error {
	// Run from repository root. Real environment variables take precedence over .env.
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return errors.New("could not load .env file")
	}
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	gin.SetMode(gin.ReleaseMode)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	pool, err := db.Open(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()
	application, err := app.NewApplication(cfg, pool, logger)
	if err != nil {
		return err
	}
	return application.Run(ctx)
}
