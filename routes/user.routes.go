package routes

import (
	"github.com/gin-gonic/gin"
	controllers "github.com/unamdev0/go-auth/controllers"
	middlewares "github.com/unamdev0/go-auth/middlewares"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middlewares.Authenticate())
	incomingRoutes.GET("/users", controllers.GetUsers())
	incomingRoutes.GET("/users/:id", controllers.GetUser())

}
