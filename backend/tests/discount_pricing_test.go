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

// B04: the advertised discount must be honored everywhere money is computed.

func TestEffectivePrice(t *testing.T) {
	p := models.Product{Price: 100.0, Discount: 20.0}
	assert.Equal(t, 80.0, p.EffectivePrice())

	noDiscount := models.Product{Price: 100.0, Discount: 0}
	assert.Equal(t, 100.0, noDiscount.EffectivePrice())

	halfOff := models.Product{Price: 850.0, Discount: 50.0}
	assert.Equal(t, 425.0, halfOff.EffectivePrice())
}

func TestCheckout_AppliesDiscountToPriceSnapshot(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Bargain Hunter", Email: "deals@camelot.com"})

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{
		ID:       productID,
		Name:     "Discounted Wand",
		Price:    100.00,
		Discount: 20.0,
		Quantity: 10,
	})

	router := setupCheckoutRouter()
	checkoutData := map[string]interface{}{
		"shipping_address": "Camelot",
		"total_price":      160.00,
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

	// The price snapshot (used later by refunds, req. 15) records the DISCOUNTED price.
	var orderItem models.OrderItem
	config.DB.First(&orderItem)
	assert.Equal(t, 80.00, orderItem.Price)
}

func TestCheckout_FullPriceWhenNoDiscount(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Full Payer", Email: "full@camelot.com"})

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{
		ID:       productID,
		Name:     "Plain Sword",
		Price:    150.50,
		Discount: 0,
		Quantity: 5,
	})

	router := setupCheckoutRouter()
	checkoutData := map[string]interface{}{
		"shipping_address": "Camelot",
		"total_price":      150.50,
		"cart_items": []models.CartItem{
			{ProductID: productID.Hex(), Quantity: 1},
		},
	}
	jsonValue, _ := json.Marshal(checkoutData)

	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var orderItem models.OrderItem
	config.DB.First(&orderItem)
	assert.Equal(t, 150.50, orderItem.Price)
}

func TestGetCart_AppliesDiscount(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{
		ID:       productID,
		Name:     "Sale Shield",
		Price:    100.0,
		Discount: 25.0,
		Quantity: 9,
	})

	config.DB.Create(&models.CartItem{UserID: 1, ProductID: productID.Hex(), Quantity: 2})

	router := setupCartRouter()
	req, _ := http.NewRequest("GET", "/api/cart", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.CartItemResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Len(t, response, 1)
	assert.Equal(t, 75.0, response[0].Price, "unit price must have the discount applied")
	assert.Equal(t, 100.0, response[0].OriginalPrice)
	assert.Equal(t, 25.0, response[0].Discount)
	assert.Equal(t, 150.0, response[0].Subtotal, "subtotal = discounted price * quantity")
}
