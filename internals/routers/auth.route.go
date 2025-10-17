package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/handlers"
	"github.com/prospera/internals/middlewares"
	"github.com/prospera/internals/repositories"
	"github.com/redis/go-redis/v9"
)

func InitAuthRoutes(router *gin.Engine, db *pgxpool.Pool, rdb *redis.Client) {
	repo := repositories.NewAuthRepo(db, rdb)
	handler := handlers.NewAuthHandler(repo, rdb)

	auth := router.Group("/auth")

	// Login
	auth.POST("", handler.Login)

	// Register
	auth.POST("/register", handler.Register)

	// Logout
	auth.DELETE("", middlewares.Authentication, handler.Logout)

	// Forgot Password and PIN
	auth.POST("/forgot", handler.Forgot)

	// Reset PIN
	auth.POST("/reset-pin", handler.ResetPIN)

	// Reset Password
	auth.POST("/reset-password", handler.ResetPassword)

	// Create PIN
	auth.POST("/pin", middlewares.Authentication, handler.UpdatePIN)

	// Change PIN (used in profile/change)
	auth.POST("/change-pin", middlewares.Authentication, handler.ChangePIN)

	// Verify PIN
	auth.POST("/verify-pin", middlewares.Authentication, handler.VerifyPIN)

	// // Verify email
	// auth.GET("/verify-password", handler.CheckEmail)

}
