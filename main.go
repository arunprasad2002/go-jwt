package main

import (
	"log"
	"os"

	"github.com/arunprasad2002/go-jwt/routes"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080" // Default to port 8080 if not set
	}

	router := gin.New()
	router.Use(gin.Logger())

	// Enable CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // Change to specific origins for security
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	log.Printf("Server running on port %s", PORT)
	router.Run(":" + PORT)
}
