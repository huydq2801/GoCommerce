package app

import (
	"log"

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
	config *config.Config
	routes *gin.Engine
	modules []Module
}

func NewApplication(cfg *config.Config) *Application {
	r := gin.Default()
	if err := validation.InitValidator(); err != nil{
		log.Fatal("Failed to initialize validator: ", err)
	}
	loadEnv()
	modules := []Module{
		NewUserModule(),
	}
	routes.RegisterRoutes(r, getModulRoutes(modules)...)
	return &Application{
		routes: r,
		config: cfg,
		modules: modules,
	}
}

func (a *Application) Run() error {
	return a.routes.Run(a.config.ServerAddress)
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
