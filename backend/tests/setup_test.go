package tests

import (
	"log"
	"medieval-store/config"
	"medieval-store/models"

	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates a temporary, in-memory SQLite database just for testing
func setupTestDB() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: Could not find ../.env file. Relying on system environment variables.")
	}
	// "file::memory:?cache=shared" tells SQLite to run entirely in RAM
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to the test database!")
	}

	// Migrate your actual models into this fake database
	db.AutoMigrate(&models.User{}, &models.Order{}, &models.OrderItem{}, &models.CartItem{})

	// Temporarily override your real PostgreSQL connection with this fake one
	config.DB = db
}

// clearTestDB wipes the fake database clean after every single test
func clearTestDB() {
	config.DB.Exec("DELETE FROM users")
	config.DB.Exec("DELETE FROM orders")
	config.DB.Exec("DELETE FROM order_items")
	config.DB.Exec("DELETE FROM cart_items")
}
