package routes

import (
	"github.com/arunprasad2002/go-jwt/controllers"
	"github.com/arunprasad2002/go-jwt/middleware"
	"github.com/gin-gonic/gin"
)

func UserRoutes(router *gin.Engine) {
	router.Use(middleware.Authenticate())
	router.GET("/users", controllers.GetUser())
	router.GET("/users/:user_id", controllers.GetUser())
}
