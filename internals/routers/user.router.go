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

	userGroup := router.Group("/user")

	userGroup.Use(middlewares.Authentication)

	// GET PROFILE
	userGroup.GET("", uh.GetProfile)

	// GET USER WITH ID
	userGroup.GET("/:id", uh.GetUserWithId)

	// PATCH PROFILE
	userGroup.PATCH("", uh.UpdateProfile)

	// GET BALANCE
	userGroup.GET("/wallet", uh.GetBalance)

	// GET ALL PROFILE
	userGroup.GET("/all", uh.GetAllUsers)

	// GET ALL HISTORY
	userGroup.GET("/history", uh.GetUserHistoryTransactions)

	// DELETE HISTORY
	userGroup.DELETE("/history/:id", uh.HandleSoftDeleteTransaction)

	// PATCH CHANGE PASSWORD
	userGroup.PATCH("/password", uh.ChangePassword)

	// GET SUMMARY
	userGroup.GET("/summary", uh.GetSummary)

	// REMOVE AVATAR
	userGroup.DELETE("/avatar", uh.RemoveAvatar)
}
