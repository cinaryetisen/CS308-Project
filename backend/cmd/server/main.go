package main

import (
	"log"
	"os"

	"medieval-store/config"
	"medieval-store/routes"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// 1. Loading environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env found, relying on system environment variables instead")
	}

	// 2. Connect to PostgreSQL database
	if err := config.ConnectPostgres(); err != nil {
		log.Fatalf("Could not initialize Postgres: %v", err)
	}

	// 3. Connect to MongoDB database
	if err := config.ConnectMongo(); err != nil {
		log.Fatalf("Mongo initializaiton failed: %v", err)
	}

	// 4. Set up router
	router := routes.SetupRouter()

	// 5. CORS setup so your React frontend is allowed to talk to this backend
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// 6. Health check for API
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Medieval Store API is up and running",
		})
	})

	// 8. Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...", port)
	router.Run(":" + port)
}
