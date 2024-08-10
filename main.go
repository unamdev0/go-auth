package main

import (
	"os"

	"github.com/gin-gonic/gin"
	routes "github.com/unamdev0/go-auth/routes"
)

func main() {

	port := os.Getenv("PORT")

	if port == "" {
		port = "8000"
	}

	router := gin.New()
	router.Use(gin.Logger())

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	router.Run(":" + port)

}
