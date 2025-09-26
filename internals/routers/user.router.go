package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/handlers"
	"github.com/prospera/internals/middlewares"
	"github.com/prospera/internals/repositories"
	"github.com/redis/go-redis/v9"
)

func InitUserRouter(router *gin.Engine, db *pgxpool.Pool, rdb *redis.Client) {
	ur := repositories.NewUserRepository(db)
	uh := handlers.NewUserHandler(ur, rdb)

	userGroup := router.Group("/users")
	userGroup.Use(
		middlewares.Authentication,
	)

	userGroup.GET("/", uh.HandleGetUserInfo)
	userGroup.GET("/all", uh.HandlerGetAllUsers)
	userGroup.GET("/transactions", uh.GetUserHistoryTransactions)
	userGroup.DELETE("transactions/:id", uh.HandleSoftDeleteTransaction)        // Soft Delete
	userGroup.PATCH("/password", middlewares.Authentication, uh.ChangePassword) // Changer Password
}
