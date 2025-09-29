package handlers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/models"
	"github.com/prospera/internals/pkg"
	"github.com/prospera/internals/repositories"
	"github.com/prospera/internals/utils"
	"github.com/redis/go-redis/v9"
)

type AuthHandler struct {
	Repo *repositories.Auth
	rdb  *redis.Client
}

func NewAuthHandler(repo *repositories.Auth, rdb *redis.Client) *AuthHandler {
	return &AuthHandler{Repo: repo, rdb: rdb}
}

func (h *AuthHandler) Register(ctx *gin.Context) {
	var req models.RegisterRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed binding data", err)
		return
	}

	// Hash Password
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()
	hashedPassword, err := hashConfig.GenHash(req.Password)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed hashed password", err)
		return
	}

	listId, err := h.Repo.Register(ctx, req.Email, hashedPassword)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "Email is already registered", err)
		return
	}

	for _, v := range listId {
		var redisKey = fmt.Sprintf("Prospera-AllUser-%d", v)
		if err := utils.InvalidateCache(ctx, h.rdb, redisKey); err != nil {
			log.Println("Failed invalidate cache:", err)
		}
	}

	ctx.JSON(http.StatusCreated, models.Response[any]{
		Success: true,
		Message: "Register account successful",
	})
}

func (h *AuthHandler) Login(ctx *gin.Context) {
	var req models.LoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed binding data", err)
		return
	}

	// Cari akun
	userID, hashedPassword, isPinExist, err := h.Repo.Login(ctx.Request.Context(), req.Email)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "user not found", err)
		return
	}
	if userID == 0 {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "user not found", err)
		return
	}

	// Verifikasi password
	hashConfig := pkg.NewHashConfig()
	match, err := hashConfig.ComparePasswordAndHash(req.Password, hashedPassword)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed compare password", err)
		return
	}
	if !match {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "invalid username or password", errors.New("invalid password"))
		return
	}

	// Generate JWT
	claims := pkg.NewJWTClaims(userID, req.Email)
	token, err := claims.GenToken()
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed generate token", err)
		return
	}

	ctx.JSON(http.StatusOK, models.ResponseLogin{
		Success:    true,
		Message:    "Login successful",
		Token:      token,
		IsPinExist: isPinExist,
		Email:      claims.Email,
	})
}

func (h *AuthHandler) Logout(ctx *gin.Context) {
	token, err := utils.GetToken(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "failed get token", err)
		return
	}

	expiresAt, err := utils.GetExpiredFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "failed get expired time token", err)
		return
	}

	expiresIn := time.Until(expiresAt)
	if expiresIn <= 0 {
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "token already expired", err)
		return
	}

	if err = h.Repo.Logout(context.Background(), token, expiresIn); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to blacklist token", err)
		return
	}

	ctx.JSON(http.StatusOK, models.Response[any]{
		Success: true,
		Message: "Successfully logged out",
	})
}

func (h *AuthHandler) VerifyPIN(c *gin.Context) {
	var req models.PINRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.HandleError(c, http.StatusBadRequest, "Bad Request", "invalid request", err)
		return
	}

	// Ambil ID dari token
	id, err := utils.GetUserIDFromJWT(c)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	// Ambil pin yang tersimpan dari repo
	storedPIN, err := h.Repo.VerifyUserPIN(c.Request.Context(), id)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Internal Server Error", "failed to fetch account", err)
		return
	}

	// Compare di handler
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()
	valid, err := hashConfig.ComparePasswordAndHash(req.PIN, storedPIN)
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Internal Server Error", "failed to compare pin", err)
		return
	}

	if !valid {
		c.JSON(http.StatusBadRequest, models.Response[bool]{
			Success: false,
			Message: "PIN does not match",
			Data:    valid,
		})
		return
	}

	c.JSON(http.StatusOK, models.Response[bool]{
		Success: true,
		Message: "Success Verify PIN",
		Data:    valid,
	})
}

func (h *AuthHandler) UpdatePIN(ctx *gin.Context) {
	var req models.PINRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed binding data", err)
		return
	}

	//Get ID from Token
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	// Hash Password
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()
	hashedPIN, err := hashConfig.GenHash(req.PIN)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed hashed password", err)
		return
	}

	if err := h.Repo.UpdatePIN(ctx, hashedPIN, uid); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed created account", err)
		return
	}

	ctx.JSON(http.StatusCreated, models.Response[string]{
		Success: true,
		Message: "Register PIN successful",
		Data:    "",
	})
}

// Check old pin, if matches then update to new pin
func (h *AuthHandler) ChangePIN(ctx *gin.Context) {
	var req models.ChangePINRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed binding data", err)
		return
	}

	// Get ID from Token
	uid, err := utils.GetUserIDFromJWT(ctx)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "unable to get user's token", err)
		return
	}

	// Verify Old Pin
	// Ambil pin yang tersimpan dari repo
	storedPIN, err := h.Repo.VerifyUserPIN(ctx.Request.Context(), uid)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to fetch account", err)
		return
	}

	// Compare di handler
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()
	valid, err := hashConfig.ComparePasswordAndHash(req.OldPIN, storedPIN)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to compare pin", err)
		return
	}

	if !valid {
		ctx.JSON(http.StatusBadRequest, models.Response[bool]{
			Success: false,
			Message: "PIN does not match",
			Data:    valid,
		})
		return
	}

	// Hash Password
	hashConfig.UseRecommended()
	hashedPIN, err := hashConfig.GenHash(req.NewPIN)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed hashed password", err)
		return
	}

	if err := h.Repo.UpdatePIN(ctx, hashedPIN, uid); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed created account", err)
		return
	}

	ctx.JSON(http.StatusCreated, models.Response[string]{
		Success: true,
		Message: "Change PIN successful",
		Data:    "",
	})
}

// func (h *AuthHandler) CheckEmail(ctx *gin.Context) {
// 	email := ctx.Query("email")
// 	if email == "" {
// 		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "missing email query param", nil)
// 		return
// 	}

// 	exists, err := h.Repo.CheckEmail(ctx, email)
// 	if err != nil {
// 		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed to check email", err)
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, models.Response[map[string]bool]{
// 		Success: true,
// 		Message: "Email check successful",
// 		Data: map[string]bool{
// 			"exists": exists,
// 		},
// 	})
// }

func (h *AuthHandler) Forgot(ctx *gin.Context) {
	var req models.ForgotRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, 400, "Bad Request", "Invalid request", err)
		return
	}

	if req.Type != "password" && req.Type != "pin" {
		utils.HandleError(ctx, 400, "Bad Request", "Invalid request", fmt.Errorf("type must be password or pin"))
		return
	}

	// Cek user
	user, err := h.Repo.FindByEmail(req.Email)
	if err != nil {
		utils.HandleError(ctx, 404, "Not Found", "email not found", err)
		return
	}

	// Generate token
	token := utils.GenerateRandomToken()
	h.Repo.SaveResetToken(user.ID, token, time.Now().Add(15*time.Minute))

	// Kirim email
	var resetURL string
	if req.Type == "password" {
		resetURL = fmt.Sprintf("%sauth/create/password?token=%s", os.Getenv("URL_BASE"), token)
	} else {
		resetURL = fmt.Sprintf("%sauth/create/pin?token=%s", os.Getenv("URL_BASE"), token)
	}

	utils.SendResetPasswordEmail(user.Email, resetURL, req.Type)

	ctx.JSON(200, models.Response[any]{
		Success: true,
		Message: "Reset link sent to your email",
	})
}

func (h *AuthHandler) ResetPIN(ctx *gin.Context) {
	var req models.PINResetRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed binding data", err)
		return
	}
	// Hash Password
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()
	hashedPIN, err := hashConfig.GenHash(req.PIN)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed hashed pin", err)
		return
	}

	if err := h.Repo.ResetPIN(ctx, hashedPIN, req.Token); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed reset pin", err)
		return
	}

	ctx.JSON(http.StatusCreated, models.Response[any]{
		Success: true,
		Message: "Reset PIN successful",
	})
}

func (h *AuthHandler) ResetPassword(ctx *gin.Context) {
	var req models.PasswordResetRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		utils.HandleError(ctx, http.StatusBadRequest, "Bad Request", "failed binding data", err)
		return
	}
	// Hash Password
	hashConfig := pkg.NewHashConfig()
	hashConfig.UseRecommended()
	hashedPassword, err := hashConfig.GenHash(req.Password)
	if err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed hashed password", err)
		return
	}

	if err := h.Repo.ResetPassword(ctx, hashedPassword, req.Token); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed reset password", err)
		return
	}

	ctx.JSON(http.StatusCreated, models.Response[any]{
		Success: true,
		Message: "Reset Password successful",
	})
}
