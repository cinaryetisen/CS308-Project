package routes

import (
	"medieval-store/controllers"

	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	router := gin.Default()

	api := router.Group("/api")

	api.POST("/signup", controllers.Signup)
	api.POST("/login", controllers.Login)

	return router
}
