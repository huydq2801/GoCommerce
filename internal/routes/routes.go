package routes

import "github.com/gin-gonic/gin"

type Route interface{ RegisterRoutes(r *gin.RouterGroup) }

func RegisterRoutes(r *gin.Engine, modules ...Route) {
	v1 := r.Group("/api/v1")
	for _, module := range modules {
		module.RegisterRoutes(v1)
	}
}
