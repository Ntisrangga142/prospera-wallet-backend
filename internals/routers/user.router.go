package routers

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/handlers"
	"github.com/prospera/internals/middlewares"
	"github.com/prospera/internals/repositories"
)

func InitUserRouter(router *gin.Engine, db *pgxpool.Pool) {
	ur := repositories.NewUserRepository(db)
	uh := handlers.NewUserHandler(ur)

	userGroup := router.Group("/users")
	userGroup.Use(
		middlewares.Authentication,
	)

	userGroup.GET("/", uh.HandleGetUserInfo)
	userGroup.GET("/all", uh.HandlerGetAllUsers)
	userGroup.GET("/transactions", uh.HandleGetUserTransactionsHistory)
	userGroup.DELETE("transactions/:id", uh.HandleSoftDeleteTransaction)
	userGroup.PATCH("/password", uh.ChangePassword)
}
