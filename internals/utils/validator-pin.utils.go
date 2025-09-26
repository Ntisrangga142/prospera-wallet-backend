package utils

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/pkg"
)

// VerifyUserPIN mengecek apakah PIN yang diberikan sesuai dengan user
func VerifyUserPIN(ctx context.Context, db *pgxpool.Pool, userID int, pin string) (bool, error) {
	var storedPIN string
	err := db.QueryRow(ctx, "SELECT pin FROM accounts WHERE id=$1", userID).Scan(&storedPIN)
	if err != nil {
		return false, err
	}
	hashconfig := pkg.NewHashConfig()
	verify, err := hashconfig.ComparePasswordAndHash(pin, storedPIN)
	if err != nil {
		return false, err
	}
	return verify, nil
}
