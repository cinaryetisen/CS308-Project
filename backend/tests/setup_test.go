package tests

import (
	"context"
	"log"
	"medieval-store/config"
	"medieval-store/models"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB creates a temporary, in-memory SQLite database just for testing
func setupTestDB() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("Warning: Could not find ../.env file. Relying on system environment variables.")
	}
	// Tests need a fixed DATA_ENC_KEY so the F1 encryption hooks work.
	// Production sets this in .env; here we fall back to an all-zero 32-byte key.
	if os.Getenv("DATA_ENC_KEY") == "" {
		os.Setenv("DATA_ENC_KEY", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
	}
	// "file::memory:?cache=shared" tells SQLite to run entirely in RAM
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to the test database!")
	}

	// Migrate your actual models into this fake database
	db.AutoMigrate(&models.User{}, &models.Order{}, &models.OrderItem{}, &models.CartItem{}, &models.WishlistItem{}, &models.Refund{})

	// Temporarily override your real PostgreSQL connection with this fake one
	config.DB = db
	config.MongoDBName = "medieval_store_test"
}

func clearMongoCollection(name string) {
	if config.MongoDBName == "medieval_store" {
		panic("test attempted to clear the production database; check setup_test.go")
	}
	if config.MongoClient != nil {
		config.MongoClient.Database(config.MongoDBName).Collection(name).DeleteMany(context.Background(), bson.M{})
	}
}

// clearTestDB wipes the fake database clean after every single test
func clearTestDB() {
	config.DB.Exec("DELETE FROM users")
	config.DB.Exec("DELETE FROM orders")
	config.DB.Exec("DELETE FROM order_items")
	config.DB.Exec("DELETE FROM cart_items")
	config.DB.Exec("DELETE FROM wishlist_items")
	config.DB.Exec("DELETE FROM refunds")
}
