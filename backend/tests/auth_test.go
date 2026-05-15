package tests

import (
	"bytes"
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
)

// Helper function to create a fresh Gin router for each test
func setupAuthRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/signup", controllers.Signup)
	router.POST("/api/login", controllers.Login)
	return router
}

// ==========================================
// SIGNUP TESTS
// ==========================================

func TestSignup_Success(t *testing.T) {
	setupTestDB()       // Spin up the fake database
	defer clearTestDB() // Wipe it clean when the test finishes

	router := setupAuthRouter()

	// The mock JSON data a frontend would send
	signupData := map[string]string{
		"name":         "Arthur Pendragon",
		"email":        "arthur@camelot.com",
		"password":     "Excalibur123!",
		"tax_id":       "9999",
		"home_address": "Camelot Castle",
	}
	jsonValue, _ := json.Marshal(signupData)

	// Create a fake HTTP request
	req, _ := http.NewRequest("POST", "/api/signup", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Send it to our router
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Contains(t, w.Body.String(), "User registered successfully")

	// Verify the user actually exists in the SQLite database
	var user models.User
	config.DB.Where("email = ?", "arthur@camelot.com").First(&user)
	assert.Equal(t, "Arthur Pendragon", user.Name)
}

func TestSignup_DuplicateEmail(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	// Inject a user into the DB directly
	config.DB.Create(&models.User{
		Name:     "Existing User",
		Email:    "taken@email.com",
		Password: "hashedpassword",
	})

	router := setupAuthRouter()
	signupData := map[string]string{
		"name":     "New User",
		"email":    "taken@email.com", // Same email!
		"password": "Password123!",
	}
	jsonValue, _ := json.Marshal(signupData)

	req, _ := http.NewRequest("POST", "/api/signup", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should fail with a 409 error because the email is taken
	assert.Equal(t, http.StatusConflict, w.Code)
}

func TestSignup_MissingFields(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupAuthRouter()
	// Missing the required "email" field
	signupData := map[string]string{
		"name":     "Anonymous",
		"password": "Password123!",
	}
	jsonValue, _ := json.Marshal(signupData)

	req, _ := http.NewRequest("POST", "/api/signup", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Should fail the Gin binding validation
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// ==========================================
// LOGIN TESTS
// ==========================================

func TestLogin_Success(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	// Create a user with a properly hashed password in our fake DB
	hashedPassword, _ := security.HashPassword("MySecret123!")
	config.DB.Create(&models.User{
		Name:     "Merlin",
		Email:    "merlin@magic.com",
		Password: hashedPassword,
		Role:     "customer",
	})

	router := setupAuthRouter()

	// Attempt to log in with the plaintext password
	loginData := map[string]string{
		"email":    "merlin@magic.com",
		"password": "MySecret123!",
	}
	jsonValue, _ := json.Marshal(loginData)

	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token") // Should return a JWT
}

func TestLogin_WrongPassword(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	hashedPassword, _ := security.HashPassword("MySecret123!")
	config.DB.Create(&models.User{Email: "merlin@magic.com", Password: hashedPassword})

	router := setupAuthRouter()
	loginData := map[string]string{
		"email":    "merlin@magic.com",
		"password": "WrongPassword!!", // Wrong!
	}
	jsonValue, _ := json.Marshal(loginData)

	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestLogin_UserNotFound(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	router := setupAuthRouter()
	loginData := map[string]string{
		"email":    "ghost@nobody.com", // Doesn't exist in DB
		"password": "Password123!",
	}
	jsonValue, _ := json.Marshal(loginData)

	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ==========================================
// VALIDATION TESTS
// ==========================================

func TestSignup_WeakPassword(t *testing.T) {
	// `binding:"min=6"` on Password — anything shorter must be rejected at validation time.
	setupTestDB()
	defer clearTestDB()

	router := setupAuthRouter()
	signupData := map[string]string{
		"name":     "Galahad",
		"email":    "galahad@camelot.com",
		"password": "short", // 5 chars
	}
	jsonValue, _ := json.Marshal(signupData)

	req, _ := http.NewRequest("POST", "/api/signup", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify the user was NOT created
	var count int64
	config.DB.Model(&models.User{}).Where("email = ?", "galahad@camelot.com").Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestLogin_InvalidEmailFormat(t *testing.T) {
	// `binding:"email"` on Email — non-email-format input must be rejected before any DB lookup.
	setupTestDB()
	defer clearTestDB()

	router := setupAuthRouter()
	loginData := map[string]string{
		"email":    "not-an-email-address",
		"password": "Password123!",
	}
	jsonValue, _ := json.Marshal(loginData)

	req, _ := http.NewRequest("POST", "/api/login", bytes.NewBuffer(jsonValue))
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// errs.AuthInvalidEmail maps to 400 — see backend/errs/registry.go.
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
