package routes

import (
	"github.com/gin-gonic/gin"
	"gitub.com/mehulsuthar-000/golang-jwt-project/controller"
)

func AuthRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.POST("users/signup", controller.Signup())
	incomingRoutes.POST("users/login", controller.Login())
}
