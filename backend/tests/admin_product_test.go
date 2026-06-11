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

func setupAdminProductRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	pm := router.Group("/api/admin")
	pm.Use(security.AuthMiddleware(), security.Authorize("product_manager"))
	pm.POST("/products", controllers.CreateProduct)
	pm.PATCH("/products/:id", controllers.UpdateProduct)
	pm.DELETE("/products/:id", controllers.DeleteProduct)
	pm.PATCH("/products/:id/stock", controllers.UpdateStock)
	return router
}

// seedCategory registers a category so product create/update validation passes.
func seedCategory(t *testing.T, name string) {
	t.Helper()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("categories")
	collection.InsertOne(context.Background(), models.Category{Name: name})
}

func validProductPayload() map[string]interface{} {
	return map[string]interface{}{
		"name":          "Test Halberd",
		"model":         "TST-001",
		"serial_number": "SN-TEST-1",
		"description":   "A halberd for unit tests.",
		"quantity":      7,
		"category":      "Weapons",
		"distributor":   "Test Forge",
		"warranty":      "1 Year",
	}
}

// ==========================================
// CreateProduct (B1)
// ==========================================

func TestCreateProduct_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	seedCategory(t, "Weapons")
	defer clearMongoCollection("categories")
	defer clearMongoCollection("products")

	router := setupAdminProductRouter()
	body, _ := json.Marshal(validProductPayload())
	req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	// Lands in the CONFIGURED test database (regression check for the
	// hardcoded "medieval_store" bug), with the sentinel price applied.
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	var saved models.Product
	err := collection.FindOne(context.Background(), bson.M{"serial_number": "SN-TEST-1"}).Decode(&saved)
	assert.NoError(t, err)
	assert.Equal(t, "Test Halberd", saved.Name)
	assert.Equal(t, 7, saved.Quantity)
	assert.Equal(t, 99999.99, saved.Price)
}

func TestCreateProduct_MissingRequiredFields(t *testing.T) {
	setupTestDB()
	ensureMongo()

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{"name": "No other fields"})
	req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateProduct_RejectsCustomer(t *testing.T) {
	setupTestDB()
	ensureMongo()
	seedCategory(t, "Weapons")
	defer clearMongoCollection("categories")

	router := setupAdminProductRouter()
	body, _ := json.Marshal(validProductPayload())
	req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateProduct_RejectsSalesManager(t *testing.T) {
	// Separation of duties: catalog CRUD belongs to the product manager.
	setupTestDB()
	ensureMongo()

	router := setupAdminProductRouter()
	body, _ := json.Marshal(validProductPayload())
	req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "sales_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ==========================================
// UpdateProduct (B2)
// ==========================================

func TestUpdateProduct_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	seedCategory(t, "Weapons")
	defer clearMongoCollection("categories")
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{
		ID: productID, Name: "Old Name", Model: "M-1", SerialNumber: "SN-1",
		Description: "Old description", Category: "Weapons", Distributor: "D", Warranty: "W",
		Price: 100, Quantity: 5,
	})

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{
		"name": "New Name", "model": "M-1", "serial_number": "SN-1",
		"description": "New description", "category": "Weapons",
		"distributor": "D", "warranty": "W",
	})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+productID.Hex(), bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var saved models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&saved)
	assert.Equal(t, "New Name", saved.Name)
	assert.Equal(t, "New description", saved.Description)
	// Price/quantity are locked out of this endpoint
	assert.Equal(t, 100.0, saved.Price)
	assert.Equal(t, 5, saved.Quantity)
}

func TestUpdateProduct_PartialUpdateKeepsOtherFields(t *testing.T) {
	// B08: PATCH with a single field must not blank out the omitted ones.
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{
		ID: productID, Name: "Original Name", Description: "Original description",
		Category: "Weapons", Distributor: "Forge", Warranty: "1 Year",
		ImageURL: "http://img/x.png", Tags: []string{"keep", "me"},
	})

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{"name": "Renamed Only"})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+productID.Hex(), bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var saved models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&saved)
	assert.Equal(t, "Renamed Only", saved.Name)
	assert.Equal(t, "Original description", saved.Description, "omitted fields must survive")
	assert.Equal(t, "Weapons", saved.Category)
	assert.Equal(t, "Forge", saved.Distributor)
	assert.Equal(t, "1 Year", saved.Warranty)
	assert.Equal(t, "http://img/x.png", saved.ImageURL)
	assert.Equal(t, []string{"keep", "me"}, saved.Tags)
}

func TestUpdateProduct_EmptyBodyRejected(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: productID, Name: "Untouched"})

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+productID.Hex(), bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateProduct_NotFound(t *testing.T) {
	setupTestDB()
	ensureMongo()

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{"name": "Ghost"})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+primitive.NewObjectID().Hex(), bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ==========================================
// DeleteProduct (B3)
// ==========================================

func TestDeleteProduct_SoftDeletes(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: productID, Name: "Doomed"})

	router := setupAdminProductRouter()
	req, _ := http.NewRequest("DELETE", "/api/admin/products/"+productID.Hex(), nil)
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Document still exists but carries deleted_at
	var raw bson.M
	err := collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&raw)
	assert.NoError(t, err)
	assert.NotNil(t, raw["deleted_at"], "soft delete must set deleted_at")
}

func TestDeleteProduct_NotFound(t *testing.T) {
	setupTestDB()
	ensureMongo()

	router := setupAdminProductRouter()
	req, _ := http.NewRequest("DELETE", "/api/admin/products/"+primitive.NewObjectID().Hex(), nil)
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// ==========================================
// UpdateStock (B4)
// ==========================================

func TestUpdateStock_Increase(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: productID, Name: "Restockable", Quantity: 2})

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{"delta": 5})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+productID.Hex()+"/stock", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var saved models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&saved)
	assert.Equal(t, 7, saved.Quantity)
}

func TestUpdateStock_DecreaseBlockedWhenInsufficient(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: productID, Name: "Scarce", Quantity: 3})

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{"delta": -5})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+productID.Hex()+"/stock", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Filter requires quantity >= -delta, so nothing matches
	assert.NotEqual(t, http.StatusOK, w.Code)

	var saved models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&saved)
	assert.Equal(t, 3, saved.Quantity, "stock must be untouched")
}

func TestUpdateStock_DecreaseSuccess(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: productID, Name: "Plenty", Quantity: 10})

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{"delta": -4})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+productID.Hex()+"/stock", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var saved models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&saved)
	assert.Equal(t, 6, saved.Quantity)
}

func TestUpdateStock_RejectsCustomer(t *testing.T) {
	setupTestDB()
	ensureMongo()

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{"delta": 1})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+primitive.NewObjectID().Hex()+"/stock", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

// ==========================================
// Category validation (B12)
// ==========================================

func TestCreateProduct_UnknownCategoryRejected(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")
	defer clearMongoCollection("products")

	// No categories seeded — "Weapons" is unknown here.
	router := setupAdminProductRouter()
	body, _ := json.Marshal(validProductPayload())
	req, _ := http.NewRequest("POST", "/api/admin/products", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "CATEGORY_NOT_FOUND")

	// Nothing persisted
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	count, _ := collection.CountDocuments(context.Background(), bson.M{"serial_number": "SN-TEST-1"})
	assert.Equal(t, int64(0), count)
}

func TestUpdateProduct_UnknownCategoryRejected(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("categories")
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: productID, Name: "Stable", Category: "Weapons"})

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{"category": "Nonexistent"})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+productID.Hex(), bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "CATEGORY_NOT_FOUND")

	var saved models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&saved)
	assert.Equal(t, "Weapons", saved.Category, "category must be unchanged")
}

func TestUpdateProduct_KnownCategoryAccepted(t *testing.T) {
	setupTestDB()
	ensureMongo()
	seedCategory(t, "Spells")
	defer clearMongoCollection("categories")
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: productID, Name: "Recat", Category: "Weapons"})

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{"category": "Spells"})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+productID.Hex(), bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var saved models.Product
	collection.FindOne(context.Background(), bson.M{"_id": productID}).Decode(&saved)
	assert.Equal(t, "Spells", saved.Category)
}

// ==========================================
// UpdateStock error semantics
// ==========================================

func TestUpdateStock_NonexistentProductReturns404(t *testing.T) {
	setupTestDB()
	ensureMongo()

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{"delta": -1})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+primitive.NewObjectID().Hex()+"/stock", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code, "a missing product is 404, not 'out of stock'")
	assert.Contains(t, w.Body.String(), "PRODUCT_NOT_FOUND")
}

func TestUpdateStock_InsufficientStockStillConflict(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: productID, Name: "Thin Stock", Quantity: 1})

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{"delta": -5})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+productID.Hex()+"/stock", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Contains(t, w.Body.String(), "PRODUCT_OUT_OF_STOCK")
}

func TestUpdateStock_ZeroDeltaRejected(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearMongoCollection("products")

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: productID, Name: "No-op Target", Quantity: 3})

	router := setupAdminProductRouter()
	body, _ := json.Marshal(map[string]interface{}{"delta": 0})
	req, _ := http.NewRequest("PATCH", "/api/admin/products/"+productID.Hex()+"/stock", bytes.NewBuffer(body))
	req.Header.Set("Authorization", getTestToken(1, "product_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "non-zero")
}
