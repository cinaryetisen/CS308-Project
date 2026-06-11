package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
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

// B13: the product's rating average and count are updated with a single atomic
// MongoDB pipeline, so concurrent raters can no longer lose each other's writes.

func setupAtomicRatingRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	protected := router.Group("/api")
	protected.Use(security.AuthMiddleware())
	protected.POST("/ratings", controllers.CreateRating)
	return router
}

func postRating(router *gin.Engine, userID uint, productHex string, rating int) *httptest.ResponseRecorder {
	body, _ := json.Marshal(map[string]interface{}{"product_id": productHex, "rating": rating})
	req, _ := http.NewRequest("POST", "/api/ratings", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(userID, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestRating_ConcurrentFirstRatingsAllCounted(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")
	defer clearMongoCollection("ratings")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{
		ID: productID, Name: "Hot Item", Rating: 0, ReviewCount: 0,
	})

	// 10 distinct users, each with a delivered order containing the product.
	const raters = 10
	for i := 1; i <= raters; i++ {
		config.DB.Create(&models.User{ID: uint(i), Name: fmt.Sprintf("User %d", i), Email: fmt.Sprintf("u%d@x.com", i)})
		config.DB.Create(&models.Order{ID: uint(i), CustomerID: uint(i), Status: "delivered", DeliveryAddress: "X"})
		config.DB.Create(&models.OrderItem{OrderID: uint(i), ProductID: productID.Hex(), Quantity: 1, Price: 10})
	}

	router := setupAtomicRatingRouter()

	// Fire all 10 first-ratings concurrently. Ratings alternate 1..5,1..5 => avg 3.0.
	var wg sync.WaitGroup
	for i := 1; i <= raters; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()
			rating := (userID-1)%5 + 1
			w := postRating(router, uint(userID), productID.Hex(), rating)
			assert.Equal(t, http.StatusCreated, w.Code)
		}(i)
	}
	wg.Wait()

	var product models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&product)
	assert.Equal(t, raters, product.ReviewCount, "every concurrent rating must be counted")
	assert.InDelta(t, 3.0, product.Rating, 0.0001, "average of 1..5 twice is exactly 3.0")
}

func TestRating_SequentialAverageStillExact(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")
	defer clearMongoCollection("ratings")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{
		ID: productID, Name: "Steady Item", Rating: 4.0, ReviewCount: 2, // pre-existing avg
	})

	config.DB.Create(&models.User{ID: 1, Name: "Rater", Email: "r@x.com"})
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, Status: "delivered", DeliveryAddress: "X"})
	config.DB.Create(&models.OrderItem{OrderID: 1, ProductID: productID.Hex(), Quantity: 1, Price: 10})

	router := setupAtomicRatingRouter()

	// New rating 5 on top of avg 4.0 x2 => (8+5)/3 = 4.3333
	w := postRating(router, 1, productID.Hex(), 5)
	assert.Equal(t, http.StatusCreated, w.Code)

	var product models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&product)
	assert.Equal(t, 3, product.ReviewCount)
	assert.InDelta(t, 13.0/3.0, product.Rating, 0.0001)

	// Same user updates to 2: avg shifts by (2-5)/3 => (13-3)/3 = 10/3
	w2 := postRating(router, 1, productID.Hex(), 2)
	assert.Equal(t, http.StatusOK, w2.Code)

	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&product)
	assert.Equal(t, 3, product.ReviewCount, "update must not change the count")
	assert.InDelta(t, 10.0/3.0, product.Rating, 0.0001)
}
