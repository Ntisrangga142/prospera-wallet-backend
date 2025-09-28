package middlewares

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(ctx *gin.Context) {
	// Specific allowed origins
	whitelist := []string{
		"http://127.0.0.1:5500",
		"http://localhost:5173",
		"http://localhost",
		"http://localhost:80",
		"http://frontend",
		"http://frontend:80",
	}

	origin := ctx.GetHeader("Origin")
	allowed := false

	// Check specific whitelist first
	if slices.Contains(whitelist, origin) {
		allowed = true
	} else {
		// Allow any origin from local networks
		localNetworks := []string{
			"http://192.168.", // 192.168.x.x networks
			"http://10.",      // 10.x.x.x networks
			"http://172.",     // 172.x.x.x networks
		}

		for _, network := range localNetworks {
			if strings.HasPrefix(origin, network) {
				allowed = true
				break
			}
		}
	}

	if allowed {
		ctx.Header("Access-Control-Allow-Origin", origin)
	} else {
		// Log rejected origins for debugging
		println("CORS origin rejected:", origin)
	}

	// Header CORS standar
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
	ctx.Header("Access-Control-Allow-Credentials", "true")

	// Jika request adalah preflight (OPTIONS)
	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}

	// Lanjutkan ke handler berikutnya
	ctx.Next()
}
