package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"huydq2801.com/golearn/internal/config"
	"huydq2801.com/golearn/internal/routes"
	"huydq2801.com/golearn/internal/validation"
)

type Module interface {
	Routes() routes.Route
}
type Application struct {
	config  *config.Config
	routes  *gin.Engine
	modules []Module
}

func NewApplication(cfg *config.Config) *Application {
	r := gin.Default()
	if err := validation.InitValidator(); err != nil {
		log.Fatal("Failed to initialize validator: ", err)
	}
	loadEnv()
	modules := []Module{
		NewUserModule(),
	}
	routes.RegisterRoutes(r, getModulRoutes(modules)...)
	return &Application{
		routes:  r,
		config:  cfg,
		modules: modules,
	}
}

func (a *Application) Run() error {
	server := &http.Server{
		Addr:    a.config.ServerAddress,
		Handler: a.routes,
	}
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		log.Println("Server is running on ", a.config.ServerAddress)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %s\n", err)
		}
	}()
	<-quit
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %s\n", err)
		return err
	}
	log.Println("Server gracefully stopped")
	return nil
}

func getModulRoutes(modules []Module) []routes.Route {
	routesList := make([]routes.Route, len(modules))
	for i, module := range modules {
		routesList[i] = module.Routes()
	}
	return routesList
}

func loadEnv() {
	if err := godotenv.Load("../../.env"); err != nil {
		log.Println("No .env file found")
	}
}
