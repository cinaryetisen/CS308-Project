package tests

import (
	"bytes"
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
)

func setupOrderRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	protected := router.Group("/api")
	protected.Use(security.AuthMiddleware())
	protected.GET("/deliveries", controllers.GetDeliveryList)
	protected.PATCH("/deliveries/:id/status", controllers.UpdateOrderStatus)
	protected.GET("/orders/me", controllers.GetMyOrders)

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
