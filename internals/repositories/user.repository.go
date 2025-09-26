package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// GET PROFILE
func (ur *UserRepository) GetProfile(ctx context.Context, uid int) (*models.Profile, error) {
	sql := `
		SELECT fullname, phone, img, verified
		FROM profiles
		WHERE id = $1
	`

	row := ur.db.QueryRow(ctx, sql, uid)

	var profile models.Profile
	err := row.Scan(
		&profile.FullName,
		&profile.PhoneNumber,
		&profile.Avatar,
		&profile.Verified,
	)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// UPDATE PROFILE
func (ur *UserRepository) UpdateProfile(ctx context.Context, uid int, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}

	setClauses := []string{}
	args := []any{}
	i := 1

	for col, val := range updates {
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", col, i))
		args = append(args, val)
		i++
	}

	// tambah updated_at
	setClauses = append(setClauses, "updated_at = NOW()")

	query := fmt.Sprintf(`
		UPDATE profiles
		SET %s
		WHERE id = $%d
	`, strings.Join(setClauses, ", "), i)

	args = append(args, uid)

	_, err := ur.db.Exec(ctx, query, args...)
	return err
}

// GET ALL USERS
func (ur *UserRepository) GetAllUser(rctx context.Context, uid int) ([]models.User, error) {
	sql := `
		SELECT id, fullname, phone, img
		FROM profiles
		WHERE id != $1
	`

	rows, err := ur.db.Query(rctx, sql, uid)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(
			&user.ID,
			&user.FullName,
			&user.PhoneNumber,
			&user.Avatar,
		); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}

// GET HISTORY TRANSACTIONS
func (ur *UserRepository) GetUserHistoryTransactions(ctx context.Context, userID int) ([]models.TransactionHistory, error) {
	query := `
	WITH user_participant AS (
		SELECT p.id AS participant_id
		FROM accounts a
		JOIN wallets w ON w.id = a.id
		JOIN participants p ON p.ref_id = w.id AND p.type = 'wallet'
		WHERE a.id = $1
	)
	SELECT 
		t.id,
		t.type,
		t.total,
		CASE 
			WHEN t.id_sender = up.participant_id THEN 'debit'
			WHEN t.id_receiver = up.participant_id THEN 'credit'
		END AS direction,
		CASE 
			WHEN t.id_sender = up.participant_id THEN pr.type
			ELSE ps.type
		END AS counterparty_type,
		COALESCE(
			(CASE 
				WHEN t.id_sender = up.participant_id AND pr.type = 'wallet' THEN prf.fullname
				WHEN t.id_receiver = up.participant_id AND ps.type = 'wallet' THEN prf.fullname
			END),
			(CASE 
				WHEN t.id_sender = up.participant_id AND pr.type = 'internal' THEN ia.name
				WHEN t.id_receiver = up.participant_id AND ps.type = 'internal' THEN ia.name
			END)
		) AS counterparty_name,
		COALESCE(
			(CASE 
				WHEN t.id_sender = up.participant_id AND pr.type = 'wallet' THEN prf.img
				WHEN t.id_receiver = up.participant_id AND ps.type = 'wallet' THEN prf.img
			END),
			(CASE 
				WHEN t.id_sender = up.participant_id AND pr.type = 'internal' THEN ia.img
				WHEN t.id_receiver = up.participant_id AND ps.type = 'internal' THEN ia.img
			END)
		) AS counterparty_img,
		CASE 
			WHEN (t.id_sender = up.participant_id AND pr.type = 'wallet') THEN prf.phone
			WHEN (t.id_receiver = up.participant_id AND ps.type = 'wallet') THEN prf.phone
			ELSE NULL
		END AS counterparty_phone,
		t.created_at
	FROM transactions t
	JOIN user_participant up 
		ON t.id_sender = up.participant_id OR t.id_receiver = up.participant_id
	JOIN participants ps ON ps.id = t.id_sender
	JOIN participants pr ON pr.id = t.id_receiver
	LEFT JOIN wallets w 
		ON ( (ps.type = 'wallet' AND ps.ref_id = w.id) OR (pr.type = 'wallet' AND pr.ref_id = w.id) )
	LEFT JOIN profiles prf ON prf.id = w.id
	LEFT JOIN internal_accounts ia 
		ON ( (ps.type = 'internal' AND ps.ref_id = ia.id) OR (pr.type = 'internal' AND pr.ref_id = ia.id) )
	ORDER BY DATE(t.created_at) DESC, t.created_at DESC;
	`

	rows, err := ur.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.TransactionHistory
	for rows.Next() {
		var tx models.TransactionHistory
		err := rows.Scan(
			&tx.ID,
			&tx.Type,
			&tx.Total,
			&tx.Direction,
			&tx.CounterpartyType,
			&tx.CounterpartyName,
			&tx.CounterpartyImg,
			&tx.CounterpartyPhone,
			&tx.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, nil
}

// DELETE HISTORY TRANSACTIONS
func (ur *UserRepository) SoftDeleteTransaction(rctx context.Context, uid, transactionId int) error {
	sql := `
		UPDATE transactions SET 
		deleted_sender =
		CASE
		WHEN id_sender = $1 THEN CURRENT_DATE
		ELSE deleted_sender
		END, 
		deleted_receiver =
		CASE
		WHEN id_receiver = $1 THEN CURRENT_DATE
		ELSE deleted_receiver
		END
		WHERE id = $2
	`

	ctag, err := ur.db.Exec(rctx, sql, uid, transactionId)
	if err != nil {
		return err
	}

	if ctag.RowsAffected() == 0 {
		return errors.New("no matching transaction id")
	}

	return nil
}

// PATCH CHANGE PASSWORD
func (ur *UserRepository) GetPasswordFromID(ctx context.Context, id int) (string, error) {
	query := `
		SELECT
			password
		FROM
			accounts
		WHERE
			id = $1`

	var userPass string
	if err := ur.db.QueryRow(ctx, query, id).Scan(&userPass); err != nil {
		return "", errors.New("failed to get password")
	}
	return userPass, nil
}

func (ur *UserRepository) ChangePassword(ctx context.Context, userID int, hashedPassword string) error {
	query := `
		UPDATE
			accounts
		SET
			password = $1,
			updated_at = CURRENT_TIMESTAMP
		WHERE
			id = $2`
	_, err := ur.db.Exec(ctx, query, hashedPassword, userID)
	if err != nil {
		return err
	}
	return nil
}

func (ur *UserRepository) GetUserInfo(ctx context.Context, uid int) (models.User, error) {
	sql := `
		SELECT
			id, fullname, phone, img, verified
		FROM
			profiles
		WHERE
			id = $1
	`

	var userInfo models.User
	if err := ur.db.QueryRow(ctx, sql, uid).Scan(
		&userInfo.ID,
		&userInfo.FullName,
		&userInfo.PhoneNumber,
		&userInfo.Avatar,
		&userInfo.IsVerified,
	); err != nil {
		return models.User{}, err
	}

	return userInfo, nil
}
