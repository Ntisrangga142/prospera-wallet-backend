package handlers

import (
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
		utils.HandleError(ctx, 401, "Unauthorized", "invalid token", err)
		return
	}

	var req models.TransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, 400, "Bad Request", "invalid payload", err)
		return
	}

	// Verifikasi PIN sebelum membuat transaksi
	verify, err := utils.VerifyUserPIN(ctx.Request.Context(), h.repo.DB, uid, req.PIN)
	if err != nil || !verify {
		utils.HandleError(ctx, 403, "Forbidden", "invalid PIN", err)
		return
	}

	if err := h.repo.CreateTransaction(ctx.Request.Context(), &req, uid); err != nil {
		utils.HandleError(ctx, 500, "Internal Server Error", "failed to create transaction", err)
		return
	}

	ctx.JSON(200, models.Response[any]{
		Success: true,
		Message: "Transaction Successs",
	})
}
