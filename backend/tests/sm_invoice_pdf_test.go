package tests

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"medieval-store/config"
	"medieval-store/controllers"
	"medieval-store/models"
	"medieval-store/security"
	"medieval-store/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// B10: sales managers can export invoice PDFs — single and save-all bulk.

func setupSMInvoiceRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	protected := router.Group("/api")
	protected.Use(security.AuthMiddleware())
	protected.GET("/orders/:id/invoice", controllers.DownloadInvoice)

	sm := router.Group("/api/admin")
	sm.Use(security.AuthMiddleware(), security.Authorize("sales_manager"))
	sm.GET("/invoices", controllers.GetInvoicesByDateRange)

	return router
}

func TestDownloadInvoice_SalesManagerAllowed(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()

	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})
	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, DeliveryAddress: "Camelot"})
	config.DB.Create(&models.OrderItem{OrderID: 1, ProductID: primitive.NewObjectID().Hex(), Quantity: 1, Price: 100})

	router := setupSMInvoiceRouter()
	req, _ := http.NewRequest("GET", "/api/orders/1/invoice", nil)
	req.Header.Set("Authorization", getTestToken(99, "sales_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "sales manager must be able to fetch any invoice PDF")
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.NotEmpty(t, w.Body.Bytes())
}

func TestDownloadInvoice_OtherCustomerStillForbidden(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	config.DB.Create(&models.Order{ID: 1, CustomerID: 1})

	router := setupSMInvoiceRouter()
	req, _ := http.NewRequest("GET", "/api/orders/1/invoice", nil)
	req.Header.Set("Authorization", getTestToken(2, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetInvoices_BulkPDFExport(t *testing.T) {
	setupTestDB()
	ensureMongo()
	defer clearTestDB()
	defer clearMongoCollection("products")

	config.DB.Create(&models.User{ID: 1, Name: "Arthur", Email: "arthur@camelot.com"})

	productID := primitive.NewObjectID()
	collection := config.MongoClient.Database(config.MongoDBName).Collection("products")
	collection.InsertOne(context.Background(), models.Product{ID: productID, Name: "Bulk Sword", Price: 10})

	// Two orders inside the range, one outside.
	for i, daysAgo := range []int{2, 3, 40} {
		order := models.Order{
			CustomerID: 1, TotalPrice: 10, DeliveryAddress: "Camelot",
			Status: "delivered", CreatedAt: time.Now().AddDate(0, 0, -daysAgo),
		}
		config.DB.Create(&order)
		config.DB.Create(&models.OrderItem{OrderID: order.ID, ProductID: productID.Hex(), Quantity: i + 1, Price: 10})
	}

	from := time.Now().AddDate(0, 0, -7).Format("2006-01-02")
	to := time.Now().Format("2006-01-02")

	router := setupSMInvoiceRouter()
	req, _ := http.NewRequest("GET", "/api/admin/invoices?from="+from+"&to="+to+"&format=pdf", nil)
	req.Header.Set("Authorization", getTestToken(99, "sales_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
	assert.Contains(t, w.Header().Get("Content-Disposition"), "invoices_")
	assert.NotEmpty(t, w.Body.Bytes())
	assert.Equal(t, "%PDF", string(w.Body.Bytes()[:4]), "response must be a real PDF document")
}

func TestGetInvoices_JSONStillDefault(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	config.DB.Create(&models.Order{ID: 1, CustomerID: 1, DeliveryAddress: "Camelot", CreatedAt: time.Now()})

	from := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	to := time.Now().Format("2006-01-02")

	router := setupSMInvoiceRouter()
	req, _ := http.NewRequest("GET", "/api/admin/invoices?from="+from+"&to="+to, nil)
	req.Header.Set("Authorization", getTestToken(99, "sales_manager"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
}

func TestGenerateBulkInvoicePDF_EmptyStillValidPDF(t *testing.T) {
	pdfBytes, err := services.GenerateBulkInvoicePDF(nil)
	assert.NoError(t, err)
	assert.Equal(t, "%PDF", string(pdfBytes[:4]))
}
