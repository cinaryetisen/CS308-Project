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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupPricePendingRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/products", controllers.GetProducts)
	router.GET("/api/products/:id", controllers.GetProduct)

	pm := router.Group("/api/admin")
	pm.Use(security.AuthMiddleware(), security.Authorize("product_manager"))
	pm.POST("/products", controllers.CreateProduct)

	sm := router.Group("/api/admin")
	sm.Use(security.AuthMiddleware(), security.Authorize("sales_manager"))
	sm.PATCH("/products/:id/price", controllers.UpdateProductPrice)

	return router
}

func createPendingProduct(t *testing.T, router *gin.Engine) primitive.ObjectID {
	t.Helper()
	body, _ := json.Marshal(map[string]interface{}{
		"name": "Fresh Blade", "model": "FB-1", "serial_number": "SN-FB-1",
		"description": "Just added by the PM.", "quantity": 5,
		"category": "Weapons", "distributor": "Forge", "warranty": "1 Year",
	})
	req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	var resp struct {
		Product models.Product `json:"product"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	return resp.Product.ID
}

func TestPendingProduct_HiddenFromPublicListing(t *testing.T) {
	setupTestDB()
	ensureMongo()
	seedCategory(t, "Weapons")
	defer clearMongoCollection("categories")
	defer clearMongoCollection("products")

	router := setupPricePendingRouter()
	createPendingProduct(t, router)

	req, _ := http.NewRequest("GET", "/api/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotContains(t, w.Body.String(), "Fresh Blade", "pending-price products must not be listed publicly")
}

func TestPendingProduct_VisibleWithIncludePending(t *testing.T) {
	setupTestDB()
	ensureMongo()
	seedCategory(t, "Weapons")
	defer clearMongoCollection("categories")
	defer clearMongoCollection("products")

	router := setupPricePendingRouter()
	createPendingProduct(t, router)

	req, _ := http.NewRequest("GET", "/api/products?include_pending=true", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Fresh Blade", "manager panels must still see pending products")
}

func TestPendingProduct_DetailHiddenWithoutParam(t *testing.T) {
	setupTestDB()
	ensureMongo()
	seedCategory(t, "Weapons")
	defer clearMongoCollection("categories")
	defer clearMongoCollection("products")

	router := setupPricePendingRouter()
	id := createPendingProduct(t, router)

	req, _ := http.NewRequest("GET", "/api/products/"+id.Hex(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	req2, _ := http.NewRequest("GET", "/api/products/"+id.Hex()+"?include_pending=true", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestPendingProduct_CheckoutBlocked(t *testing.T) {
	setupTestDB()
	ensureMongo()
	seedCategory(t, "Weapons")
	defer clearMongoCollection("categories")
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Eager Buyer", Email: "eager@camelot.com"})

	router := setupPricePendingRouter()
	id := createPendingProduct(t, router)

	checkoutRouter := setupCheckoutRouter()
	checkoutData := map[string]interface{}{
		"shipping_address": "Camelot",
		"cart_items": []models.CartItem{
			{ProductID: id.Hex(), Quantity: 1},
		},
	}
	jsonValue, _ := json.Marshal(checkoutData)
	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	checkoutRouter.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code, "a pending-price product must not be purchasable")
}

func TestPendingProduct_PriceSetPublishesIt(t *testing.T) {
	setupTestDB()
	ensureMongo()
	seedCategory(t, "Weapons")
	defer clearMongoCollection("categories")
	defer clearMongoCollection("products")

	router := setupPricePendingRouter()
	id := createPendingProduct(t, router)

	// Sales manager sets the real price.
	body, _ := json.Marshal(map[string]interface{}{"price": 199.99})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+id.Hex()+"/price", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(2, "sales_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Flag cleared in the document...
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	var saved models.Product
	collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&saved)
	assert.False(t, saved.PricePending)
	assert.Equal(t, 199.99, saved.Price)

	// ...and the product now appears publicly.
	listReq, _ := http.NewRequest("GET", "/api/products", nil)
	listW := httptest.NewRecorder()
	router.ServeHTTP(listW, listReq)
	assert.Contains(t, listW.Body.String(), "Fresh Blade")
}

func TestUpdateProductPrice_SetsCost(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{
		ID: productID, Name: "Costless", Price: 99999.99, Cost: 0, PricePending: true,
	})

	router := setupPricePendingRouter()
	body, _ := json.Marshal(map[string]interface{}{"price": 200.0, "cost": 120.0})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+productID.Hex()+"/price", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(2, "sales_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var saved models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&saved)
	assert.Equal(t, 200.0, saved.Price)
	assert.Equal(t, 120.0, saved.Cost)
	assert.False(t, saved.PricePending)
}

func TestUpdateProductPrice_CostOptional(t *testing.T) {
	// Omitting cost leaves the existing cost untouched.
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{
		ID: productID, Name: "Has Cost", Price: 100, Cost: 55,
	})

	router := setupPricePendingRouter()
	body, _ := json.Marshal(map[string]interface{}{"price": 150.0})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+productID.Hex()+"/price", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(2, "sales_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var saved models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&saved)
	assert.Equal(t, 150.0, saved.Price)
	assert.Equal(t, 55.0, saved.Cost, "omitted cost must be left unchanged")
}

func TestUpdateProductPrice_NegativeCostRejected(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: productID, Name: "X", Price: 100, Cost: 10})

	router := setupPricePendingRouter()
	body, _ := json.Marshal(map[string]interface{}{"price": 150.0, "cost": -5.0})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+productID.Hex()+"/price", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(2, "sales_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
