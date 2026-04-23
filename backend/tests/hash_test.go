package tests

import (
	"github.com/stretchr/testify/assert"
	"medieval-store/security"
	"testing"
)

func TestHashPassword_Success(t *testing.T) {
	password := "Password1234**"
	hash, err := security.HashPassword(password)

	assert.NoError(t, err)
	assert.NotEmpty(t, hash)

	match := security.CheckPasswordHash(password, hash)
	assert.True(t, match)
}

func TestHashPassword_Mismatch(t *testing.T) {
	password := "Password1234**"
	hash, _ := security.HashPassword(password)

	match := security.CheckPasswordHash("WrongPassword!", hash)
	assert.False(t, match)
}

func TestHashPassword_EmptyString(t *testing.T) {
	hash, err := security.HashPassword("")
	assert.NoError(t, err)

	match := security.CheckPasswordHash("", hash)
	assert.True(t, match)
}
