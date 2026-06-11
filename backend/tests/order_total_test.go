package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"medieval-store/config"
	"medieval-store/models"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// B05: Order.TotalPrice must come from server-verified prices, never the client.

func seedCheckoutProduct(t *testing.T, price, discount float64, qty int) primitive.ObjectID {
	t.Helper()
	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{
		ID: productID, Name: "Totaled Item", Price: price, Discount: discount, Quantity: qty,
	})
	return productID
}

func TestCheckout_IgnoresClientTotal(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Spoofer", Email: "spoof@camelot.com"})
	productID := seedCheckoutProduct(t, 100.0, 20.0, 10) // effective 80

	router := setupCheckoutRouter()
	checkoutData := map[string]interface{}{
		"shipping_address": "Camelot",
		"total_price":      0.01, // tampered client total
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

	assert.Equal(t, http.StatusCreated, w.Code)

	var order models.Order
	config.DB.First(&order)
	assert.Equal(t, 160.0, order.TotalPrice, "stored total must be 2 x 80 regardless of the client's value")
}

func TestCheckout_WorksWithoutClientTotal(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Minimalist", Email: "min@camelot.com"})
	productID := seedCheckoutProduct(t, 50.0, 0, 5)

	router := setupCheckoutRouter()
	checkoutData := map[string]interface{}{
		"shipping_address": "Camelot",
		// total_price intentionally omitted
		"cart_items": []models.CartItem{
			{ProductID: productID.Hex(), Quantity: 3},
		},
	}
	jsonValue, _ := json.Marshal(checkoutData)

	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var order models.Order
	config.DB.First(&order)
	assert.Equal(t, 150.0, order.TotalPrice)
}

func TestCheckout_MultiItemTotal(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Bulk Buyer", Email: "bulk@camelot.com"})

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	p1 := primitive.NewObjectID()
	p2 := primitive.NewObjectID()
	collection.InsertOne(context.Background(), models.Product{ID: p1, Name: "Item A", Price: 10.0, Quantity: 5})
	collection.InsertOne(context.Background(), models.Product{ID: p2, Name: "Item B", Price: 200.0, Discount: 50.0, Quantity: 5}) // effective 100

	router := setupCheckoutRouter()
	checkoutData := map[string]interface{}{
		"shipping_address": "Camelot",
		"cart_items": []models.CartItem{
			{ProductID: p1.Hex(), Quantity: 2}, // 20
			{ProductID: p2.Hex(), Quantity: 1}, // 100
		},
	}
	jsonValue, _ := json.Marshal(checkoutData)

	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var order models.Order
	config.DB.First(&order)
	assert.Equal(t, 120.0, order.TotalPrice)
}
