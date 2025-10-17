package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/handlers"
	"github.com/prospera/internals/middlewares"
	"github.com/prospera/internals/repositories"
	"github.com/redis/go-redis/v9"
)

func InitInternalAccountRoute(router *gin.Engine, db *pgxpool.Pool, rdb *redis.Client) {
	repo := repositories.NewInternalAccountRepository(db)
	handler := handlers.NewInternalAccountHandler(repo, rdb)

	internal := router.Group("/internal")

	internal.Use(middlewares.Authentication)

	internal.GET("", handler.GetAll)
}
