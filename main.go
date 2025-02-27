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
	// Load .env only in local development
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	}

	// Get PORT from environment (Railway sets this automatically)
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
	}

	// Initialize Gin router
	router := gin.Default()

	// Enable CORS with default settings
	router.Use(cors.Default())

	// Initialize routes
	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	// Start server
	log.Printf("Server running on port %s", PORT)
	log.Fatal(router.Run(":" + PORT))
}
