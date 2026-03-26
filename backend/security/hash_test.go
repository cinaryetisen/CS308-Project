package security

import (
	"testing"
)

func TestHash(t *testing.T) {
	password := "Password1234**"
	hash, err := HashPassword(password)

	if err != nil {
		t.Errorf("Error while hashing: %s", err)
	}

	res := CheckPasswordHash(password, hash)
	if !res {
		t.Errorf("Passsword and hash do not match!")
	}
}
