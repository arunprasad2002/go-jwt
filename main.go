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
	// Attempt to load .env but do NOT crash if it's missing
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Warning: No .env file found, using system environment variables")
	}

	// Get PORT from environment variables, fallback to 8080
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
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

	// Initialize routes
	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	log.Printf("Server running on port %s", PORT)
	err = router.Run(":" + PORT)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
