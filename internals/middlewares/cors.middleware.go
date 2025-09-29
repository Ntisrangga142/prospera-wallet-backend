package middlewares

import (
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware(ctx *gin.Context) {
	// Daftar origin yang diizinkan
	whitelist := []string{
		"http://localhost:5173",
		"http://localhost",
		"http://localhost:80",
		"http://frontend",
		"http://frontend:80",
	}
	origin := ctx.GetHeader("Origin")
	if slices.Contains(whitelist, origin) {
		ctx.Header("Access-Control-Allow-Origin", origin)

	}
	// ctx.Header("Access-Control-Allow-Origin", "*")

	// Header CORS standar
	ctx.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, PUT, DELETE, OPTIONS")
	ctx.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")

	// Jika request adalah preflight (OPTIONS)
	if ctx.Request.Method == http.MethodOptions {
		ctx.AbortWithStatus(http.StatusNoContent)
		return
	}

	// Lanjutkan ke handler berikutnya
	ctx.Next()
}
