package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"medieval-store/config"
	"medieval-store/controllers"
	"medieval-store/models"
	"medieval-store/security"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupRefundRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	customer := router.Group("/api")
	customer.Use(security.AuthMiddleware())
	customer.POST("/orders/:id/refund", controllers.RequestRefund)
	customer.GET("/orders/me/refunds", controllers.GetMyRefunds)

	sm := router.Group("/api/admin")
	sm.Use(security.AuthMiddleware(), security.Authorize("sales_manager"))
	sm.GET("/refunds", controllers.GetRefundRequests)
	sm.PATCH("/refunds/:id", controllers.ResolveRefund)

	return router
}

// seedDeliveredOrder creates a delivered order for user 1 with the given items.
// Each entry is (productID, quantity, price). Returns the order and its items.
func seedRefundableOrder(t *testing.T, daysAgo int, entries ...[3]interface{}) (models.Order, []models.OrderItem) {
	t.Helper()
	order := models.Order{
		CustomerID:      1,
		TotalPrice:      0,
		DeliveryAddress: "Camelot",
		Status:          "delivered",
		Completed:       true,
		CreatedAt:       time.Now().AddDate(0, 0, -daysAgo),
	}
	if err := config.DB.Create(&order).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}

	var items []models.OrderItem
	for _, e := range entries {
		item := models.OrderItem{
			OrderID:   order.ID,
			ProductID: e[0].(string),
			Quantity:  e[1].(int),
			Price:     e[2].(float64),
		}
		if err := config.DB.Create(&item).Error; err != nil {
			t.Fatalf("failed to seed order item: %v", err)
		}
		items = append(items, item)
	}
	return order, items
}

func requestRefund(router *gin.Engine, orderID, itemID uint, token string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]interface{}{
		"order_item_id": itemID,
		"reason":        "It broke",
	})
	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/orders/%d/refund", orderID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func resolveRefund(router *gin.Engine, refundID uint, action string) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]string{"action": action})
	req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/admin/refunds/%d", refundID), bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(99, "sales_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

// ==========================================
// RequestRefund — eligibility rules
// ==========================================

func TestRequestRefund_SuccessWithinWindow(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	config.DB.Create(&models.User{ID: 1, Name: "Arthur"})
	order, items := seedRefundableOrder(t, 10, [3]interface{}{primitive.NewObjectID().Hex(), 1, 80.0})

	router := setupRefundRouter()
	w := requestRefund(router, order.ID, items[0].ID, getTestToken(1, "customer"))

	assert.Equal(t, http.StatusCreated, w.Code)

	var refund models.Refund
	config.DB.First(&refund)
	assert.Equal(t, "pending", refund.Status)
	assert.Equal(t, 80.0, refund.RefundAmount, "refund amount = purchase-time price snapshot")
}

func TestRequestRefund_RejectedOutside30Days(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	config.DB.Create(&models.User{ID: 1, Name: "Arthur"})
	order, items := seedRefundableOrder(t, 45, [3]interface{}{primitive.NewObjectID().Hex(), 1, 80.0})

	router := setupRefundRouter()
	w := requestRefund(router, order.ID, items[0].ID, getTestToken(1, "customer"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "REFUND_WINDOW_EXPIRED")
}

func TestRequestRefund_RejectedWhenNotDelivered(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	config.DB.Create(&models.User{ID: 1, Name: "Arthur"})
	order := models.Order{CustomerID: 1, Status: "processing", DeliveryAddress: "X", CreatedAt: time.Now()}
	config.DB.Create(&order)
	item := models.OrderItem{OrderID: order.ID, ProductID: primitive.NewObjectID().Hex(), Quantity: 1, Price: 10}
	config.DB.Create(&item)

	router := setupRefundRouter()
	w := requestRefund(router, order.ID, item.ID, getTestToken(1, "customer"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestRequestRefund_DuplicateBlocked(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	config.DB.Create(&models.User{ID: 1, Name: "Arthur"})
	order, items := seedRefundableOrder(t, 5, [3]interface{}{primitive.NewObjectID().Hex(), 1, 80.0})

	router := setupRefundRouter()
	first := requestRefund(router, order.ID, items[0].ID, getTestToken(1, "customer"))
	assert.Equal(t, http.StatusCreated, first.Code)

	second := requestRefund(router, order.ID, items[0].ID, getTestToken(1, "customer"))
	assert.Equal(t, http.StatusConflict, second.Code)
}

func TestRequestRefund_NotOwnerForbidden(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	config.DB.Create(&models.User{ID: 1, Name: "Arthur"})
	order, items := seedRefundableOrder(t, 5, [3]interface{}{primitive.NewObjectID().Hex(), 1, 80.0})

	router := setupRefundRouter()
	w := requestRefund(router, order.ID, items[0].ID, getTestToken(2, "customer"))

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ==========================================
// ResolveRefund — selective per-item semantics (B06)
// ==========================================

func TestResolveRefund_PartialApprovalKeepsOrderRefundable(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})

	// Two products, both in Mongo so restock works.
	p1 := primitive.NewObjectID()
	p2 := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: p1, Name: "Item One", Quantity: 5})
	collection.InsertOne(context.Background(), models.Product{ID: p2, Name: "Item Two", Quantity: 5})

	order, items := seedRefundableOrder(t, 5,
		[3]interface{}{p1.Hex(), 1, 50.0},
		[3]interface{}{p2.Hex(), 1, 70.0},
	)

	router := setupRefundRouter()

	// Refund item 1 and approve it.
	w := requestRefund(router, order.ID, items[0].ID, getTestToken(1, "customer"))
	assert.Equal(t, http.StatusCreated, w.Code)
	var refund models.Refund
	config.DB.First(&refund)
	assert.Equal(t, http.StatusOK, resolveRefund(router, refund.ID, "approved").Code)

	// Order must STILL be delivered — one of two items was returned.
	var saved models.Order
	config.DB.First(&saved, order.ID)
	assert.Equal(t, "delivered", saved.Status, "partial refund must not lock the whole order")

	// The second item must still be refundable.
	w2 := requestRefund(router, order.ID, items[1].ID, getTestToken(1, "customer"))
	assert.Equal(t, http.StatusCreated, w2.Code, "remaining items must stay individually refundable (req. 15)")
}

func TestResolveRefund_FullReturnFlipsOrderStatus(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})

	p1 := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: p1, Name: "Only Item", Quantity: 5})

	order, items := seedRefundableOrder(t, 5, [3]interface{}{p1.Hex(), 1, 50.0})

	router := setupRefundRouter()
	requestRefund(router, order.ID, items[0].ID, getTestToken(1, "customer"))
	var refund models.Refund
	config.DB.First(&refund)
	assert.Equal(t, http.StatusOK, resolveRefund(router, refund.ID, "approved").Code)

	var saved models.Order
	config.DB.First(&saved, order.ID)
	assert.Equal(t, "returned", saved.Status, "all items refunded -> whole order returned")
}

func TestResolveRefund_ApprovalRestocksProduct(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})

	p1 := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: p1, Name: "Restock Me", Quantity: 3})

	order, items := seedRefundableOrder(t, 5, [3]interface{}{p1.Hex(), 2, 50.0})

	router := setupRefundRouter()
	requestRefund(router, order.ID, items[0].ID, getTestToken(1, "customer"))
	var refund models.Refund
	config.DB.First(&refund)
	resolveRefund(router, refund.ID, "approved")

	var product models.Product
	collection.FindOne(context.Background(), bson.M{"_id": p1}).Decode(&product)
	assert.Equal(t, 5, product.Quantity, "3 + refunded qty 2 = 5")

	var saved models.Refund
	config.DB.First(&saved, refund.ID)
	assert.Equal(t, "approved", saved.Status)
	assert.NotNil(t, saved.ResolvedAt)
}

func TestResolveRefund_RejectionLeavesStockAndOrder(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})

	p1 := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: p1, Name: "Kept Item", Quantity: 3})

	order, items := seedRefundableOrder(t, 5, [3]interface{}{p1.Hex(), 1, 50.0})

	router := setupRefundRouter()
	requestRefund(router, order.ID, items[0].ID, getTestToken(1, "customer"))
	var refund models.Refund
	config.DB.First(&refund)
	assert.Equal(t, http.StatusOK, resolveRefund(router, refund.ID, "rejected").Code)

	var product models.Product
	collection.FindOne(context.Background(), bson.M{"_id": p1}).Decode(&product)
	assert.Equal(t, 3, product.Quantity, "rejection must not restock")

	var savedOrder models.Order
	config.DB.First(&savedOrder, order.ID)
	assert.Equal(t, "delivered", savedOrder.Status)

	var saved models.Refund
	config.DB.First(&saved, refund.ID)
	assert.Equal(t, "rejected", saved.Status)
}

func TestResolveRefund_AlreadyResolvedConflict(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})

	p1 := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: p1, Name: "Once Only", Quantity: 3})

	order, items := seedRefundableOrder(t, 5, [3]interface{}{p1.Hex(), 1, 50.0})

	router := setupRefundRouter()
	requestRefund(router, order.ID, items[0].ID, getTestToken(1, "customer"))
	var refund models.Refund
	config.DB.First(&refund)

	assert.Equal(t, http.StatusOK, resolveRefund(router, refund.ID, "approved").Code)
	assert.Equal(t, http.StatusConflict, resolveRefund(router, refund.ID, "approved").Code)

	// Stock restocked exactly once.
	var product models.Product
	collection.FindOne(context.Background(), bson.M{"_id": p1}).Decode(&product)
	assert.Equal(t, 4, product.Quantity)
}

func TestResolveRefund_RejectsCustomerRole(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupRefundRouter()
	body, _ := json.Marshal(map[string]string{"action": "approved"})
	req, _ := http.NewRequest("PATCH", "/api/admin/refunds/1", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ==========================================
// Restock ordering (B14)
// ==========================================

func TestResolveRefund_ApprovalSurvivesMissingProduct(t *testing.T) {
	// The PG approval commits first; a product that has vanished from MongoDB
	// must not block the refund (it is logged for manual stock correction).
	setupTestDB()
	ensureMongo()
	defer clearTestDB()

	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})

	// Order item references a product that does NOT exist in Mongo.
	ghostProduct := primitive.NewObjectID()
	order, items := seedRefundableOrder(t, 5, [3]interface{}{ghostProduct.Hex(), 1, 50.0})

	router := setupRefundRouter()
	requestRefund(router, order.ID, items[0].ID, getTestToken(1, "customer"))
	var refund models.Refund
	config.DB.First(&refund)

	w := resolveRefund(router, refund.ID, "approved")
	assert.Equal(t, http.StatusOK, w.Code, "approval must succeed even when restock has no target")

	var saved models.Refund
	config.DB.First(&saved, refund.ID)
	assert.Equal(t, "approved", saved.Status)
	assert.NotNil(t, saved.ResolvedAt)
}
