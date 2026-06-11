package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"medieval-store/config"
	"medieval-store/controllers"
	"medieval-store/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// B02: soft-deleted products must vanish from every public/customer surface.

func seedLiveAndDeletedProducts(t *testing.T) (live, deleted primitive.ObjectID) {
	t.Helper()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")

	live = primitive.NewObjectID()
	deleted = primitive.NewObjectID()
	now := time.Now()

	collection.InsertOne(context.Background(), models.Product{
		ID: live, Name: "Living Sword", Description: "still on sale", Price: 100, Quantity: 5,
	})
	collection.InsertOne(context.Background(), models.Product{
		ID: deleted, Name: "Deleted Sword", Description: "removed from catalog", Price: 100, Quantity: 5,
		DeletedAt: &now,
	})
	return live, deleted
}

func TestGetProducts_ExcludesSoftDeleted(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	seedLiveAndDeletedProducts(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/products", controllers.GetProducts)

	req, _ := http.NewRequest("GET", "/api/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Living Sword")
	assert.NotContains(t, w.Body.String(), "Deleted Sword")
}

func TestGetProducts_SearchExcludesSoftDeleted(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	seedLiveAndDeletedProducts(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/products", controllers.GetProducts)

	// Both names match "Sword", but only the live one may return.
	req, _ := http.NewRequest("GET", "/api/products?search=Sword", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Living Sword")
	assert.NotContains(t, w.Body.String(), "Deleted Sword")
}

func TestGetProduct_SoftDeletedReturnsNotFound(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	_, deleted := seedLiveAndDeletedProducts(t)

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/products/:id", controllers.GetProduct)

	req, _ := http.NewRequest("GET", "/api/products/"+deleted.Hex(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCheckout_SoftDeletedProductBlocked(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})
	_, deleted := seedLiveAndDeletedProducts(t)

	router := setupCheckoutRouter()
	checkoutData := map[string]interface{}{
		"shipping_address": "Camelot",
		"total_price":      100.00,
		"cart_items": []models.CartItem{
			{ProductID: deleted.Hex(), Quantity: 1},
		},
	}
	jsonValue, _ := json.Marshal(checkoutData)

	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Stock filter excludes deleted products, so the purchase must fail
	assert.Equal(t, http.StatusConflict, w.Code)

	// No ghost order persisted
	var orderCount int64
	config.DB.Model(&models.Order{}).Count(&orderCount)
	assert.Equal(t, int64(0), orderCount)
}

func TestGetCart_ExcludesSoftDeleted(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	live, deleted := seedLiveAndDeletedProducts(t)
	config.DB.Create(&models.CartItem{UserID: 1, ProductID: live.Hex(), Quantity: 1})
	config.DB.Create(&models.CartItem{UserID: 1, ProductID: deleted.Hex(), Quantity: 1})

	router := setupCartRouter()
	req, _ := http.NewRequest("GET", "/api/cart", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Living Sword")
	assert.NotContains(t, w.Body.String(), "Deleted Sword")
}

func TestGetWishlist_ExcludesSoftDeleted(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	live, deleted := seedLiveAndDeletedProducts(t)
	config.DB.Create(&models.WishlistItem{UserID: 1, ProductID: live.Hex()})
	config.DB.Create(&models.WishlistItem{UserID: 1, ProductID: deleted.Hex()})

	router := setupWishlistRouter()
	req, _ := http.NewRequest("GET", "/api/wishlist", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Living Sword")
	assert.NotContains(t, w.Body.String(), "Deleted Sword")
}

// Regression: bson.M{"deleted_at": {"$exists": false}} must not hide products
// that have an explicit null deleted_at (e.g. inserted by other tooling).
func TestGetProducts_NullDeletedAtStillVisible(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	id := primitive.NewObjectID()
	collection.InsertOne(context.Background(), bson.M{
		"_id": id, "name": "Null Marker Blade", "price": 10.0, "quantity": 1, "deleted_at": nil,
	})

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/products", controllers.GetProducts)

	req, _ := http.NewRequest("GET", "/api/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Null Marker Blade")
}
