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
	// Check if running on Railway
	isRailway := os.Getenv("RAILWAY_ENVIRONMENT") != "" || os.Getenv("RAILWAY_STATIC_URL") != ""

	// Load .env in local development
	if !isRailway {
		err := godotenv.Load(".env")
		if err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	} else {
		log.Println("Running in Railway environment, using Railway variables")
	}

	// Get PORT from environment (Railway sets this automatically)
	PORT := os.Getenv("PORT")
	if PORT == "" {
		PORT = "8080"
		log.Println("PORT not found in environment, defaulting to 8080")
	}

	// Set Gin mode based on environment
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize Gin router
	router := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowCredentials = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(config))

	// Initialize routes
	routes.AuthRoutes(router)
	routes.UserRoutes(router)

	// Start server
	log.Printf("Server running on port %s", PORT)
	log.Fatal(router.Run(":" + PORT))
}
