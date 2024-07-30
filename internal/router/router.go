package router

import (
	"aidashboard/internal/controller"
	"aidashboard/pkg/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(middleware.LoggingMiddleware())

	apiGroup := r.Group("/api")
	{
		apiGroup.POST("/ai-dashboard", controller.GetResponse)
	}

	return r
}
