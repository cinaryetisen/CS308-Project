package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"medieval-store/security"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
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

// setupAuthorizedRouter builds a router that requires both auth AND a specific role.
// Handler echoes the user_id from context so we can assert it was set correctly.
func setupAuthorizedRouter(role string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	protected := router.Group("/")
	protected.Use(security.AuthMiddleware(), security.Authorize(role))
	protected.GET("/restricted", func(c *gin.Context) {
		uid, _ := c.Get("user_id")
		c.JSON(200, gin.H{"user_id": uid})
	})
	return router
}

// ==========================================
// FAILURE PATH TESTS (existing)
// ==========================================

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

// ==========================================
// HAPPY PATH + JWT ROUND-TRIP TESTS
// (security has no exported ValidateToken — round-trip is exercised through the middleware.)
// ==========================================

func TestAuthMiddleware_ValidToken_PassesThrough(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret_key")
	router := setupTestRouter()

	token, err := security.GenerateToken(1, "customer")
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
}

func TestAuthMiddleware_SetsUserContext(t *testing.T) {
	// Round-trip: the middleware must extract user_id and role from the token
	// and put them into the gin context for downstream handlers.
	os.Setenv("JWT_SECRET", "test_secret_key")

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(security.AuthMiddleware())
	router.GET("/whoami", func(c *gin.Context) {
		uid, uidExists := c.Get("user_id")
		role, roleExists := c.Get("role")
		c.JSON(200, gin.H{
			"user_id":      uid,
			"role":         role,
			"uid_present":  uidExists,
			"role_present": roleExists,
		})
	})

	token, _ := security.GenerateToken(42, "product_manager")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/whoami", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":42`)
	assert.Contains(t, w.Body.String(), `"role":"product_manager"`)
	assert.Contains(t, w.Body.String(), `"uid_present":true`)
	assert.Contains(t, w.Body.String(), `"role_present":true`)
}

func TestAuthMiddleware_TamperedToken(t *testing.T) {
	// Tampering with the payload must invalidate the token.
	os.Setenv("JWT_SECRET", "test_secret_key")
	router := setupTestRouter()

	token, _ := security.GenerateToken(1, "customer")
	// Mutate a character in the PAYLOAD segment (between the two dots) rather than
	// the last signature char — flipping the final base64 char is unreliable
	// because trailing bits can decode to the same signature bytes.
	parts := strings.Split(token, ".")
	payload := parts[1]
	flipped := payload[:len(payload)-1]
	if payload[len(payload)-1] == 'A' {
		flipped += "B"
	} else {
		flipped += "A"
	}
	parts[1] = flipped
	tampered := strings.Join(parts, ".")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tampered)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_WrongSecret(t *testing.T) {
	// Token signed with secret A; middleware validates with secret B → reject.
	os.Setenv("JWT_SECRET", "secretA")
	token, _ := security.GenerateToken(1, "customer")

	os.Setenv("JWT_SECRET", "secretB")
	router := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==========================================
// AUTHORIZE (RBAC) TESTS
// ==========================================

func TestAuthorize_AllowsCorrectRole(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret_key")
	router := setupAuthorizedRouter("product_manager")

	token, _ := security.GenerateToken(99, "product_manager")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/restricted", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":99`)
}

func TestAuthorize_RejectsWrongRole(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret_key")
	router := setupAuthorizedRouter("product_manager")

	// Customer trying to hit a PM-only route
	token, _ := security.GenerateToken(1, "customer")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/restricted", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "USER_FORBIDDEN")
}

func TestAuthorize_RejectsMissingAuth(t *testing.T) {
	// No Authorization header → AuthMiddleware blocks before Authorize runs (401, not 403).
	router := setupAuthorizedRouter("product_manager")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/restricted", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==========================================
// DEFENSIVE CLAIM PARSING (B19)
// A validly-signed token with missing/wrong-typed claims must be rejected
// cleanly (401), never panic the handler.
// ==========================================

func signCustomClaims(claims jwt.MapClaims) string {
	secret := []byte(os.Getenv("JWT_SECRET"))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString(secret)
	return s
}

func TestAuthMiddleware_MissingRoleClaim(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret_key")
	router := setupTestRouter()

	// Signed correctly but has no "role" claim.
	token := signCustomClaims(jwt.MapClaims{
		"user_id": float64(1),
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_MissingUserIDClaim(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret_key")
	router := setupTestRouter()

	token := signCustomClaims(jwt.MapClaims{
		"role": "customer",
		"exp":  time.Now().Add(time.Hour).Unix(),
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_WrongTypedRoleClaim(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret_key")
	router := setupTestRouter()

	// role is a number, not a string — must not panic.
	token := signCustomClaims(jwt.MapClaims{
		"user_id": float64(1),
		"role":    42,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
