package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/models"
	"github.com/prospera/internals/pkg"
	"github.com/prospera/internals/repositories"
	"github.com/prospera/internals/utils"
	"github.com/redis/go-redis/v9"
)

type TransactionHandler struct {
	repo     *repositories.TransactionRepository
	repoAuth *repositories.Auth
	rdb      *redis.Client
}

func NewTransactionHandler(repo *repositories.TransactionRepository, rdb *redis.Client, repoAuth *repositories.Auth) *TransactionHandler {
	return &TransactionHandler{repo: repo, rdb: rdb, repoAuth: repoAuth}
}

// Create Transactions
func (h *TransactionHandler) CreateTransaction(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "invalid token", err)
		return
	}

	var req models.TransactionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "invalid payload", err)
		return
	}

	// Ambil PIN user dari DB
	storedPIN, err := h.repoAuth.VerifyUserPIN(ctx.Request.Context(), uid)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to fetch user pin", err)
		return
	}

	// Bandingkan PIN yang dikirim dengan hash
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()

	valid, err := hashConfig.ComparePasswordAndHash(req.PIN, storedPIN)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to verify pin", err)
		return
	}
	if !valid {
		utils.HandleError(ctx, http.StatusForbidden, "Forbidden", "invalid PIN", nil)
		return
	}

	// Buat transaksi
	if err := h.repo.CreateTransaction(ctx.Request.Context(), &req, uid); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to create transaction", err)
		return
	}

	if req.Type == "transfer" {
		message := fmt.Sprintf("Kamu menerima transfer Rp%d.00 dari user %d", req.Amount, uid)
		pkg.WebSocketHub.SendToUser(*req.ReceiverAccountID, message)
	} else {
		message := fmt.Sprintf("Kamu menerima transfer Rp%d.00 dari user %d", req.Amount, uid)
		pkg.WebSocketHub.SendToUser(uid, message)
	}

	ctx.JSON(http.StatusOK, models.Response[any]{
		Success: true,
		Message: "Transaction Success",
	})
}
