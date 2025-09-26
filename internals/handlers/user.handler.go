package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/models"
	"github.com/prospera/internals/pkg"
	"github.com/prospera/internals/repositories"
	"github.com/prospera/internals/utils"
)

type UserHandler struct {
	ur *repositories.UserRepository
}

func NewUserHandler(ur *repositories.UserRepository) *UserHandler {
	return &UserHandler{ur: ur}
}

func (uh *UserHandler) HandlerGetAllUsers(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	users, err := uh.ur.GetUser(ctx.Request.Context(), uid)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable get users", err)
		return
	}

	ctx.JSON(http.StatusOK, models.Response[[]models.User]{
		Success: true,
		Message: "User's list",
		Data:    users,
	})
}

func (uh *UserHandler) HandleGetUserTransactionsHistory(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable get user's history transaction", err)
		return
	}

	history, err := uh.ur.GetUserHistoryTransactions(ctx, uid, 0, 0)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable get user's history transaction", err)
		return
	}

	ctx.JSON(http.StatusOK, models.Response[models.UserHistoryTransactions]{
		Success: true,
		Message: "success",
		Data:    history,
	})
}

func (uh *UserHandler) HandleSoftDeleteTransaction(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	transId, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "invalid param input", err)
		return
	}

	if err := uh.ur.SoftDeleteTransaction(ctx.Request.Context(), uid, transId); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to delete history", err)
	}

	ctx.JSON(http.StatusOK, models.Response[string]{
		Success: true,
		Message: "success",
		Data:    "history deleted",
	})
}

// Digunakan di halaman Change Password
func (uh *UserHandler) ChangePassword(ctx *gin.Context) {
	var req models.ChangePassword

	if err := ctx.ShouldBind(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed to bind", err)
		return
	}

	// Get ID from token
	userID, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "user not found", err)
		return
	}

	// Fetch current password hash from DB
	oldPassDB, err := uh.ur.GetPasswordFromID(ctx, userID)
	if err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "bad request", "failed to fetch user credentials", err)
		return
	}

	// Compare old password
	hashCfg := pkg.NewHashConfig()
	isMatched, err := hashCfg.ComparePasswordAndHash(req.OldPassword, oldPassDB)
	if err != nil || !isMatched {
		var handleErr error = err
		if handleErr == nil && !isMatched {
			handleErr = errors.New("old password does not match")
		}
		utils.HandleError(ctx, http.StatusBadRequest, "bad request", "old password does not match", handleErr)
		return
	}

	// Hash new password
	hashCfg.UseRecommended()
	hashedPassword, err := hashCfg.GenHash(req.NewPassword)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", "failed to hash new password", err)
		return
	}

	// Update password
	if err := uh.ur.ChangePassword(ctx, userID, hashedPassword); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "internal server error", "failed to change password", err)
		return
	}

	ctx.JSON(http.StatusCreated, models.Response[any]{
		Success: true,
		Message: "Change password successful",
	})
}
