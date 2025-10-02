package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/prospera/internals/pkg"
	"github.com/prospera/internals/utils"
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client

func InitRedis(rdb *redis.Client) {
	RDB = rdb
}

func Authentication(ctx *gin.Context) {
	// ambil token dari header
	bearerToken := ctx.GetHeader("Authorization")
	if bearerToken == "" {
		utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "Unauthorized Access", "Authorization header is missing, please login first")
		ctx.Abort()
		return
	}

	// Bearer token
	tokens := strings.Split(bearerToken, " ")
	if len(tokens) != 2 {
		utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "Unauthorized Access", "Invalid authorization format, expected 'Bearer <token>'")
		ctx.Abort()
		return
	}

	token := tokens[1]
	if token == "" {
		utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "Unauthorized Access", "Access token is missing, please login first")
		ctx.Abort()
		return
	}

	// cek apakah token sudah di-blacklist
	isBlacklisted, err := utils.IsBlacklisted(ctx, RDB, token)
	if err != nil {
		utils.HandleMiddlewareError(ctx, http.StatusInternalServerError, "Internal Server Error", err.Error())
		ctx.Abort()
		return
	}
	if isBlacklisted {
		utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "Unauthorized Access", "Token has been revoked, please login again")
		ctx.Abort()
		return
	}

	// verifikasi JWT
	var claims pkg.Claims
	if err := claims.VerifyToken(token); err != nil {
		if strings.Contains(err.Error(), jwt.ErrTokenInvalidIssuer.Error()) {
			utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "Unauthorized Access", jwt.ErrTokenInvalidIssuer.Error())
			ctx.Abort()
			return
		}
		if strings.Contains(err.Error(), jwt.ErrTokenExpired.Error()) {
			utils.HandleMiddlewareError(ctx, http.StatusUnauthorized, "Unauthorized Access", jwt.ErrTokenExpired.Error())
			ctx.Abort()
			return
		}
		utils.HandleMiddlewareError(ctx, http.StatusInternalServerError, "Internal Server Error", err.Error())
		ctx.Abort()
		return
	}

	// simpan claims ke context
	ctx.Set("claims", claims)
	ctx.Next()
}
