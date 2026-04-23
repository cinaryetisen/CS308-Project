package tests

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"medieval-store/security"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(security.AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "success"})
	})
	return router
}

func TestAuthMiddleware_NoHeader(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header is missing")
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "JustTheTokenNoBearerFlag")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid Authorization header format")
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret_key")
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer faketoken123")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
