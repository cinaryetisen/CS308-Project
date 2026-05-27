package security

import (
	"slices"

	"medieval-store/errs"

	"github.com/gin-gonic/gin"
)

// Authorize checks if the user has one of the required roles
func Authorize(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get role from context (extracted from JWT by Auth middleware)
		userRole, exists := c.Get("role")
		if !exists {
			errs.Abort(c, errs.UserForbidden)
			return
		}

		// 2. Check if the user's role is in the allowed list
		authorized := slices.Contains(allowedRoles, userRole.(string))

		if !authorized {
			errs.Abort(c, errs.UserForbidden)
			return
		}

		c.Next()
	}
}
