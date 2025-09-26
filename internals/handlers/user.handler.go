package handlers

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/models"
	"github.com/prospera/internals/pkg"
	"github.com/prospera/internals/repositories"
	"github.com/prospera/internals/utils"
	"github.com/redis/go-redis/v9"
)

type UserHandler struct {
	ur  *repositories.UserRepository
	rdb *redis.Client
}

func NewUserHandler(ur *repositories.UserRepository, rdb *redis.Client) *UserHandler {
	return &UserHandler{ur: ur, rdb: rdb}
}

// GET PROFILE
func (uh *UserHandler) GetProfile(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	var cachedData models.Profile
	var redisKey = fmt.Sprintf("Prospera-Profile-%d", uid)
	if err := utils.CacheHit(ctx.Request.Context(), uh.rdb, redisKey, &cachedData); err == nil {
		ctx.JSON(http.StatusOK, models.Response[models.Profile]{
			Success: true,
			Message: "Success Get Profile User (from cache)",
			Data:    cachedData,
		})
		return
	}

	profile, err := uh.ur.GetProfile(ctx.Request.Context(), uid)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable get profile user", err)
		return
	}

	if err := utils.RenewCache(ctx.Request.Context(), uh.rdb, redisKey, profile, 10); err != nil {
		log.Println("Failed to set redis cache:", err)
	}

	ctx.JSON(http.StatusOK, models.Response[models.Profile]{
		Success: true,
		Message: "Success Get Profile User",
		Data:    *profile,
	})
}

// UPDATE PROFILE
func (uh *UserHandler) UpdateProfile(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "invalid token", err)
		return
	}

	updates := make(map[string]interface{})

	// ambil field dari form-data (jika ada)
	if fullname := ctx.PostForm("fullname"); fullname != "" {
		updates["fullname"] = fullname
	}
	if phone := ctx.PostForm("phone"); phone != "" {
		updates["phone"] = phone
	}

	// upload image jika ada
	file, err := ctx.FormFile("img")
	if err == nil {
		destDir := "public/profile"
		filename := fmt.Sprintf("profile_%d", uid)

		path, saveErr := utils.SaveUploadedFile(ctx, file, destDir, filename)
		if saveErr != nil {
			utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "Upload Failed", saveErr)
			return
		}

		updates["img"] = path
	}

	// update ke DB
	if err := uh.ur.UpdateProfile(ctx.Request.Context(), uid, updates); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to update profile", err)
		return
	}

	var redisKey = fmt.Sprintf("Prospera-Profile-%d", uid)
	if err := utils.InvalidateCache(ctx, uh.rdb, redisKey); err != nil {
		log.Println("Failed invalidate cache:", err)
	}

	ctx.JSON(http.StatusOK, models.Response[any]{
		Success: true,
		Message: "Profile updated successfully",
	})
}

// GET ALL USERS
func (uh *UserHandler) GetAllUsers(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	var cachedData []models.User
	var redisKey = fmt.Sprintf("Prospera-AllUser-%d", uid)
	if err := utils.CacheHit(ctx.Request.Context(), uh.rdb, redisKey, &cachedData); err == nil {
		ctx.JSON(http.StatusOK, models.Response[[]models.User]{
			Success: true,
			Message: "Success Get History (from cache)",
			Data:    cachedData,
		})
		return
	}

	users, err := uh.ur.GetAllUser(ctx.Request.Context(), uid)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable get users", err)
		return
	}

	if err := utils.RenewCache(ctx.Request.Context(), uh.rdb, redisKey, users, 10); err != nil {
		log.Println("Failed to set redis cache:", err)
	}

	ctx.JSON(http.StatusOK, models.Response[[]models.User]{
		Success: true,
		Message: "User's list",
		Data:    users,
	})
}

// GET HISTORY TRANSACTIONS
func (h *UserHandler) GetUserHistoryTransactions(c *gin.Context) {
	// Ambil user_id dari Token
	userID, err := utils.GetUserIDFromJWT(c)
	if err != nil {
		utils.HandleError(c, http.StatusUnauthorized, "Unauthorized", "user not found", err)
		return
	}

	var cachedData []models.TransactionHistory
	var redisKey = fmt.Sprintf("Prospera-HistoryTransactions_%d", userID)
	if err := utils.CacheHit(c.Request.Context(), h.rdb, redisKey, &cachedData); err == nil {
		c.JSON(http.StatusOK, models.Response[[]models.TransactionHistory]{
			Success: true,
			Message: "Success Get History (from cache)",
			Data:    cachedData,
		})
		return
	}

	transactions, err := h.ur.GetUserHistoryTransactions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		utils.HandleError(c, http.StatusInternalServerError, "Internal Server Error", "failed get history transaction", err)
		return
	}

	if err := utils.RenewCache(c.Request.Context(), h.rdb, redisKey, transactions, 10); err != nil {
		log.Println("Failed to set redis cache:", err)
	}

	c.JSON(http.StatusOK, models.Response[[]models.TransactionHistory]{
		Success: true,
		Message: "Success Get History",
		Data:    transactions,
	})
}

// DELETE HISTORY TRANSACTIONS
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

	var redisKey = fmt.Sprintf("Prospera-HistoryTransactions_%d", uid)
	if err := utils.InvalidateCache(ctx, uh.rdb, redisKey); err != nil {
		log.Println("Failed invalidate cache:", err)
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

func (uh *UserHandler) HandleGetUserInfo(ctx *gin.Context) {
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's id", err)
		return
	}

	userInfo, err := uh.ur.GetUserInfo(ctx, uid)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's information", err)
		return
	}

	ctx.JSON(http.StatusOK, models.Response[models.User]{
		Success: true,
		Message: "success",
		Data:    userInfo,
	})
}
