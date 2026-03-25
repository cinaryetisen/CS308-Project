package config

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectPostgres() error {
	dsn := os.Getenv("POSTGRES_DSN")

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	//Failed to connect to server
	if err != nil {
		log.Printf("Failed to connect to PostgreSQL: %v\n", err)
		return err
	}

	DB = database
	log.Println("Successfully connected to PostgreSQL!")
	return nil
}
