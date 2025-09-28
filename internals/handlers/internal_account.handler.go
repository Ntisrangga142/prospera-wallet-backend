package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/models"
	"github.com/prospera/internals/repositories"
	"github.com/prospera/internals/utils"
	"github.com/redis/go-redis/v9"
)

type InternalAccountHandler struct {
	repo *repositories.InternalAccountRepository
	rdb  *redis.Client
}

func NewInternalAccountHandler(repo *repositories.InternalAccountRepository, rdb *redis.Client) *InternalAccountHandler {
	return &InternalAccountHandler{repo: repo, rdb: rdb}
}

func (h *InternalAccountHandler) GetAll(c *gin.Context) {
	var cachedData []models.InternalAccount
	var redisKey = "Prospera-InternalAccount"
	if err := utils.CacheHit(c.Request.Context(), h.rdb, redisKey, &cachedData); err == nil {
		c.JSON(http.StatusOK, models.Response[any]{
			Success: true,
			Message: "Success Get Internal Account (from cache)",
			Data:    cachedData,
		})
		return
	}

	accounts, err := h.repo.GetAll(c.Request.Context())
	if err != nil {
		utils.HandleError(c, http.StatusInternalServerError, "Internal Server Error", "Failed to get internal accounts", err)
		return
	}

	if err := utils.RenewCache(c.Request.Context(), h.rdb, redisKey, accounts, 10); err != nil {
		log.Println("Failed to set redis cache:", err)
	}

	c.JSON(http.StatusOK, models.Response[any]{
		Success: true,
		Message: "Success get internal accounts",
		Data:    accounts,
	})
}
