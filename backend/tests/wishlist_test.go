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

func setupWishlistRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	protected := router.Group("/api")
	protected.Use(security.AuthMiddleware())
	protected.GET("/wishlist", controllers.GetWishlist)
	protected.POST("/wishlist", controllers.AddToWishlist)
	protected.DELETE("/wishlist/:productId", controllers.RemoveFromWishlist)
	return router
}

// ==========================================
// AddToWishlist
// ==========================================

func TestAddToWishlist_Success(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	productHex := primitive.NewObjectID().Hex()
	router := setupWishlistRouter()

	body, _ := json.Marshal(map[string]string{"product_id": productHex})
	req, _ := http.NewRequest("POST", "/api/wishlist", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var item models.WishlistItem
	err := config.DB.Where("user_id = ? AND product_id = ?", uint(1), productHex).First(&item).Error
	assert.NoError(t, err)
}

func TestAddToWishlist_Idempotent(t *testing.T) {
	// Adding the same product twice must not duplicate the row or error.
	setupTestDB()
	defer clearTestDB()

	productHex := primitive.NewObjectID().Hex()
	router := setupWishlistRouter()

	body, _ := json.Marshal(map[string]string{"product_id": productHex})

	for i := 0; i < 2; i++ {
		req, _ := http.NewRequest("POST", "/api/wishlist", bytes.NewBuffer(body))
		req.Header.Set("Authorization", getTestToken(1, "customer"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	var count int64
	config.DB.Model(&models.WishlistItem{}).
		Where("user_id = ? AND product_id = ?", uint(1), productHex).
		Count(&count)
	assert.Equal(t, int64(1), count, "duplicate add must not create a second row")
}

func TestAddToWishlist_MissingProductID(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupWishlistRouter()
	body, _ := json.Marshal(map[string]interface{}{})
	req, _ := http.NewRequest("POST", "/api/wishlist", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestAddToWishlist_Unauthorized(t *testing.T) {
	router := setupWishlistRouter()
	body, _ := json.Marshal(map[string]string{"product_id": primitive.NewObjectID().Hex()})
	req, _ := http.NewRequest("POST", "/api/wishlist", bytes.NewBuffer(body))
	// no Authorization header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==========================================
// GetWishlist (joins PG wishlist_items with Mongo products)
// ==========================================

func TestGetWishlist_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	productsCollection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	productsCollection.InsertOne(context.Background(), models.Product{
		ID:       productID,
		Name:     "Dragon Bow",
		Price:    250.0,
		Discount: 10.0,
		Quantity: 4,
		Category: "Weapons",
		ImageURL: "http://example.com/bow.png",
	})

	config.DB.Create(&models.WishlistItem{
		UserID:    1,
		ProductID: productID.Hex(),
	})

	router := setupWishlistRouter()
	req, _ := http.NewRequest("GET", "/api/wishlist", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.WishlistItemResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, "Dragon Bow", response[0].Name)
	assert.Equal(t, 250.0, response[0].Price)
	assert.Equal(t, 10.0, response[0].Discount)
	assert.Equal(t, 4, response[0].Stock)
	assert.Equal(t, "Weapons", response[0].Category)
}

func TestGetWishlist_Empty(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()

	router := setupWishlistRouter()
	req, _ := http.NewRequest("GET", "/api/wishlist", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

func TestGetWishlist_OnlyMine(t *testing.T) {
	// User 1's GET must not include User 2's wishlist rows.
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	p1ID := primitive.NewObjectID()
	p2ID := primitive.NewObjectID()

	productsCollection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	productsCollection.InsertOne(context.Background(), models.Product{ID: p1ID, Name: "Mine"})
	productsCollection.InsertOne(context.Background(), models.Product{ID: p2ID, Name: "Theirs"})

	config.DB.Create(&models.WishlistItem{UserID: 1, ProductID: p1ID.Hex()})
	config.DB.Create(&models.WishlistItem{UserID: 2, ProductID: p2ID.Hex()})

	router := setupWishlistRouter()
	req, _ := http.NewRequest("GET", "/api/wishlist", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Mine")
	assert.NotContains(t, w.Body.String(), "Theirs")
}

// ==========================================
// RemoveFromWishlist
// ==========================================

func TestRemoveFromWishlist_Success(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	productHex := primitive.NewObjectID().Hex()
	config.DB.Create(&models.WishlistItem{UserID: 1, ProductID: productHex})

	router := setupWishlistRouter()
	req, _ := http.NewRequest("DELETE", "/api/wishlist/"+productHex, nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	var count int64
	config.DB.Model(&models.WishlistItem{}).
		Where("user_id = ? AND product_id = ?", uint(1), productHex).
		Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestRemoveFromWishlist_NotFound(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupWishlistRouter()
	req, _ := http.NewRequest("DELETE", "/api/wishlist/"+primitive.NewObjectID().Hex(), nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestRemoveFromWishlist_Unauthorized(t *testing.T) {
	router := setupWishlistRouter()
	req, _ := http.NewRequest("DELETE", "/api/wishlist/"+primitive.NewObjectID().Hex(), nil)
	// no Authorization header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
