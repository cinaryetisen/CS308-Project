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

func setupOrderRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	protected := router.Group("/api")
	protected.Use(security.AuthMiddleware())
	protected.GET("/deliveries", controllers.GetDeliveryList)
	protected.PATCH("/deliveries/:id/status", controllers.UpdateOrderStatus)
	protected.GET("/orders/me", controllers.GetMyOrders)
	protected.POST("/orders/:id/cancel", controllers.CancelOrder)
	protected.GET("/orders/:id/invoice", controllers.DownloadInvoice)

	return router
}
func TestGetDeliveryList_Success(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	config.DB.Create(&models.Order{ID: 1, DeliveryAddress: "Kingdom 1"})
	config.DB.Create(&models.Order{ID: 2, DeliveryAddress: "Kingdom 2"})

	router := setupOrderRouter()
	req, _ := http.NewRequest("GET", "/api/deliveries", nil)
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Kingdom 1")
	assert.Contains(t, w.Body.String(), "Kingdom 2")
}
func TestUpdateOrderStatus_Success_Delivered(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	config.DB.Create(&models.Order{ID: 1, Status: "processing", Completed: false})
	router := setupOrderRouter()

	updateData := map[string]string{"status": "delivered"}
	jsonValue, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PATCH", "/api/deliveries/1/status", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var order models.Order
	config.DB.First(&order, 1)
	assert.Equal(t, "delivered", order.Status)
	assert.True(t, order.Completed) // Should automatically be set to true
}
func TestUpdateOrderStatus_Success_InTransit(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	config.DB.Create(&models.Order{ID: 1, Status: "processing", Completed: false})
	router := setupOrderRouter()

	updateData := map[string]string{"status": "in-transit"}
	jsonValue, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PATCH", "/api/deliveries/1/status", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var order models.Order
	config.DB.First(&order, 1)
	assert.False(t, order.Completed) // In-transit means NOT completed
}
func TestUpdateOrderStatus_InvalidStatus(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupOrderRouter()
	updateData := map[string]string{"status": "exploded"} // Not a valid status
	jsonValue, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PATCH", "/api/deliveries/1/status", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid status")
}
func TestUpdateOrderStatus_OrderNotFound(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupOrderRouter()
	updateData := map[string]string{"status": "delivered"}
	jsonValue, _ := json.Marshal(updateData)

	// Trying to update order ID 999
	req, _ := http.NewRequest("PATCH", "/api/deliveries/999/status", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
func TestGetMyOrders_Success(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	// Customer 1 has two orders. Customer 2 has one.
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, DeliveryAddress: "Customer 1 Home"})
	config.DB.Create(&models.Order{ID: 2, CustomerID: 1, DeliveryAddress: "Customer 1 Work"})
	config.DB.Create(&models.Order{ID: 3, CustomerID: 2, DeliveryAddress: "Customer 2 Home"})

	router := setupOrderRouter()
	req, _ := http.NewRequest("GET", "/api/orders/me", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer")) // Fetching for Customer 1
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Customer 1 Home")
	assert.NotContains(t, w.Body.String(), "Customer 2 Home") // Should not see other people's orders
}
func TestGetMyOrders_EmptyList(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupOrderRouter()
	req, _ := http.NewRequest("GET", "/api/orders/me", nil)
	req.Header.Set("Authorization", getTestToken(99, "customer")) // User with no orders
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String()) // Should return an empty JSON array
}

func TestDownloadInvoice_Success(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	// Setup a user, order, and an order item to ensure preload works
	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, DeliveryAddress: "Camelot"})
	config.DB.Create(&models.OrderItem{OrderID: 1, ProductID: "60d5ec49f1b2c8b1f8e4b1a1", Quantity: 1, Price: 100.00})

	router := setupOrderRouter()

	// Try to download the invoice for Order #1
	req, _ := http.NewRequest("GET", "/api/orders/1/invoice", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer")) // Logging in as Arthur
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert the endpoint successfully generated and streamed a PDF
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.NotEmpty(t, w.Body.Bytes()) // Ensure the PDF byte stream is not empty
}

func TestDownloadInvoice_Unauthorized(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	// Order belongs to User #1
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1})

	router := setupOrderRouter()

	// Try to download the invoice while logged in as User #2 (A hacker!)
	req, _ := http.NewRequest("GET", "/api/orders/1/invoice", nil)
	req.Header.Set("Authorization", getTestToken(2, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should be blocked!
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Unauthorized to view this invoice")
}

// ==========================================
// CANCEL ORDER TESTS (D3)
// ==========================================

func TestCancelOrder_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	// Seed a product in Mongo with stock 5 so we can confirm it gets restored to 7.
	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{
		ID:       productID,
		Name:     "Iron Sword",
		Price:    100.00,
		Quantity: 5,
	})

	// Seed a processing order with one line item of qty 2 belonging to user 1.
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, Status: "processing", DeliveryAddress: "Camelot"})
	config.DB.Create(&models.OrderItem{OrderID: 1, ProductID: productID.Hex(), Quantity: 2, Price: 100.00})

	router := setupOrderRouter()
	req, _ := http.NewRequest("POST", "/api/orders/1/cancel", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Order cancelled successfully")

	// Status flipped in Postgres
	var saved models.Order
	config.DB.First(&saved, 1)
	assert.Equal(t, "cancelled", saved.Status)

	// Stock restored in Mongo: 5 + 2 = 7
	var restored models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&restored)
	assert.Equal(t, 7, restored.Quantity)
}

func TestCancelOrder_NotOwner(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	// Order belongs to user 1; user 2 attempts to cancel.
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, Status: "processing"})

	router := setupOrderRouter()
	req, _ := http.NewRequest("POST", "/api/orders/1/cancel", nil)
	req.Header.Set("Authorization", getTestToken(2, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "does not belong to you")

	// Order untouched
	var saved models.Order
	config.DB.First(&saved, 1)
	assert.Equal(t, "processing", saved.Status)
}

func TestCancelOrder_AlreadyDelivered(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	// Delivered orders must use the refund flow, not cancel.
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, Status: "delivered"})

	router := setupOrderRouter()
	req, _ := http.NewRequest("POST", "/api/orders/1/cancel", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Only orders still in processing")

	var saved models.Order
	config.DB.First(&saved, 1)
	assert.Equal(t, "delivered", saved.Status)
}

func TestCancelOrder_InTransitRejected(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, Status: "in-transit"})

	router := setupOrderRouter()
	req, _ := http.NewRequest("POST", "/api/orders/1/cancel", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCancelOrder_OrderNotFound(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupOrderRouter()
	req, _ := http.NewRequest("POST", "/api/orders/9999/cancel", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestCancelOrder_Unauthorized(t *testing.T) {
	router := setupOrderRouter()
	req, _ := http.NewRequest("POST", "/api/orders/1/cancel", nil)
	// No Authorization header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
