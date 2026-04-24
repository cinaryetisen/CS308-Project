package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"medieval-store/config"
	"medieval-store/controllers"
	"medieval-store/models"
	"medieval-store/security"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// Helper to set up protected router
func setupProtectedRouter() *gin.Engine {
	os.Setenv("JWT_SECRET", "test_secret")
	gin.SetMode(gin.TestMode)
	router := gin.New()

	protected := router.Group("/api")
	protected.Use(security.AuthMiddleware())
	protected.GET("/users/me", controllers.GetProfile)
	protected.PATCH("/users/me", controllers.UpdateProfile)

	return router
}

// Helper to generate a token for our tests
func getTestToken(userID uint, role string) string {
	token, _ := security.GenerateToken(userID, role)
	return "Bearer " + token
}
func TestGetProfile_Success(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	config.DB.Create(&models.User{ID: 1, Name: "Lancelot", Email: "lance@knight.com"})
	router := setupProtectedRouter()

	req, _ := http.NewRequest("GET", "/api/users/me", nil)
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Lancelot")
	assert.Contains(t, w.Body.String(), "\"password\":\"\"") // Ensure password value is wiped!
}
func TestGetProfile_Unauthorized(t *testing.T) {
	router := setupProtectedRouter()

	req, _ := http.NewRequest("GET", "/api/users/me", nil)
	// Notice we are NOT attaching an Authorization header here
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
func TestUpdateProfile_Success(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	config.DB.Create(&models.User{ID: 1, Name: "Old Name", HomeAddress: "Old Address"})
	router := setupProtectedRouter()

	updateData := map[string]string{"name": "Sir Lancelot", "home_address": "Camelot"}
	jsonValue, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PATCH", "/api/users/me", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(1, "customer"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it actually changed in the DB
	var user models.User
	config.DB.First(&user, 1)
	assert.Equal(t, "Sir Lancelot", user.Name)
	assert.Equal(t, "Camelot", user.HomeAddress)
}

func TestUpdateProfile_UserNotFound(t *testing.T) {
	setupTestDB()
	defer clearTestDB()
	router := setupProtectedRouter()

	updateData := map[string]string{"name": "Ghost"}
	jsonValue, _ := json.Marshal(updateData)

	req, _ := http.NewRequest("PATCH", "/api/users/me", bytes.NewBuffer(jsonValue))
	req.Header.Set("Authorization", getTestToken(999, "customer")) // ID 999 doesn't exist
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
