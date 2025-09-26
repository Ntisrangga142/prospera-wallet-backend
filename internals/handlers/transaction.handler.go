package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/models"
	"github.com/prospera/internals/repositories"
	"github.com/prospera/internals/utils"
	"github.com/redis/go-redis/v9"
)

type TransactionHandler struct {
	repo *repositories.TransactionRepository
	rdb  *redis.Client
}

func NewTransactionHandler(repo *repositories.TransactionRepository, rdb *redis.Client) *TransactionHandler {
	return &TransactionHandler{repo: repo, rdb: rdb}
}

// Create Transactions
func (h *TransactionHandler) CreateTransaction(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "invalid or missing token", err)
		return
	}

	var req models.TransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "invalid request payload", err)
		return
	}

	err = h.repo.CreateTransaction(ctx.Request.Context(), &req, uid)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to create transaction", err)
		return
	}

	var redisKey = fmt.Sprintf("Prospera-HistoryTransactions_%d", uid)
	if err := utils.InvalidateCache(ctx, h.rdb, redisKey); err != nil {
		log.Println("Failed invalidate cache:", err)
	}

	ctx.JSON(http.StatusOK, models.Response[any]{
		Success: true,
		Message: "Transaction created successfully",
	})
}
