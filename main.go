package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	routes "gitub.com/mehulsuthar-000/golang-jwt-project/routes"
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

	router.GET("/api/v1", func(ctx *gin.Context) {
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"success": "Access granted for api - v1",
			},
		)
	})

	router.GET("api/v2", func(ctx *gin.Context) {
		ctx.JSON(
			http.StatusOK,
			gin.H{
				"success": "Access granted for api - v2",
			},
		)
	})

	router.Run(":" + port)
}
