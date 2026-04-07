package security

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateToken(userID uint, role string) (string, error) {
	secretKey := []byte(os.Getenv("JWT_SECRET"))

	//Create new token object
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	//Get complete encoded token as string
	tokenString, err := token.SignedString(secretKey)
	return tokenString, err
}
