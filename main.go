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
	// More robust check for Railway environment
	isRailway := os.Getenv("RAILWAY_ENVIRONMENT") != "" || os.Getenv("RAILWAY_STATIC_URL") != ""

	// Load .env only in local development
	if !isRailway {
		err := godotenv.Load(".env")
		if err != nil {
			log.Println("No .env file found, using system environment variables")
		}
	} else {
		log.Println("Running in Railway environment, using Railway environment variables")
	}

	// Debug: Print important environment variables
	logEnvironmentVariables()

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

	// Configure CORS with more specific settings if needed
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

// Helper function to log important environment variables
func logEnvironmentVariables() {
	log.Printf("PORT: %s", os.Getenv("PORT"))
	log.Printf("RAILWAY_ENVIRONMENT: %s", os.Getenv("RAILWAY_ENVIRONMENT"))
	log.Printf("RAILWAY_STATIC_URL: %s", os.Getenv("RAILWAY_STATIC_URL"))

	// Log other important environment variables your app needs
	// For example:
	// log.Printf("DATABASE_URL: %s", maskSensitiveInfo(os.Getenv("DATABASE_URL")))
	// log.Printf("JWT_SECRET exists: %v", os.Getenv("JWT_SECRET") != "")
}

// Optional: Helper function to mask sensitive information in logs
func maskSensitiveInfo(input string) string {
	if len(input) <= 8 {
		return "****"
	}
	return input[:4] + "..." + input[len(input)-4:]
}
