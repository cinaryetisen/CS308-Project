package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"medieval-store/config"
	"medieval-store/controllers"
	"medieval-store/models"
	"medieval-store/security"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Helper to set up the checkout router
func setupCheckoutRouter() *gin.Engine {
	os.Setenv("JWT_SECRET", "test_secret")
	gin.SetMode(gin.TestMode)
	router := gin.New()

	protected := router.Group("/api")
	protected.Use(security.AuthMiddleware())
	protected.POST("/checkout", controllers.Checkout)

	return router
}

// Helper to clean up MongoDB products after checkout tests
func clearMongoProducts() {
	if config.MongoClient != nil {
		config.MongoClient.Database("medieval_store").Collection("products").DeleteMany(context.Background(), bson.M{})
	}
}

func TestCheckout_Success(t *testing.T) {
	setupTestDB()
	ensureMongo() // Ensure MongoDB is connected
	defer clearTestDB()
	defer clearMongoProducts()

	// 1. Create a fake user in PostgreSQL
	config.DB.Create(&models.User{ID: 1, Name: "King Arthur", Email: "arthur@camelot.com"})

	// 2. Create a fake product in MongoDB with a stock of 10
	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database("medieval_store").Collection("products")
	collection.InsertOne(context.Background(), models.Product{
		ID:       productID,
		Name:     "Iron Sword",
		Price:    150.50,
		Quantity: 10, // Plenty in stock!
	})

	router := setupCheckoutRouter()

	// 3. Create the mock payload buying 2 swords
	checkoutData := map[string]interface{}{
		"shipping_address": "123 Camelot Castle",
		"total_price":      301.00,
		"cart_items": []models.CartItem{
			{ProductID: productID.Hex(), Quantity: 2},
		},
	}
	jsonValue, _ := json.Marshal(checkoutData)

	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 4. Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Order placed successfully")

	// 5. Verify PostgreSQL saved the real price from MongoDB
	var orderItem models.OrderItem
	config.DB.First(&orderItem)
	assert.Equal(t, 150.50, orderItem.Price) // Proves the backend ignored the frontend and fetched the real price!

	// 6. Verify MongoDB stock went down from 10 to 8!
	var updatedProduct models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&updatedProduct)
	assert.Equal(t, 8, updatedProduct.Quantity)
}

func TestCheckout_OutOfStock(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoProducts()

	config.DB.Create(&models.User{ID: 1, Name: "Lancelot"})

	// Create a product with ONLY 1 left in stock
	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database("medieval_store").Collection("products")
	collection.InsertOne(context.Background(), models.Product{
		ID:       productID,
		Name:     "Holy Grail",
		Price:    999.99,
		Quantity: 1,
	})

	router := setupCheckoutRouter()

	// The user attempts to buy 2 of them!
	checkoutData := map[string]interface{}{
		"shipping_address": "Avalon",
		"total_price":      1999.98,
		"cart_items": []models.CartItem{
			{ProductID: productID.Hex(), Quantity: 2},
		},
	}
	jsonValue, _ := json.Marshal(checkoutData)

	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert it was BLOCKED with a 409 Conflict
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "out of stock")

	// Verify PostgreSQL Transaction Rollback worked (No ghost orders created)
	var orderCount int64
	config.DB.Model(&models.Order{}).Count(&orderCount)
	assert.Equal(t, int64(0), orderCount)

	// Verify MongoDB stock was untouched
	var unchangedProduct models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&unchangedProduct)
	assert.Equal(t, 1, unchangedProduct.Quantity)
}

func TestCheckout_MissingRequiredFields(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupCheckoutRouter()

	checkoutData := map[string]interface{}{
		"total_price": 100.00,
	}
	jsonValue, _ := json.Marshal(checkoutData)

	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCheckout_Unauthorized(t *testing.T) {
	router := setupCheckoutRouter()

	checkoutData := map[string]interface{}{
		"shipping_address": "Nowhere",
		"total_price":      50.00,
		"cart_items":       []models.CartItem{},
	}
	jsonValue, _ := json.Marshal(checkoutData)

	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
