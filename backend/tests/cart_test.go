package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"medieval-store/config"
	"medieval-store/controllers"
	"medieval-store/models"
	"medieval-store/security"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupCartRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	protected := router.Group("/api")
	protected.Use(security.AuthMiddleware())
	protected.PATCH("/cart/item", controllers.AddToCart)
	protected.GET("/cart", controllers.GetCart)
	protected.DELETE("/cart", controllers.ClearCart)
	protected.DELETE("/cart/:id", controllers.RemoveFromCart)
	protected.POST("/cart/merge", controllers.MergeCarts)
	return router
}

// ==========================================
// AddToCart
// ==========================================

func TestAddToCart_NewItem(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	productHex := primitive.NewObjectID().Hex()
	router := setupCartRouter()

	body, _ := json.Marshal(map[string]interface{}{
		"product_id": productHex,
		"quantity":   2,
	})
	req, _ := http.NewRequest("PATCH", "/api/cart/item", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Row appears in cart_items with the right composite key + quantity
	var item models.CartItem
	err := config.DB.Where("user_id = ? AND product_id = ?", uint(1), productHex).First(&item).Error
	assert.NoError(t, err)
	assert.Equal(t, 2, item.Quantity)
}

func TestAddToCart_IncrementsExistingItem(t *testing.T) {
	// OnConflict in cart_controller.go:37-42 must increment, not replace.
	setupTestDB()
	defer clearTestDB()

	productHex := primitive.NewObjectID().Hex()

	// Pre-seed quantity 2
	config.DB.Create(&models.CartItem{
		UserID:    1,
		ProductID: productHex,
		Quantity:  2,
	})

	router := setupCartRouter()
	body, _ := json.Marshal(map[string]interface{}{
		"product_id": productHex,
		"quantity":   3,
	})
	req, _ := http.NewRequest("PATCH", "/api/cart/item", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var item models.CartItem
	config.DB.Where("user_id = ? AND product_id = ?", uint(1), productHex).First(&item)
	assert.Equal(t, 5, item.Quantity, "existing quantity 2 + new 3 should sum to 5")

	// And only one row exists for this (user, product) pair
	var count int64
	config.DB.Model(&models.CartItem{}).Where("user_id = ? AND product_id = ?", uint(1), productHex).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestAddToCart_RejectsMissingFields(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupCartRouter()

	// Both product_id and quantity are `binding:"required"`. Empty payload should 400.
	body, _ := json.Marshal(map[string]interface{}{})
	req, _ := http.NewRequest("PATCH", "/api/cart/item", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddToCart_Unauthorized(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupCartRouter()

	body, _ := json.Marshal(map[string]interface{}{
		"product_id": primitive.NewObjectID().Hex(),
		"quantity":   1,
	})
	req, _ := http.NewRequest("PATCH", "/api/cart/item", bytes.NewBuffer(body))
	// No Authorization header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==========================================
// GetCart (joins PG cart_items with Mongo products)
// ==========================================

func TestGetCart_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	productsCollection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	productsCollection.InsertOne(context.Background(), models.Product{
		ID:    productID,
		Name:  "Iron Sword",
		Price: 50.0,
	})

	config.DB.Create(&models.CartItem{
		UserID:    1,
		ProductID: productID.Hex(),
		Quantity:  3,
	})

	router := setupCartRouter()
	req, _ := http.NewRequest("GET", "/api/cart", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.CartItemResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Len(t, response, 1)
	assert.Equal(t, "Iron Sword", response[0].Name)
	assert.Equal(t, 50.0, response[0].Price)
	assert.Equal(t, 3, response[0].Quantity)
	assert.Equal(t, 150.0, response[0].Subtotal, "subtotal = price * quantity")
}

func TestGetCart_Empty(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()

	router := setupCartRouter()
	req, _ := http.NewRequest("GET", "/api/cart", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Empty cart short-circuits to []models.Product{} (cart_controller.go:73-76)
	assert.Equal(t, "[]", w.Body.String())
}

// ==========================================
// RemoveFromCart
// ==========================================

func TestRemoveFromCart_Success(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	productHex := primitive.NewObjectID().Hex()
	config.DB.Create(&models.CartItem{
		UserID:    1,
		ProductID: productHex,
		Quantity:  1,
	})

	router := setupCartRouter()
	req, _ := http.NewRequest("DELETE", "/api/cart/"+productHex, nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Row gone
	var count int64
	config.DB.Model(&models.CartItem{}).Where("user_id = ? AND product_id = ?", uint(1), productHex).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestRemoveFromCart_NotFound(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupCartRouter()
	// Not in any cart
	req, _ := http.NewRequest("DELETE", "/api/cart/"+primitive.NewObjectID().Hex(), nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ==========================================
// ClearCart
// ==========================================

func TestClearCart_Success(t *testing.T) {
	// User 1's items must be wiped; user 2's items must NOT be touched.
	setupTestDB()
	defer clearTestDB()

	p1 := primitive.NewObjectID().Hex()
	p2 := primitive.NewObjectID().Hex()
	p3 := primitive.NewObjectID().Hex()

	config.DB.Create(&models.CartItem{UserID: 1, ProductID: p1, Quantity: 1})
	config.DB.Create(&models.CartItem{UserID: 1, ProductID: p2, Quantity: 2})
	config.DB.Create(&models.CartItem{UserID: 2, ProductID: p3, Quantity: 5})

	router := setupCartRouter()
	req, _ := http.NewRequest("DELETE", "/api/cart", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// User 1: empty
	var u1Count int64
	config.DB.Model(&models.CartItem{}).Where("user_id = ?", uint(1)).Count(&u1Count)
	assert.Equal(t, int64(0), u1Count)

	// User 2: untouched
	var u2Count int64
	config.DB.Model(&models.CartItem{}).Where("user_id = ?", uint(2)).Count(&u2Count)
	assert.Equal(t, int64(1), u2Count)
}

// ==========================================
// MergeCarts (guest cart → server cart on login)
// ==========================================

func TestMergeCarts_NewItems(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	p1 := primitive.NewObjectID().Hex()
	p2 := primitive.NewObjectID().Hex()

	guestItems := []map[string]interface{}{
		{"product_id": p1, "quantity": 2},
		{"product_id": p2, "quantity": 1},
	}
	body, _ := json.Marshal(guestItems)

	router := setupCartRouter()
	req, _ := http.NewRequest("POST", "/api/cart/merge", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var items []models.CartItem
	config.DB.Where("user_id = ?", uint(1)).Find(&items)
	assert.Len(t, items, 2)
}

func TestMergeCarts_IncrementsExisting(t *testing.T) {
	// Existing user cart has product P qty 2; guest cart has same P qty 3.
	// After merge, P quantity should be 5 (the OnConflict increment, same as AddToCart).
	setupTestDB()
	defer clearTestDB()

	p := primitive.NewObjectID().Hex()
	config.DB.Create(&models.CartItem{UserID: 1, ProductID: p, Quantity: 2})

	guestItems := []map[string]interface{}{
		{"product_id": p, "quantity": 3},
	}
	body, _ := json.Marshal(guestItems)

	router := setupCartRouter()
	req, _ := http.NewRequest("POST", "/api/cart/merge", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var item models.CartItem
	config.DB.Where("user_id = ? AND product_id = ?", uint(1), p).First(&item)
	assert.Equal(t, 5, item.Quantity)

	// Still exactly one row
	var count int64
	config.DB.Model(&models.CartItem{}).Where("user_id = ? AND product_id = ?", uint(1), p).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestMergeCarts_RejectsInvalidPayload(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupCartRouter()

	// MergeCarts expects a JSON array; sending an object instead must fail validation.
	body, _ := json.Marshal(map[string]interface{}{"not": "an array"})
	req, _ := http.NewRequest("POST", "/api/cart/merge", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
