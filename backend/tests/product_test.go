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

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupProductRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	// Both routes are public — no auth middleware.
	router.GET("/api/products", controllers.GetProducts)
	router.GET("/api/products/:id", controllers.GetProduct)
	return router
}

// seedProduct inserts a single product into the test Mongo and returns its ObjectID.
func seedProduct(p models.Product) primitive.ObjectID {
	if p.ID.IsZero() {
		p.ID = primitive.NewObjectID()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), p)
	return p.ID
}

// ==========================================
// GetProduct (single)
// ==========================================

func TestGetProduct_Success(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	id := seedProduct(models.Product{
		Name:        "Iron Sword",
		Model:       "SWD-001",
		Price:       150.00,
		Quantity:    10,
		Category:    "Weapons",
		Description: "A sturdy blade.",
	})

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products/"+id.Hex(), nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Iron Sword")
	assert.Contains(t, w.Body.String(), "SWD-001")
}

func TestGetProduct_NotFound(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	// Valid ObjectID format, no matching document
	bogusID := primitive.NewObjectID().Hex()

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products/"+bogusID, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetProduct_InvalidID(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products/not-a-real-id", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==========================================
// GetProducts (list with search/sort/filter)
// ==========================================

func TestGetProducts_All(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	seedProduct(models.Product{Name: "Sword", Price: 100, Category: "Weapons"})
	seedProduct(models.Product{Name: "Shield", Price: 50, Category: "Apparel"})
	seedProduct(models.Product{Name: "Potion", Price: 5, Category: "Spells"})

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var products []models.Product
	json.Unmarshal(w.Body.Bytes(), &products)
	assert.Len(t, products, 3)
}

func TestGetProducts_EmptyArrayNotNull(t *testing.T) {
	// Frontend iterates the response; null would crash. Empty must be `[]`.
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

func TestGetProducts_SearchByName(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	seedProduct(models.Product{Name: "Iron Sword", Description: "A blade.", Category: "Weapons"})
	seedProduct(models.Product{Name: "Mana Potion", Description: "Restores energy.", Category: "Spells"})

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products?search=Sword", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var products []models.Product
	json.Unmarshal(w.Body.Bytes(), &products)
	assert.Len(t, products, 1)
	assert.Equal(t, "Iron Sword", products[0].Name)
}

func TestGetProducts_SearchByDescription(t *testing.T) {
	// This is the literal demo Step 3.6 path: search Product C "by description".
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	seedProduct(models.Product{
		Name:        "Tome of Ancient Lore",
		Description: "A dusty, leather-bound book containing forgotten history.",
		Category:    "Spells",
	})
	seedProduct(models.Product{
		Name:        "Iron Sword",
		Description: "A sturdy blade.",
		Category:    "Weapons",
	})

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products?search=forgotten", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var products []models.Product
	json.Unmarshal(w.Body.Bytes(), &products)
	assert.Len(t, products, 1)
	assert.Equal(t, "Tome of Ancient Lore", products[0].Name)
}

func TestGetProducts_SearchCaseInsensitive(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	seedProduct(models.Product{Name: "Iron Sword", Description: "A blade."})

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products?search=sword", nil) // lowercase
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var products []models.Product
	json.Unmarshal(w.Body.Bytes(), &products)
	assert.Len(t, products, 1)
	assert.Equal(t, "Iron Sword", products[0].Name)
}

func TestGetProducts_SortByPriceAsc(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	seedProduct(models.Product{Name: "Mid", Price: 100})
	seedProduct(models.Product{Name: "Cheap", Price: 50})
	seedProduct(models.Product{Name: "Premium", Price: 200})

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products?sort=price_asc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var products []models.Product
	json.Unmarshal(w.Body.Bytes(), &products)
	assert.Len(t, products, 3)
	assert.Equal(t, "Cheap", products[0].Name)
	assert.Equal(t, "Mid", products[1].Name)
	assert.Equal(t, "Premium", products[2].Name)
}

func TestGetProducts_SortByPriceDesc(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	seedProduct(models.Product{Name: "Mid", Price: 100})
	seedProduct(models.Product{Name: "Cheap", Price: 50})
	seedProduct(models.Product{Name: "Premium", Price: 200})

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products?sort=price_desc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var products []models.Product
	json.Unmarshal(w.Body.Bytes(), &products)
	assert.Len(t, products, 3)
	assert.Equal(t, "Premium", products[0].Name)
	assert.Equal(t, "Mid", products[1].Name)
	assert.Equal(t, "Cheap", products[2].Name)
}

func TestGetProducts_SortByPriceWithDiscount(t *testing.T) {
	// The aggregation computes tmp_sort_price = price - price*(discount/100).
	// A: 100 with 50% off → effective 50.
	// B: 60 with no discount → effective 60.
	// price_asc must put A before B even though A's listed price is higher.
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	seedProduct(models.Product{Name: "Discounted Premium", Price: 100, Discount: 50})
	seedProduct(models.Product{Name: "Regular Mid", Price: 60, Discount: 0})

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products?sort=price_asc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var products []models.Product
	json.Unmarshal(w.Body.Bytes(), &products)
	assert.Len(t, products, 2)
	assert.Equal(t, "Discounted Premium", products[0].Name)
	assert.Equal(t, "Regular Mid", products[1].Name)
}

func TestGetProducts_SortByPopularity(t *testing.T) {
	// popularity_score = rating * review_count, sorted descending.
	// A: 5*10 = 50. B: 4*50 = 200. C: 1*10 = 10. Expect order B, A, C.
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	seedProduct(models.Product{Name: "MidPopular", Rating: 5.0, ReviewCount: 10})
	seedProduct(models.Product{Name: "VeryPopular", Rating: 4.0, ReviewCount: 50})
	seedProduct(models.Product{Name: "Niche", Rating: 1.0, ReviewCount: 10})

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products?sort=popular", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var products []models.Product
	json.Unmarshal(w.Body.Bytes(), &products)
	assert.Len(t, products, 3)
	assert.Equal(t, "VeryPopular", products[0].Name)
	assert.Equal(t, "MidPopular", products[1].Name)
	assert.Equal(t, "Niche", products[2].Name)
}

func TestGetProducts_FilterByCategory(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	seedProduct(models.Product{Name: "Iron Sword", Category: "Weapons"})
	seedProduct(models.Product{Name: "Steel Halberd", Category: "Weapons"})
	seedProduct(models.Product{Name: "Mana Potion", Category: "Spells"})
	seedProduct(models.Product{Name: "Leather Boots", Category: "Apparel"})

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products?category=Weapons", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var products []models.Product
	json.Unmarshal(w.Body.Bytes(), &products)
	assert.Len(t, products, 2)
	for _, p := range products {
		assert.Equal(t, "Weapons", p.Category)
	}
}

func TestGetProducts_SearchAndCategoryCombined(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	seedProduct(models.Product{Name: "Mana Potion", Description: "Restores energy", Category: "Spells"})
	seedProduct(models.Product{Name: "Healing Potion", Description: "Restores HP", Category: "Spells"})
	seedProduct(models.Product{Name: "Potion Bottle", Description: "Empty container", Category: "Accessories"}) // matches "potion" but wrong category

	router := setupProductRouter()
	req, _ := http.NewRequest("GET", "/api/products?search=potion&category=Spells", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var products []models.Product
	json.Unmarshal(w.Body.Bytes(), &products)
	assert.Len(t, products, 2)
	for _, p := range products {
		assert.Equal(t, "Spells", p.Category)
	}
}
