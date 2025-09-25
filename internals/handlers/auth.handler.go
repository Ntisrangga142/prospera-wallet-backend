package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/models"
	"github.com/prospera/internals/pkg"
	"github.com/prospera/internals/repositories"
	"github.com/prospera/internals/utils"
)

type AuthHandler struct {
	Repo *repositories.Auth
}

func NewAuthHandler(repo *repositories.Auth) *AuthHandler {
	return &AuthHandler{Repo: repo}
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

	if err := h.Repo.Register(ctx, req.Email, hashedPassword); err != nil {
		utils.HandleError(ctx, http.StatusInternalServerError, "Internal Server Error", "failed created account", err)
		return
	}

	ctx.JSON(http.StatusCreated, models.Response[string]{
		Success: true,
		Message: "Register account successful",
		Data:    "",
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
		utils.HandleError(ctx, http.StatusUnauthorized, "Unauthorized", "invalid password", err)
		return
	}

	// Generate JWT
	claims := pkg.NewJWTClaims(userID)
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
	})
}
