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

// Wipe the MongoDB reviews collection clean after every test
func clearMongoReviews() {
	if config.MongoClient != nil {
		collection := config.MongoClient.Database("medieval_store").Collection("reviews")
		collection.DeleteMany(context.Background(), bson.M{})
	}
}

// ==========================================
// CUSTOMER REVIEW TESTS
// ==========================================

func TestCreateReview_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoReviews()

	// 1. Create a User, a Delivered Order, and an OrderItem in SQLite
	config.DB.Create(&models.User{ID: 1, Name: "Arthur"})
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, Status: "delivered"})

	// Link the order to a specific MongoDB Product ID
	fakeProductID := "60d5ec49f1b2c8b1f8e4b1a1"
	config.DB.Create(&models.OrderItem{OrderID: 1, ProductID: fakeProductID})

	router := setupReviewRouter()

	reviewData := map[string]interface{}{
		"product_id": fakeProductID,
		"rating":     5,
		"comment":    "This sword is incredibly sharp!",
	}
	jsonValue, _ := json.Marshal(reviewData)

	req, _ := http.NewRequest("POST", "/api/reviews", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert it was created successfully
	assert.Equal(t, http.StatusCreated, w.Code)

	// Verify it saved to MongoDB as "pending"
	collection := config.MongoClient.Database("medieval_store").Collection("reviews")
	var savedReview models.Review
	collection.FindOne(context.Background(), bson.M{"comment": "This sword is incredibly sharp!"}).Decode(&savedReview)

	assert.Equal(t, "pending", savedReview.Status)
	assert.Equal(t, "Arthur", savedReview.UserName)
}

func TestCreateReview_NotDelivered(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoReviews()

	// Order exists, but status is still "processing"!
	config.DB.Create(&models.User{ID: 1})
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, Status: "processing"})

	fakeProductID := "60d5ec49f1b2c8b1f8e4b1a1"
	config.DB.Create(&models.OrderItem{OrderID: 1, ProductID: fakeProductID})

	router := setupReviewRouter()

	reviewData := map[string]interface{}{
		"product_id": fakeProductID,
		"rating":     4,
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

// ==========================================
// MANAGER MODERATION TESTS
// ==========================================

func TestGetPendingReviews_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoReviews()

	// Inject a pending review directly into MongoDB
	collection := config.MongoClient.Database("medieval_store").Collection("reviews")
	collection.InsertOne(context.Background(), models.Review{
		Comment: "Waiting for approval",
		Status:  "pending",
	})

	router := setupReviewRouter()
	req, _ := http.NewRequest("GET", "/api/reviews/pending", nil)
	req.Header.Set("Authorization", getTestToken(99, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Waiting for approval")
}

func TestModerateReview_Approve(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoReviews()

	// 1. Inject a pending review into MongoDB
	collection := config.MongoClient.Database("medieval_store").Collection("reviews")
	res, _ := collection.InsertOne(context.Background(), models.Review{
		Comment: "Awesome shield!",
		Status:  "pending",
		Rating:  5,
	})
	reviewID := res.InsertedID.(primitive.ObjectID).Hex()

	router := setupReviewRouter()

	// 2. Product Manager approves it
	actionData := map[string]string{"action": "approve"}
	jsonValue, _ := json.Marshal(actionData)

	req, _ := http.NewRequest("PATCH", "/api/reviews/"+reviewID+"/moderate", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(99, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Verify it actually changed to "approved" in MongoDB
	var updatedReview models.Review
	collection.FindOne(context.Background(), bson.M{"_id": res.InsertedID}).Decode(&updatedReview)
	assert.Equal(t, "approved", updatedReview.Status)
}
