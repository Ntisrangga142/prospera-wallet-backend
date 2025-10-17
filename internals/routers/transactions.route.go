package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/handlers"
	"github.com/prospera/internals/middlewares"
	"github.com/prospera/internals/repositories"
	"github.com/redis/go-redis/v9"
)

func InitTransactions(router *gin.Engine, db *pgxpool.Pool, rdb *redis.Client) {
	repo := repositories.NewTransactionRepository(db)
	repoAuth := repositories.NewAuthRepo(db, rdb)
	handler := handlers.NewTransactionHandler(repo, rdb, repoAuth)

	transaction := router.Group("/transaction")
	transaction.Use(middlewares.Authentication)

	transaction.POST("", handler.CreateTransaction)
}
