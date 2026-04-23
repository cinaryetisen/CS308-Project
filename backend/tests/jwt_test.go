package tests

import (
	"github.com/stretchr/testify/assert"
	"medieval-store/security"
	"os"
	"testing"
)

func TestGenerateToken_Success(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret_key") // Mock the env variable

	token, err := security.GenerateToken(1, "customer")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
func TestGenerateToken_DifferentUsers(t *testing.T) {
	os.Setenv("JWT_SECRET", "test_secret_key")

	token1, _ := security.GenerateToken(1, "customer")
	token2, _ := security.GenerateToken(2, "manager")

	assert.NotEqual(t, token1, token2, "Tokens for different users should not match")
}
