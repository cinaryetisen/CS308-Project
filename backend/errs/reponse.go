package errs

import (
	"github.com/gin-gonic/gin"
)

// Abort finds the error in the registry and terminates the request
func Abort(c *gin.Context, code Code) {
	meta, exists := registry[code]
	if !exists {
		meta = registry[InternalError]
	}

	c.AbortWithStatusJSON(meta.status, gin.H{
		"code":    code,
		"message": meta.message,
	})
}
