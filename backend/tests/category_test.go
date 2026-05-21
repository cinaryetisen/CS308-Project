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

// ==========================================
// Category model — round-trip + uniqueness (B5)
// ==========================================

func TestCategory_RoundTrip(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")
	_, err := collection.InsertOne(context.Background(), models.Category{
		Name: "Weapons",
	})
	assert.NoError(t, err)

	var fetched models.Category
	err = collection.FindOne(context.Background(), bson.M{"name": "Weapons"}).Decode(&fetched)
	assert.NoError(t, err)
	assert.Equal(t, "Weapons", fetched.Name)
	assert.False(t, fetched.ID.IsZero(), "Mongo should populate _id on insert")
}

func TestCategory_NameMustBeUnique(t *testing.T) {
	// The unique index in config/mongo.go must reject duplicate category names.
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")

	_, err := collection.InsertOne(context.Background(), models.Category{Name: "Spells"})
	assert.NoError(t, err)

	// Second insert with the same name must fail.
	_, err = collection.InsertOne(context.Background(), models.Category{Name: "Spells"})
	assert.Error(t, err)
}

// ==========================================
// Category endpoints
// ==========================================

func setupCategoryRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Public
	router.GET("/api/categories", controllers.GetCategories)

	// Product manager only
	pm := router.Group("/api/admin")
	pm.Use(security.AuthMiddleware(), security.Authorize("product_manager"))
	pm.POST("/categories", controllers.CreateCategory)
	pm.DELETE("/categories/:id", controllers.DeleteCategory)

	return router
}

// --- GetCategories ---

func TestGetCategories_Empty(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")

	router := setupCategoryRouter()
	req, _ := http.NewRequest("GET", "/api/categories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

func TestGetCategories_SortedByName(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")
	collection.InsertOne(context.Background(), models.Category{Name: "Weapons"})
	collection.InsertOne(context.Background(), models.Category{Name: "Apparel"})
	collection.InsertOne(context.Background(), models.Category{Name: "Spells"})

	router := setupCategoryRouter()
	req, _ := http.NewRequest("GET", "/api/categories", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var got []models.Category
	json.Unmarshal(w.Body.Bytes(), &got)
	assert.Len(t, got, 3)
	assert.Equal(t, "Apparel", got[0].Name)
	assert.Equal(t, "Spells", got[1].Name)
	assert.Equal(t, "Weapons", got[2].Name)
}

// --- CreateCategory ---

func TestCreateCategory_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")

	router := setupCategoryRouter()
	body, _ := json.Marshal(map[string]string{"name": "Potions"})
	req, _ := http.NewRequest("POST", "/api/admin/categories", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Persisted in Mongo
	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")
	var saved models.Category
	err := collection.FindOne(context.Background(), bson.M{"name": "Potions"}).Decode(&saved)
	assert.NoError(t, err)
	assert.False(t, saved.ID.IsZero())
}

func TestCreateCategory_MissingName(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")

	router := setupCategoryRouter()
	body, _ := json.Marshal(map[string]interface{}{})
	req, _ := http.NewRequest("POST", "/api/admin/categories", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateCategory_DuplicateName(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")
	collection.InsertOne(context.Background(), models.Category{Name: "Weapons"})

	router := setupCategoryRouter()
	body, _ := json.Marshal(map[string]string{"name": "Weapons"})
	req, _ := http.NewRequest("POST", "/api/admin/categories", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "already exists")
}

func TestCreateCategory_RejectsCustomer(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")

	router := setupCategoryRouter()
	body, _ := json.Marshal(map[string]string{"name": "Spells"})
	req, _ := http.NewRequest("POST", "/api/admin/categories", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateCategory_RejectsUnauthenticated(t *testing.T) {
	router := setupCategoryRouter()
	body, _ := json.Marshal(map[string]string{"name": "Spells"})
	req, _ := http.NewRequest("POST", "/api/admin/categories", bytes.NewBuffer(body))
	// no Authorization header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// --- DeleteCategory ---

func TestDeleteCategory_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")
	result, err := collection.InsertOne(context.Background(), models.Category{Name: "Trinkets"})
	assert.NoError(t, err)
	id := result.InsertedID.(primitive.ObjectID)

	router := setupCategoryRouter()
	req, _ := http.NewRequest("DELETE", "/api/admin/categories/"+id.Hex(), nil)
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	// Verify gone
	var count int64
	count, _ = collection.CountDocuments(context.Background(), bson.M{"_id": id})
	assert.Equal(t, int64(0), count)
}

func TestDeleteCategory_NotFound(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")

	router := setupCategoryRouter()
	req, _ := http.NewRequest("DELETE", "/api/admin/categories/"+primitive.NewObjectID().Hex(), nil)
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteCategory_InvalidID(t *testing.T) {
	setupTestDB()
	ensureMongo()

	router := setupCategoryRouter()
	req, _ := http.NewRequest("DELETE", "/api/admin/categories/not-a-valid-id", nil)
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestDeleteCategory_RejectsCustomer(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")

	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")
	result, _ := collection.InsertOne(context.Background(), models.Category{Name: "Protected"})
	id := result.InsertedID.(primitive.ObjectID)

	router := setupCategoryRouter()
	req, _ := http.NewRequest("DELETE", "/api/admin/categories/"+id.Hex(), nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}
