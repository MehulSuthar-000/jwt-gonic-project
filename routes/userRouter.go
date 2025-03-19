package routes

import (
	"github.com/gin-gonic/gin"
	"gitub.com/mehulsuthar-000/golang-jwt-project/controller"
	"gitub.com/mehulsuthar-000/golang-jwt-project/middleware"
)

func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authenticate())
	incomingRoutes.GET("/users", controller.GetUsers())
	incomingRoutes.GET("/users/:user_id", controller.GetUser())
}
