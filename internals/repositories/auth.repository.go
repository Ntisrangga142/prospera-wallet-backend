package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/models"
	"github.com/redis/go-redis/v9"
)

type Auth struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewAuthRepo(Db *pgxpool.Pool, Rdb *redis.Client) *Auth {
	return &Auth{db: Db, rdb: Rdb}
}

func (r *Auth) Register(ctx context.Context, email, password string) ([]int, error) {
	// mulai transaction
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	// insert ke account
	var userID int
	queryAccount := `INSERT INTO accounts (email, password) VALUES ($1, $2) RETURNING id`
	if err = tx.QueryRow(ctx, queryAccount, email, password).Scan(&userID); err != nil {
		return nil, fmt.Errorf("failed to insert accounts = %w", err)
	}

	// insert ke profiles
	queryUser := `INSERT INTO profiles (id, fullname, phone, img) VALUES ($1, NULL, NULL, NULL);`
	if _, err = tx.Exec(ctx, queryUser, userID); err != nil {
		return nil, fmt.Errorf("failed to insert profiles = %w", err)
	}

	// insert ke wallets
	var walletID int
	queryWallet := `INSERT INTO wallets (id, balance) VALUES ($1, 0) RETURNING id;`
	if err = tx.QueryRow(ctx, queryWallet, userID).Scan(&walletID); err != nil {
		return nil, fmt.Errorf("failed to insert wallets = %w", err)
	}

	// insert ke participants
	queryParticipant := `INSERT INTO participants (type, ref_id, created_at) VALUES ('wallet', $1, NOW());`
	if _, err = tx.Exec(ctx, queryParticipant, walletID); err != nil {
		return nil, fmt.Errorf("failed to insert participants = %w", err)
	}

	// ambil semua user ID lain
	queryAllUser := `SELECT id FROM accounts WHERE id != $1;`
	rows, err := tx.Query(ctx, queryAllUser, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query accounts = %w", err)
	}
	defer rows.Close()

	var allUserIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan account id = %w", err)
		}
		allUserIDs = append(allUserIDs, id)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error = %w", err)
	}

	// commit transaction
	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return allUserIDs, nil
}

func (r *Auth) Login(ctx context.Context, email string) (int, string, bool, error) {
	query := `SELECT id, password, pin FROM accounts WHERE email = $1`
	var id *int
	var hashedPassword *string
	var pin *string
	err := r.db.QueryRow(ctx, query, email).Scan(&id, &hashedPassword, &pin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, "", false, fmt.Errorf("user not found")
		}
		return 0, "", false, err
	}

	isPinExist := pin != nil && *pin != ""
	return *id, *hashedPassword, isPinExist, nil
}

func (r *Auth) Logout(ctx context.Context, token string, expiresIn time.Duration) error {
	data := models.BlacklistToken{
		Token:     token,
		ExpiresIn: expiresIn,
	}
	bt, err := json.Marshal(data)
	if err != nil {
		log.Println("❌ Internal Server Error.\nCause:", err)
		return err
	}

	if err := r.rdb.Set(ctx, "blacklist:"+token, bt, expiresIn).Err(); err != nil {
		log.Printf("❌ Redis Error.\nCause: %s\n", err)
		return err
	}

	return nil
}

func (ur *Auth) UpdatePIN(rctx context.Context, newPin string, uid int) error {
	sql := `
		UPDATE accounts
		SET pin = $1,
		    updated_at = NOW()
		WHERE id = $2
	`

	ctag, err := ur.db.Exec(rctx, sql, newPin, uid)
	if err != nil {
		return err
	}
	if ctag.RowsAffected() == 0 {
		return errors.New("unable to update user's PIN")
	}

	return nil
}

func (ur *Auth) VerifyUserPIN(ctx context.Context, userID int) (string, error) {
	var storedPIN string
	err := ur.db.QueryRow(ctx, "SELECT pin FROM accounts WHERE id=$1", userID).Scan(&storedPIN)
	if err != nil {
		return "", err
	}
	return storedPIN, nil

}

// func (ur *Auth) CheckEmail(ctx context.Context, emailInput string) (bool, error) {
// 	var dummy string

// 	query := `
// 	select
// 		email
// 	from
// 		accounts
// 	where
// 		email = $1`
// 	if err := ur.db.QueryRow(ctx, query, emailInput).Scan(&dummy); err != nil {
// 		if errors.Is(err, pgx.ErrNoRows) {
// 			return false, nil // email not found
// 		}
// 		return false, err
// 	}

// 	return true, nil
// }

// Forgot Password and PIN
func (r *Auth) FindByEmail(email string) (*models.ForgotPasswordScan, error) {
	var user models.ForgotPasswordScan
	query := `
        SELECT id, email, password, created_at, updated_at
        FROM accounts
        WHERE email = $1
        LIMIT 1
    `
	err := r.db.QueryRow(context.Background(), query, email).
		Scan(&user.ID, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, errors.New("user not found")
	}
	return &user, nil
}

func (r *Auth) SaveResetToken(userID int, token string, expiredAt time.Time) error {
	query := `
        UPDATE accounts SET
		token = $2,
		expired_at = $3
		WHERE id = $1
    `
	_, err := r.db.Exec(context.Background(), query, userID, token, expiredAt)
	return err
}

// Reset PIN and Password
func (ur *Auth) ResetPIN(rctx context.Context, newPin string, token string) error {
	sql := `
		UPDATE accounts
		SET pin = $1,
			token = NULL,
		    expired_at = NULL,
		    updated_at = NOW()
		WHERE token = $2 and expired_at > NOW()
	`
	ctag, err := ur.db.Exec(rctx, sql, newPin, token)
	if err != nil {
		return err
	}
	if ctag.RowsAffected() == 0 {
		return errors.New("unable to update user's PIN")
	}

	return nil
}

func (ur *Auth) ResetPassword(rctx context.Context, newPassword string, token string) error {
	sql := `
		UPDATE accounts
		SET password = $1,
			token = NULL,
		    expired_at = NULL,
		    updated_at = NOW()
		WHERE token = $2 AND expired_at > NOW()
	`

	ctag, err := ur.db.Exec(rctx, sql, newPassword, token)
	if err != nil {
		return err
	}
	if ctag.RowsAffected() == 0 {
		return errors.New("unable to update user's Password")
	}

	return nil
}
