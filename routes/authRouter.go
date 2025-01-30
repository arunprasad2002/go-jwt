package routes

import (
	"github.com/arunprasad2002/go-jwt/controllers"
	"github.com/gin-gonic/gin"
)

func AuthRoutes(router *gin.Engine) {
	router.POST("/users/signup", controllers.SignUp())
	router.POST("/users/login", controllers.Login())
}
