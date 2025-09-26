package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/middlewares"
	"github.com/redis/go-redis/v9"
)

func InitRouter(db *pgxpool.Pool, rdb *redis.Client) *gin.Engine {
	router := gin.Default()
	middlewares.InitRedis(rdb)

	router.Static("/profile", "public/profile")

	// Init Route Authentication
	InitAuthRoutes(router, db, rdb)
	InitUserRouter(router, db, rdb)
	InitTransactions(router, db, rdb)

	return router
}
