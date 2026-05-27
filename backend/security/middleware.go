package security

import (
	"fmt"
	"os"
	"strings"

	"medieval-store/errs"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Function to intercept requests to check for a valid JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			errs.Abort(c, errs.AuthMissingHeader)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			errs.Abort(c, errs.AuthInvalidHeader)
			return
		}

		tokenString := parts[1]
		secretKey := []byte(os.Getenv("JWT_SECRET"))

		//Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			errs.Abort(c, errs.AuthInvalidToken)
			return
		}

		//Extract the data and attach it to Gin Context
		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			c.Set("user_id", uint(claims["user_id"].(float64)))
			c.Set("role", claims["role"].(string))
		} else {
			errs.Abort(c, errs.AuthClaimsParseFail)
			return
		}

		c.Next()
	}
}
