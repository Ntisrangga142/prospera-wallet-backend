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
	handler := handlers.NewAuthHandler(repo)

	auth := router.Group("/auth")

	// Login
	auth.POST("", handler.Login)

	//Register
	auth.POST("/register", handler.Register)

	//Update PIN
	auth.POST("/pin", middlewares.Authentication, handler.UpdatePIN)

	// Logout
	auth.DELETE("/logout", middlewares.Authentication, handler.Logout)
}
