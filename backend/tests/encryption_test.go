package tests

import (
	"strings"
	"testing"

	"medieval-store/config"
	"medieval-store/models"
	"medieval-store/security"

	"github.com/stretchr/testify/assert"
)

// ==========================================
// Encrypt / Decrypt helper (F1)
// ==========================================

func TestEncryptDecrypt_RoundTrip(t *testing.T) {
	setupTestDB() // sets DATA_ENC_KEY for tests
	ciphertext, err := security.Encrypt("secret tax id")
	assert.NoError(t, err)
	assert.NotEqual(t, "secret tax id", ciphertext)
	assert.True(t, strings.HasPrefix(ciphertext, "enc:"))

	plaintext, err := security.Decrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, "secret tax id", plaintext)
}

func TestEncrypt_EmptyStringPassesThrough(t *testing.T) {
	setupTestDB()
	out, err := security.Encrypt("")
	assert.NoError(t, err)
	assert.Equal(t, "", out)
}

func TestDecrypt_EmptyStringPassesThrough(t *testing.T) {
	setupTestDB()
	out, err := security.Decrypt("")
	assert.NoError(t, err)
	assert.Equal(t, "", out)
}

func TestEncrypt_IsIdempotent(t *testing.T) {
	// Re-encrypting an already-encrypted value must not double-wrap it
	// (otherwise BeforeSave running twice on the same struct would corrupt data).
	setupTestDB()
	ciphertext, _ := security.Encrypt("hello")
	ciphertext2, err := security.Encrypt(ciphertext)
	assert.NoError(t, err)
	assert.Equal(t, ciphertext, ciphertext2)
}

func TestDecrypt_LegacyPlaintextPassesThrough(t *testing.T) {
	// Values without the "enc:" prefix pass through so pre-F1 rows still work.
	setupTestDB()
	out, err := security.Decrypt("legacy plaintext value")
	assert.NoError(t, err)
	assert.Equal(t, "legacy plaintext value", out)
}

func TestEncrypt_NonceVariesAcrossCalls(t *testing.T) {
	// Same plaintext encrypted twice must produce different ciphertexts
	// (random GCM nonce). This is a baseline security property.
	setupTestDB()
	a, _ := security.Encrypt("repeat me")
	b, _ := security.Encrypt("repeat me")
	assert.NotEqual(t, a, b)
}

func TestDecrypt_TamperedCiphertextFails(t *testing.T) {
	setupTestDB()
	ciphertext, _ := security.Encrypt("treasure map")
	// Flip the last character of the base64 body to break the GCM tag.
	tampered := ciphertext[:len(ciphertext)-1] + "X"
	if tampered == ciphertext {
		tampered = ciphertext[:len(ciphertext)-1] + "Y"
	}
	_, err := security.Decrypt(tampered)
	assert.Error(t, err)
}

// ==========================================
// GORM hooks — end-to-end at-rest encryption
// ==========================================

func TestUser_PIIIsEncryptedAtRest(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	user := models.User{
		Name:        "Arthur",
		Email:       "arthur@camelot.com",
		TaxID:       "TAX-12345",
		HomeAddress: "Castle Camelot, England",
		Password:    "irrelevant",
	}
	assert.NoError(t, config.DB.Create(&user).Error)

	// AfterSave restored plaintext on the struct.
	assert.Equal(t, "TAX-12345", user.TaxID)
	assert.Equal(t, "Castle Camelot, England", user.HomeAddress)

	// Raw column holds ciphertext (bypass hooks via Raw + Scan).
	var rawTaxID, rawAddress string
	config.DB.Raw("SELECT tax_id, home_address FROM users WHERE id = ?", user.ID).
		Row().Scan(&rawTaxID, &rawAddress)
	assert.True(t, strings.HasPrefix(rawTaxID, "enc:"), "tax_id column must be encrypted")
	assert.True(t, strings.HasPrefix(rawAddress, "enc:"), "home_address column must be encrypted")
	assert.NotEqual(t, "TAX-12345", rawTaxID)

	// Re-fetch through GORM — AfterFind decrypts.
	var fetched models.User
	assert.NoError(t, config.DB.First(&fetched, user.ID).Error)
	assert.Equal(t, "TAX-12345", fetched.TaxID)
	assert.Equal(t, "Castle Camelot, England", fetched.HomeAddress)
}

func TestOrder_DeliveryAddressIsEncryptedAtRest(t *testing.T) {
	setupTestDB()
	defer clearTestDB()

	order := models.Order{
		CustomerID:      1,
		TotalPrice:      100,
		DeliveryAddress: "Knight's Lane 5, Camelot",
	}
	assert.NoError(t, config.DB.Create(&order).Error)
	assert.Equal(t, "Knight's Lane 5, Camelot", order.DeliveryAddress)

	var raw string
	config.DB.Raw("SELECT delivery_address FROM orders WHERE id = ?", order.ID).
		Row().Scan(&raw)
	assert.True(t, strings.HasPrefix(raw, "enc:"), "delivery_address must be encrypted at rest")

	var fetched models.Order
	assert.NoError(t, config.DB.First(&fetched, order.ID).Error)
	assert.Equal(t, "Knight's Lane 5, Camelot", fetched.DeliveryAddress)
}
