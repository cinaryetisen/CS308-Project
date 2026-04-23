package security

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

// Authorize checks if the user has one of the required roles
func Authorize(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get role from context (extracted from JWT by Auth middleware)
		userRole, exists := c.Get("role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Role not found"})
			return
		}

		// 2. Check if the user's role is in the allowed list
		authorized := slices.Contains(allowedRoles, userRole.(string))

		if !authorized {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You don't have permission to perform this action"})
			return
		}

		c.Next()
	}
}
