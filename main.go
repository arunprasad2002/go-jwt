package main

import (
	"log"
	"os"

	"github.com/arunprasad2002/go-jwt/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	PORT := os.Getenv("PORT")
	router := gin.New()
	router.Use(gin.Logger())
	routes.AuthRoutes(router)
	router.Run(":" + PORT)
}
