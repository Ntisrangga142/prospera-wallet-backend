package utils

import (
	"errors"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prospera/internals/pkg"
)

func GetUserIDFromJWT(ctx *gin.Context) (int, error) {
	// Ambil header Authorization
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return 0, errors.New("missing token")
	}

	// Buang prefix "Bearer "
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	// Siapkan struct Claims
	claims := &pkg.Claims{}

	// Verify token
	if err := claims.VerifyToken(tokenStr); err != nil {
		return 0, err
	}

	// Ambil UserId
	return claims.UserId, nil
}

func GetExpiredFromJWT(ctx *gin.Context) (time.Time, error) {
	// Ambil header Authorization
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return time.Time{}, errors.New("missing token")
	}

	// Buang prefix "Bearer "
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	// Siapkan struct Claims
	claims := &pkg.Claims{}

	// Verify token
	if err := claims.VerifyToken(tokenStr); err != nil {
		return time.Time{}, err
	}

	// Ambil waktu expired
	if claims.RegisteredClaims.ExpiresAt == nil {
		return time.Time{}, errors.New("token has no expiry")
	}

	return claims.RegisteredClaims.ExpiresAt.Time, nil
}

func GetToken(ctx *gin.Context) (string, error) {
	// Ambil header Authorization
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("missing token")
	}

	// Buang prefix "Bearer "
	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

	return tokenStr, nil
}
