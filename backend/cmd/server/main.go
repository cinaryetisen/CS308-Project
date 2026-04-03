package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// 1. Load the .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// 2. Access the variables using os.Getenv
	postgresDSN := os.Getenv("POSTGRES_DSN")
	mongoURI := os.Getenv("MONGO_URI")
	port := os.Getenv("PORT")

	// Tell Go it's okay that we aren't using these just yet
	_ = postgresDSN
	_ = mongoURI

	// If PORT is not set in .env, default to 8080
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server will run on port: %s\n", port)
	// fmt.Println("Postgres DSN:", postgresDSN) // Do not print this in production!
}
