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
	//Loading environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env found, relying on system environment variables instead")
	}

	//Connect to PostgreSQL database
	if err := config.ConnectPostgres(); err != nil {
		log.Fatalf("Could not initialize Postgres: %v", err)
	}

	//Connect to MongoDB database
	if err := config.ConnectMongo(); err != nil {
		log.Fatalf("Mongo initializaiton failed: %v", err)
	}

	//Set up router
	router := routes.SetupRouter()

	//Health check for API
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Medieval Store API is up and running",
		})
	})

	//Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting server on port %s...", port)
	router.Run(":" + port)
}
