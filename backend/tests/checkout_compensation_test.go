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
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// B07: when a multi-item checkout fails partway, stock already deducted for
// earlier items must be restored — otherwise inventory leaks on every conflict.

func TestCheckout_PartialFailureRestoresEarlierStock(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	pOK := primitive.NewObjectID()    // plenty of stock — deducted first
	pScarce := primitive.NewObjectID() // not enough stock — fails second
	collection.InsertOne(context.Background(), models.Product{ID: pOK, Name: "Plenty", Price: 10, Quantity: 10})
	collection.InsertOne(context.Background(), models.Product{ID: pScarce, Name: "Scarce", Price: 99, Quantity: 1})

	router := setupCheckoutRouter()
	checkoutData := map[string]interface{}{
		"shipping_address": "Camelot",
		"cart_items": []models.CartItem{
			{ProductID: pOK.Hex(), Quantity: 2},     // succeeds, stock 10 -> 8
			{ProductID: pScarce.Hex(), Quantity: 5}, // fails, only 1 left
		},
	}
	jsonValue, _ := json.Marshal(checkoutData)

	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Whole checkout rejected
	assert.Equal(t, http.StatusConflict, w.Code)

	// No ghost order rows
	var orderCount int64
	config.DB.Model(&models.Order{}).Count(&orderCount)
	assert.Equal(t, int64(0), orderCount)

	// CRITICAL: the first item's decrement must have been compensated
	var ok models.Product
	collection.FindOne(context.Background(), bson.M{"_id": pOK}).Decode(&ok)
	assert.Equal(t, 10, ok.Quantity, "stock deducted before the failure must be restored")

	// And the scarce product was never touched
	var scarce models.Product
	collection.FindOne(context.Background(), bson.M{"_id": pScarce}).Decode(&scarce)
	assert.Equal(t, 1, scarce.Quantity)
}

func TestCheckout_ThreeItemFailureRestoresAllEarlierStock(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	p1 := primitive.NewObjectID()
	p2 := primitive.NewObjectID()
	p3 := primitive.NewObjectID()
	collection.InsertOne(context.Background(), models.Product{ID: p1, Name: "First", Price: 10, Quantity: 4})
	collection.InsertOne(context.Background(), models.Product{ID: p2, Name: "Second", Price: 10, Quantity: 6})
	collection.InsertOne(context.Background(), models.Product{ID: p3, Name: "Third", Price: 10, Quantity: 0})

	router := setupCheckoutRouter()
	checkoutData := map[string]interface{}{
		"shipping_address": "Camelot",
		"cart_items": []models.CartItem{
			{ProductID: p1.Hex(), Quantity: 1},
			{ProductID: p2.Hex(), Quantity: 3},
			{ProductID: p3.Hex(), Quantity: 1}, // out of stock — fails third
		},
	}
	jsonValue, _ := json.Marshal(checkoutData)

	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var first, second models.Product
	collection.FindOne(context.Background(), bson.M{"_id": p1}).Decode(&first)
	collection.FindOne(context.Background(), bson.M{"_id": p2}).Decode(&second)
	assert.Equal(t, 4, first.Quantity)
	assert.Equal(t, 6, second.Quantity)
}

func TestCheckout_SuccessfulMultiItemStillDeducts(t *testing.T) {
	// Guard against over-compensating: a fully successful checkout must keep
	// its decrements.
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})

	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	p1 := primitive.NewObjectID()
	p2 := primitive.NewObjectID()
	collection.InsertOne(context.Background(), models.Product{ID: p1, Name: "A", Price: 10, Quantity: 5})
	collection.InsertOne(context.Background(), models.Product{ID: p2, Name: "B", Price: 10, Quantity: 5})

	router := setupCheckoutRouter()
	checkoutData := map[string]interface{}{
		"shipping_address": "Camelot",
		"cart_items": []models.CartItem{
			{ProductID: p1.Hex(), Quantity: 2},
			{ProductID: p2.Hex(), Quantity: 1},
		},
	}
	jsonValue, _ := json.Marshal(checkoutData)

	req, _ := http.NewRequest("POST", "/api/checkout", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var a, b models.Product
	collection.FindOne(context.Background(), bson.M{"_id": p1}).Decode(&a)
	collection.FindOne(context.Background(), bson.M{"_id": p2}).Decode(&b)
	assert.Equal(t, 3, a.Quantity)
	assert.Equal(t, 4, b.Quantity)
}
