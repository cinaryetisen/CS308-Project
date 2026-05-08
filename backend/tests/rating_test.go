package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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

func setupRatingRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Public
	router.GET("/api/products/:id/ratings", controllers.GetProductRatings)

	// Protected (Customer)
	protected := router.Group("/api")
	protected.Use(security.AuthMiddleware())
	protected.POST("/ratings", controllers.CreateRating)
	protected.GET("/me/ratings/:productId", controllers.GetMyRatingForProduct)

	return router
}

// seedDeliveredOrder creates a user, a delivered order, and an order item linking that user
// to the given Mongo product so that CreateRating's "must have purchased" check passes.
// Email is derived from userID to avoid colliding on the User table's UNIQUE(email) constraint
// when more than one user is seeded in the same test.
func seedDeliveredOrder(userID uint, userName string, orderID uint, productHex string) {
	config.DB.Create(&models.User{
		ID:    userID,
		Name:  userName,
		Email: fmt.Sprintf("user%d@test.local", userID),
	})
	config.DB.Create(&models.Order{ID: orderID, CustomerID: userID, Status: "delivered"})
	config.DB.Create(&models.OrderItem{OrderID: orderID, ProductID: productHex})
}

// ==========================================
// CREATE RATING TESTS
// ==========================================

func TestCreateRating_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")
	defer clearMongoCollection("reviews")
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	productsCollection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	productsCollection.InsertOne(context.Background(), models.Product{
		ID:          productID,
		Name:        "Iron Sword",
		Rating:      4.0, // current 4.0 average
		ReviewCount: 2,   // based on 2 ratings
	})

	seedDeliveredOrder(1, "Arthur", 1, productID.Hex())

	router := setupRatingRouter()

	ratingData := map[string]interface{}{
		"product_id": productID.Hex(),
		"rating":     5,
	}
	jsonValue, _ := json.Marshal(ratingData)

	req, _ := http.NewRequest("POST", "/api/ratings", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "Rating submitted")

	// Rating document was inserted
	ratingsCollection := config.MongoClient.Database(config.MongoDBName).Collection("ratings")
	var saved models.Rating
	err := ratingsCollection.FindOne(context.Background(), bson.M{"product_id": productID, "user_id": uint(1)}).Decode(&saved)
	assert.NoError(t, err)
	assert.Equal(t, 5, saved.Rating)
	assert.Equal(t, "Arthur", saved.UserName)

	// Product running rating updated immediately: ((4.0 * 2) + 5) / 3 = 4.333…
	var product models.Product
	productsCollection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&product)
	assert.Equal(t, 3, product.ReviewCount)
	assert.Greater(t, product.Rating, 4.33)
	assert.Less(t, product.Rating, 4.34)

	// And — critical for the separation — no Review document was created
	reviewsCollection := config.MongoClient.Database(config.MongoDBName).Collection("reviews")
	count, _ := reviewsCollection.CountDocuments(context.Background(), bson.M{"product_id": productID})
	assert.Equal(t, int64(0), count)
}

func TestCreateRating_NotDelivered(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")

	// Order exists but is still processing — user has not received the product yet.
	config.DB.Create(&models.User{ID: 1, Name: "Lancelot"})
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, Status: "processing"})

	productHex := "60d5ec49f1b2c8b1f8e4b1a1"
	config.DB.Create(&models.OrderItem{OrderID: 1, ProductID: productHex})

	router := setupRatingRouter()

	ratingData := map[string]interface{}{
		"product_id": productHex,
		"rating":     4,
	}
	jsonValue, _ := json.Marshal(ratingData)

	req, _ := http.NewRequest("POST", "/api/ratings", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "delivered")

	// No rating got persisted
	ratingsCollection := config.MongoClient.Database(config.MongoDBName).Collection("ratings")
	count, _ := ratingsCollection.CountDocuments(context.Background(), bson.M{})
	assert.Equal(t, int64(0), count)
}

func TestCreateRating_RejectsRatingBelowOne(t *testing.T) {
	// Binding tag: `binding:"required,min=1,max=5"`. 0 trips both `required` and `min=1`.
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")

	router := setupRatingRouter()

	ratingData := map[string]interface{}{
		"product_id": primitive.NewObjectID().Hex(),
		"rating":     0,
	}
	jsonValue, _ := json.Marshal(ratingData)

	req, _ := http.NewRequest("POST", "/api/ratings", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRating_RejectsRatingAboveFive(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")

	router := setupRatingRouter()

	ratingData := map[string]interface{}{
		"product_id": primitive.NewObjectID().Hex(),
		"rating":     6, // out of 1..5 range
	}
	jsonValue, _ := json.Marshal(ratingData)

	req, _ := http.NewRequest("POST", "/api/ratings", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRating_MissingProductID(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")

	router := setupRatingRouter()

	ratingData := map[string]interface{}{
		"rating": 4,
		// product_id intentionally omitted
	}
	jsonValue, _ := json.Marshal(ratingData)

	req, _ := http.NewRequest("POST", "/api/ratings", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateRating_Unauthorized(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")

	router := setupRatingRouter()

	ratingData := map[string]interface{}{
		"product_id": primitive.NewObjectID().Hex(),
		"rating":     4,
	}
	jsonValue, _ := json.Marshal(ratingData)

	req, _ := http.NewRequest("POST", "/api/ratings", bytes.NewBuffer(jsonValue))
	// no Authorization header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestCreateRating_DifferentUsersStack(t *testing.T) {
	// Two distinct users each rate the product once. Both ratings count;
	// the running average reflects both. The unique (product_id, user_id) index
	// only blocks duplicates from the SAME user — different users may stack freely.
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	productsCollection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	productsCollection.InsertOne(context.Background(), models.Product{
		ID:          productID,
		Name:        "Steel Kite Shield",
		Rating:      0,
		ReviewCount: 0,
	})

	// Two users with their own delivered orders for the same product.
	seedDeliveredOrder(1, "Arthur", 1, productID.Hex())
	seedDeliveredOrder(2, "Bedivere", 2, productID.Hex())

	router := setupRatingRouter()

	submit := func(userID uint, rating int) int {
		body, _ := json.Marshal(map[string]interface{}{
			"product_id": productID.Hex(),
			"rating":     rating,
		})
		req, _ := http.NewRequest("POST", "/api/ratings", bytes.NewBuffer(body))
		req.Header.Set("Authorization", getTestToken(userID, "customer"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code
	}

	assert.Equal(t, http.StatusCreated, submit(1, 5))
	assert.Equal(t, http.StatusCreated, submit(2, 3))

	// (5 + 3) / 2 = 4.0 average; count = 2
	var product models.Product
	productsCollection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&product)
	assert.Equal(t, 2, product.ReviewCount)
	assert.InDelta(t, 4.0, product.Rating, 0.001)

	// Both rating documents are persisted
	ratingsCollection := config.MongoClient.Database(config.MongoDBName).Collection("ratings")
	count, _ := ratingsCollection.CountDocuments(context.Background(), bson.M{"product_id": productID})
	assert.Equal(t, int64(2), count)
}

// ==========================================
// UPSERT (SAME USER, SAME PRODUCT) TESTS
// ==========================================

func TestCreateRating_UpdateExisting(t *testing.T) {
	// Same user submits a second rating for the same product. The first should be
	// updated in place: ratings collection still has 1 doc, review_count unchanged,
	// running average reflects the replacement (not an append).
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	productsCollection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	productsCollection.InsertOne(context.Background(), models.Product{
		ID:          productID,
		Name:        "Steel Kite Shield",
		Rating:      0,
		ReviewCount: 0,
	})

	seedDeliveredOrder(1, "Arthur", 1, productID.Hex())

	router := setupRatingRouter()

	submit := func(rating int) *httptest.ResponseRecorder {
		body, _ := json.Marshal(map[string]interface{}{
			"product_id": productID.Hex(),
			"rating":     rating,
		})
		req, _ := http.NewRequest("POST", "/api/ratings", bytes.NewBuffer(body))
		req.Header.Set("Authorization", getTestToken(1, "customer"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w
	}

	// First submission — insert path, 201.
	first := submit(5)
	assert.Equal(t, http.StatusCreated, first.Code)

	var afterInsert models.Product
	productsCollection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&afterInsert)
	assert.Equal(t, 1, afterInsert.ReviewCount)
	assert.InDelta(t, 5.0, afterInsert.Rating, 0.001)

	// Second submission, same user — update path, 200.
	second := submit(1)
	assert.Equal(t, http.StatusOK, second.Code)
	assert.Contains(t, second.Body.String(), "Rating updated")

	// Running average reflects replacement: (5*1 - 5 + 1) / 1 = 1.0; review_count still 1.
	var afterUpdate models.Product
	productsCollection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&afterUpdate)
	assert.Equal(t, 1, afterUpdate.ReviewCount)
	assert.InDelta(t, 1.0, afterUpdate.Rating, 0.001)

	// Ratings collection still contains exactly one rating doc for this user/product,
	// and its value is the most recent submission.
	ratingsCollection := config.MongoClient.Database(config.MongoDBName).Collection("ratings")
	count, _ := ratingsCollection.CountDocuments(context.Background(), bson.M{"product_id": productID, "user_id": uint(1)})
	assert.Equal(t, int64(1), count)

	var savedRating models.Rating
	ratingsCollection.FindOne(context.Background(), bson.M{"product_id": productID, "user_id": uint(1)}).Decode(&savedRating)
	assert.Equal(t, 1, savedRating.Rating)
	assert.False(t, savedRating.UpdatedAt.IsZero(), "UpdatedAt should be set after an update")
}

func TestCreateRating_UpdateAfterMultipleUsers(t *testing.T) {
	// Two users have rated the product. One of them updates their rating.
	// The running average must reflect the replaced contribution while review_count holds at 2.
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	productsCollection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	productsCollection.InsertOne(context.Background(), models.Product{
		ID:          productID,
		Name:        "Mithril Chestplate",
		Rating:      0,
		ReviewCount: 0,
	})

	seedDeliveredOrder(1, "Arthur", 1, productID.Hex())
	seedDeliveredOrder(2, "Bedivere", 2, productID.Hex())

	router := setupRatingRouter()

	submit := func(userID uint, rating int) int {
		body, _ := json.Marshal(map[string]interface{}{
			"product_id": productID.Hex(),
			"rating":     rating,
		})
		req, _ := http.NewRequest("POST", "/api/ratings", bytes.NewBuffer(body))
		req.Header.Set("Authorization", getTestToken(userID, "customer"))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		return w.Code
	}

	assert.Equal(t, http.StatusCreated, submit(1, 5))
	assert.Equal(t, http.StatusCreated, submit(2, 3))
	// avg now (5+3)/2 = 4.0, count = 2

	// Arthur changes his mind: 5 → 1.
	assert.Equal(t, http.StatusOK, submit(1, 1))

	// New running average: (4.0 * 2 - 5 + 1) / 2 = 4 / 2 = 2.0; count still 2.
	var product models.Product
	productsCollection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&product)
	assert.Equal(t, 2, product.ReviewCount)
	assert.InDelta(t, 2.0, product.Rating, 0.001)

	// Still exactly 2 rating documents.
	ratingsCollection := config.MongoClient.Database(config.MongoDBName).Collection("ratings")
	total, _ := ratingsCollection.CountDocuments(context.Background(), bson.M{"product_id": productID})
	assert.Equal(t, int64(2), total)
}

// ==========================================
// GET MY RATING FOR PRODUCT TESTS
// ==========================================

func TestGetMyRatingForProduct_NotFound(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")

	productID := primitive.NewObjectID()

	router := setupRatingRouter()
	req, _ := http.NewRequest("GET", "/api/me/ratings/"+productID.Hex(), nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// No rating exists → frontend should treat 404 as "no pre-fill".
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMyRatingForProduct_Found(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("ratings")
	collection.InsertOne(context.Background(), models.Rating{
		ProductID: productID,
		UserID:    1,
		UserName:  "Arthur",
		Rating:    4,
	})

	router := setupRatingRouter()
	req, _ := http.NewRequest("GET", "/api/me/ratings/"+productID.Hex(), nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &body)
	assert.Equal(t, float64(4), body["rating"]) // JSON numbers decode as float64
}

func TestGetMyRatingForProduct_OnlyReturnsOwnRating(t *testing.T) {
	// User A has a rating; user B asks for THEIR rating on the same product. B should get 404
	// — we never expose someone else's rating through this endpoint.
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("ratings")
	collection.InsertOne(context.Background(), models.Rating{
		ProductID: productID,
		UserID:    1, // Arthur
		Rating:    4,
	})

	router := setupRatingRouter()
	req, _ := http.NewRequest("GET", "/api/me/ratings/"+productID.Hex(), nil)
	req.Header.Set("Authorization", getTestToken(2, "customer")) // Bedivere
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetMyRatingForProduct_Unauthorized(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")

	router := setupRatingRouter()
	req, _ := http.NewRequest("GET", "/api/me/ratings/"+primitive.NewObjectID().Hex(), nil)
	// No Authorization header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetMyRatingForProduct_InvalidID(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()

	router := setupRatingRouter()
	req, _ := http.NewRequest("GET", "/api/me/ratings/not-a-real-id", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==========================================
// PUBLIC FETCH TESTS
// ==========================================

func TestGetProductRatings_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")

	productID := primitive.NewObjectID()
	otherProductID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("ratings")

	collection.InsertOne(context.Background(), models.Rating{ProductID: productID, UserID: 1, UserName: "Arthur", Rating: 5})
	collection.InsertOne(context.Background(), models.Rating{ProductID: productID, UserID: 2, UserName: "Bedivere", Rating: 3})
	// A rating on a different product — must not leak in.
	collection.InsertOne(context.Background(), models.Rating{ProductID: otherProductID, UserID: 3, UserName: "Galahad", Rating: 1})

	router := setupRatingRouter()
	req, _ := http.NewRequest("GET", "/api/products/"+productID.Hex()+"/ratings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var ratings []models.Rating
	json.Unmarshal(w.Body.Bytes(), &ratings)
	assert.Len(t, ratings, 2)
}

func TestGetProductRatings_EmptyArray(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")

	productID := primitive.NewObjectID()

	router := setupRatingRouter()
	req, _ := http.NewRequest("GET", "/api/products/"+productID.Hex()+"/ratings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Even with zero ratings, the response must be a JSON array (not null) for the frontend.
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

func TestGetProductRatings_InvalidProductID(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("ratings")

	router := setupRatingRouter()
	req, _ := http.NewRequest("GET", "/api/products/not-a-real-id/ratings", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
