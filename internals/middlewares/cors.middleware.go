package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/utils"
)

// CORSMiddleware handle CORS for frontend development
func CORSMiddleware() gin.HandlerFunc {
	// Daftar origin yang diizinkan
	whitelist := []string{
		"http://127.0.0.1:3000",
		"http://localhost:3000",
		"http://localhost:5173",
		"http://localhost:80",
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Jika origin ada dan termasuk whitelist
		allowed := false
		for _, o := range whitelist {
			if o == origin {
				allowed = true
				break
			}
		}

		if origin != "" && !allowed {
			utils.HandleMiddlewareError(c, http.StatusForbidden, "CORS origin not allowed", "CORS origin not allowed")
			return
		}

		// Set CORS headers
		if origin != "" && allowed {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "*")
		}

		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Expose-Headers", "Authorization, Content-Type")

		// Preflight request handling
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
