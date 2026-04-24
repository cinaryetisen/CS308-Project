package tests

import (
	"bytes"
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

func TestCheckout_Success(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	// 1. Create a fake user in the DB
	config.DB.Create(&models.User{ID: 1, Name: "King Arthur", Email: "arthur@camelot.com"})
	router := setupCheckoutRouter()

	// 2. Create the mock payload the frontend would send
	checkoutData := map[string]interface{}{
		"shipping_address": "123 Camelot Castle",
		"total_price":      150.50,
		"cart_items": []models.CartItem{
			{ProductID: "sword_123", Quantity: 1},
		},
	}
	jsonValue, _ := json.Marshal(checkoutData)

	// 3. Send the request
	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer")) // Using the helper from user_test.go
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 4. Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Order placed successfully")

	// 5. Verify it was actually saved to the database with the correct fields
	var order models.Order
	config.DB.First(&order)
	assert.Equal(t, "123 Camelot Castle", order.DeliveryAddress)
	assert.Equal(t, 150.50, order.TotalPrice)
	assert.Equal(t, "processing", order.Status)
}

func TestCheckout_MissingRequiredFields(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupCheckoutRouter()

	// Missing ShippingAddress and CartItems
	checkoutData := map[string]interface{}{
		"total_price": 100.00,
	}
	jsonValue, _ := json.Marshal(checkoutData)

	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should fail Gin's 'binding:"required"' validation
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid checkout data")
}

func TestCheckout_Unauthorized(t *testing.T) {
	router := setupCheckoutRouter()

	// Valid payload, but no authorization token
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
