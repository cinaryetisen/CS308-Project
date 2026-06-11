package tests

import (
	"context"
	"encoding/json"
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
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type revenueResp struct {
	TotalRevenue float64 `json:"total_revenue"`
	TotalCost    float64 `json:"total_cost"`
	Profit       float64 `json:"profit"`
	OrderCount   int     `json:"order_count"`
	ItemsSold    int     `json:"items_sold"`
	Daily        []struct {
		Date    string  `json:"date"`
		Revenue float64 `json:"revenue"`
		Profit  float64 `json:"profit"`
		Orders  int     `json:"orders"`
	} `json:"daily"`
	ByCategory []struct {
		Category string  `json:"category"`
		Revenue  float64 `json:"revenue"`
		Profit   float64 `json:"profit"`
		Units    int     `json:"units"`
	} `json:"by_category"`
}

func setupRevenueRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	sm := router.Group("/api/admin")
	sm.Use(security.AuthMiddleware(), security.Authorize("sales_manager"))
	sm.GET("/revenue", controllers.GetRevenue)
	return router
}

func TestGetRevenue_AggregatesCountsAndCategories(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	sword := primitive.NewObjectID()
	potion := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: sword, Name: "Sword", Cost: 40, Category: "Weapons"})
	collection.InsertOne(context.Background(), models.Product{ID: potion, Name: "Potion", Cost: 5, Category: "Spells"})

	// Order 1: 2 swords @100 (rev 200, cost 80). Order 2: 3 potions @20 (rev 60, cost 15).
	now := time.Now()
	o1 := models.Order{CustomerID: 1, Status: "delivered", DeliveryAddress: "X", CreatedAt: now.AddDate(0, 0, -1)}
	config.DB.Create(&o1)
	config.DB.Create(&models.OrderItem{OrderID: o1.ID, ProductID: sword.Hex(), Quantity: 2, Price: 100})

	o2 := models.Order{CustomerID: 2, Status: "processing", DeliveryAddress: "Y", CreatedAt: now}
	config.DB.Create(&o2)
	config.DB.Create(&models.OrderItem{OrderID: o2.ID, ProductID: potion.Hex(), Quantity: 3, Price: 20})

	from := now.AddDate(0, 0, -7).Format("2006-01-02")
	to := now.Format("2006-01-02")

	router := setupRevenueRouter()
	req, _ := http.NewRequest("GET", "/api/admin/revenue?from="+from+"&to="+to, nil)
	req.Header.Set("Authorization", getTestToken(99, "sales_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp revenueResp
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))

	assert.Equal(t, 260.0, resp.TotalRevenue)
	assert.Equal(t, 95.0, resp.TotalCost)
	assert.Equal(t, 165.0, resp.Profit)
	assert.Equal(t, 2, resp.OrderCount)
	assert.Equal(t, 5, resp.ItemsSold, "2 swords + 3 potions")

	// Two categories, Weapons first (higher revenue).
	assert.Len(t, resp.ByCategory, 2)
	assert.Equal(t, "Weapons", resp.ByCategory[0].Category)
	assert.Equal(t, 200.0, resp.ByCategory[0].Revenue)
	assert.Equal(t, 2, resp.ByCategory[0].Units)
	assert.Equal(t, "Spells", resp.ByCategory[1].Category)
	assert.Equal(t, 60.0, resp.ByCategory[1].Revenue)

	// Daily buckets carry order counts.
	totalDailyOrders := 0
	for _, d := range resp.Daily {
		totalDailyOrders += d.Orders
	}
	assert.Equal(t, 2, totalDailyOrders)
}

func TestGetRevenue_ExcludesCancelledAndReturned(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	prod := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: prod, Name: "Item", Cost: 10, Category: "Weapons"})

	now := time.Now()
	for _, status := range []string{"cancelled", "returned"} {
		o := models.Order{CustomerID: 1, Status: status, DeliveryAddress: "X", CreatedAt: now}
		config.DB.Create(&o)
		config.DB.Create(&models.OrderItem{OrderID: o.ID, ProductID: prod.Hex(), Quantity: 1, Price: 100})
	}

	from := now.AddDate(0, 0, -1).Format("2006-01-02")
	to := now.Format("2006-01-02")

	router := setupRevenueRouter()
	req, _ := http.NewRequest("GET", "/api/admin/revenue?from="+from+"&to="+to, nil)
	req.Header.Set("Authorization", getTestToken(99, "sales_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp revenueResp
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, 0, resp.OrderCount, "cancelled/returned orders are excluded from revenue")
	assert.Equal(t, 0.0, resp.TotalRevenue)
}

func TestGetRevenue_RejectsCustomer(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupRevenueRouter()
	req, _ := http.NewRequest("GET", "/api/admin/revenue?from=2026-01-01&to=2026-12-31", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
