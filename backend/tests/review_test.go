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

func setupReviewRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Public
	router.GET("/api/products/:id/reviews", controllers.GetProductReviews)

	// Protected (Customer)
	protected := router.Group("/api")
	protected.Use(security.AuthMiddleware())
	protected.POST("/reviews", controllers.CreateReview)

	// Protected (Manager)
	manager := router.Group("/api")
	manager.Use(security.AuthMiddleware(), security.Authorize("product_manager"))
	manager.GET("/reviews/pending", controllers.GetPendingReviews)
	manager.PATCH("/reviews/:id/moderate", controllers.ModerateReview)

	return router
}

// Ensure MongoDB is connected for our tests
func ensureMongo() {
	if config.MongoClient == nil {
		config.ConnectMongo() // Uses the .env loaded by setupTestDB()
	}
}

// ==========================================
// CUSTOMER COMMENT-ONLY REVIEW TESTS
// (After the rating/comment split, /api/reviews accepts comment only.)
// ==========================================

func TestCreateReview_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("reviews")
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Arthur"})
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, Status: "delivered"})

	// Seed a product so we can prove a comment submission does NOT touch the running rating.
	productID := primitive.NewObjectID()
	productsCollection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	productsCollection.InsertOne(context.Background(), models.Product{
		ID:          productID,
		Name:        "Iron Sword",
		Rating:      4.0,
		ReviewCount: 2,
	})

	config.DB.Create(&models.OrderItem{OrderID: 1, ProductID: productID.Hex()})

	router := setupReviewRouter()

	reviewData := map[string]interface{}{
		"product_id": productID.Hex(),
		"comment":    "This sword is incredibly sharp!",
	}
	jsonValue, _ := json.Marshal(reviewData)

	req, _ := http.NewRequest("POST", "/api/reviews", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "awaiting moderation")

	// Review document inserted as pending with the right content
	reviewsCollection := config.MongoClient.Database(config.MongoDBName).Collection("reviews")
	var saved models.Review
	err := reviewsCollection.FindOne(context.Background(), bson.M{"product_id": productID}).Decode(&saved)
	assert.NoError(t, err)
	assert.Equal(t, "pending", saved.Status)
	assert.Equal(t, "This sword is incredibly sharp!", saved.Comment)
	assert.Equal(t, uint(1), saved.UserID)
	assert.Equal(t, "Arthur", saved.UserName)

	// Critically: the product's running rating must NOT change when a comment is submitted.
	// Rating updates flow through /api/ratings only (see rating_test.go).
	var product models.Product
	productsCollection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&product)
	assert.Equal(t, 4.0, product.Rating)
	assert.Equal(t, 2, product.ReviewCount)
}

func TestCreateReview_NotDelivered(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("reviews")

	// Order exists, but status is still "processing"!
	config.DB.Create(&models.User{ID: 1})
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, Status: "processing"})

	fakeProductID := "60d5ec49f1b2c8b1f8e4b1a1"
	config.DB.Create(&models.OrderItem{OrderID: 1, ProductID: fakeProductID})

	router := setupReviewRouter()

	reviewData := map[string]interface{}{
		"product_id": fakeProductID,
		"comment":    "I haven't gotten it yet, but I'm excited.",
	}
	jsonValue, _ := json.Marshal(reviewData)

	req, _ := http.NewRequest("POST", "/api/reviews", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should block the user with a 403 Forbidden
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "delivered")
}

func TestCreateReview_MissingComment(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("reviews")

	config.DB.Create(&models.User{ID: 1, Name: "Arthur"})

	router := setupReviewRouter()

	// Comment is required by Gin's binding tags. Sending only product_id should fail validation.
	reviewData := map[string]interface{}{
		"product_id": "60d5ec49f1b2c8b1f8e4b1a1",
	}
	jsonValue, _ := json.Marshal(reviewData)

	req, _ := http.NewRequest("POST", "/api/reviews", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateReview_LegacyRatingFieldIsIgnored(t *testing.T) {
	// Pre-split, the frontend posted {product_id, rating, comment}. After the split, /api/reviews
	// should accept the call (extra field ignored) and never bump the product's running rating.
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("reviews")
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Arthur"})
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, Status: "delivered"})

	productID := primitive.NewObjectID()
	productsCollection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	productsCollection.InsertOne(context.Background(), models.Product{
		ID:          productID,
		Rating:      3.0,
		ReviewCount: 5,
	})
	config.DB.Create(&models.OrderItem{OrderID: 1, ProductID: productID.Hex()})

	router := setupReviewRouter()

	reviewData := map[string]interface{}{
		"product_id": productID.Hex(),
		"rating":     5, // legacy — must be ignored
		"comment":    "Extra rating field should be ignored",
	}
	jsonValue, _ := json.Marshal(reviewData)

	req, _ := http.NewRequest("POST", "/api/reviews", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Product running rating untouched
	var product models.Product
	productsCollection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&product)
	assert.Equal(t, 3.0, product.Rating)
	assert.Equal(t, 5, product.ReviewCount)
}

// ==========================================
// PUBLIC FETCH TESTS
// ==========================================

func TestGetProductReviews_OnlyApproved(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("reviews")

	productID := primitive.NewObjectID()
	otherProductID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("reviews")

	collection.InsertOne(context.Background(), models.Review{ProductID: productID, Comment: "approved one", Status: "approved"})
	collection.InsertOne(context.Background(), models.Review{ProductID: productID, Comment: "still pending", Status: "pending"})
	collection.InsertOne(context.Background(), models.Review{ProductID: productID, Comment: "rejected one", Status: "rejected"})
	// Approved review on a different product — must not leak in.
	collection.InsertOne(context.Background(), models.Review{ProductID: otherProductID, Comment: "different product", Status: "approved"})

	router := setupReviewRouter()
	req, _ := http.NewRequest("GET", "/api/products/"+productID.Hex()+"/reviews", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "approved one")
	assert.NotContains(t, body, "still pending")
	assert.NotContains(t, body, "rejected one")
	assert.NotContains(t, body, "different product")
}

// ==========================================
// MANAGER MODERATION TESTS
// ==========================================

func TestGetPendingReviews_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("reviews")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("reviews")
	collection.InsertOne(context.Background(), models.Review{Comment: "Waiting for approval", Status: "pending"})
	// An approved review must NOT show up in the pending queue
	collection.InsertOne(context.Background(), models.Review{Comment: "Already approved", Status: "approved"})

	router := setupReviewRouter()
	req, _ := http.NewRequest("GET", "/api/reviews/pending", nil)
	req.Header.Set("Authorization", getTestToken(99, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Waiting for approval")
	assert.NotContains(t, w.Body.String(), "Already approved")
}

func TestModerateReview_Approve(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("reviews")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("reviews")
	res, _ := collection.InsertOne(context.Background(), models.Review{
		Comment: "Awesome shield!",
		Status:  "pending",
	})
	reviewID := res.InsertedID.(primitive.ObjectID).Hex()

	router := setupReviewRouter()

	actionData := map[string]string{"action": "approve"}
	jsonValue, _ := json.Marshal(actionData)

	req, _ := http.NewRequest("PATCH", "/api/reviews/"+reviewID+"/moderate", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(99, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedReview models.Review
	collection.FindOne(context.Background(), bson.M{"_id": res.InsertedID}).Decode(&updatedReview)
	assert.Equal(t, "approved", updatedReview.Status)
}

func TestModerateReview_Reject(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("reviews")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("reviews")
	res, _ := collection.InsertOne(context.Background(), models.Review{
		Comment: "spam spam spam",
		Status:  "pending",
	})
	reviewID := res.InsertedID.(primitive.ObjectID).Hex()

	router := setupReviewRouter()

	actionData := map[string]string{"action": "reject"}
	jsonValue, _ := json.Marshal(actionData)

	req, _ := http.NewRequest("PATCH", "/api/reviews/"+reviewID+"/moderate", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(99, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var updatedReview models.Review
	collection.FindOne(context.Background(), bson.M{"_id": res.InsertedID}).Decode(&updatedReview)
	assert.Equal(t, "rejected", updatedReview.Status)
}

func TestModerateReview_InvalidAction(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("reviews")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("reviews")
	res, _ := collection.InsertOne(context.Background(), models.Review{Comment: "?", Status: "pending"})
	reviewID := res.InsertedID.(primitive.ObjectID).Hex()

	router := setupReviewRouter()

	actionData := map[string]string{"action": "burn-it"} // not a valid action
	jsonValue, _ := json.Marshal(actionData)

	req, _ := http.NewRequest("PATCH", "/api/reviews/"+reviewID+"/moderate", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(99, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// And the review's status didn't change
	var unchanged models.Review
	collection.FindOne(context.Background(), bson.M{"_id": res.InsertedID}).Decode(&unchanged)
	assert.Equal(t, "pending", unchanged.Status)
}
