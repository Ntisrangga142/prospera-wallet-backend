package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (ur *UserRepository) GetUser(rctx context.Context, uid int) ([]models.User, error) {
	sql := `
		SELECT fullname, phone, img
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

func (ur *UserRepository) GetUserHistoryTransactions(rctx context.Context, uid, limit, offset int) (models.UserHistoryTransactions, error) {
	sql := `
		SELECT
			t.id,
			t.id_receiver,
			p.img,
			p.fullname,
			p.phone,
			t.type,
			t.total
		FROM transactions t
		LEFT JOIN profiles p ON p.id = t.id_receiver
		WHERE t.deleted_sender IS NULL
		AND t.id_sender = $1
		ORDER BY t.created_at DESC;
	`

	rows, err := ur.db.Query(rctx, sql, uid)
	if err != nil {
		return models.UserHistoryTransactions{}, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var transaction models.Transaction
		if err := rows.Scan(
			&transaction.TransactionID,
			&transaction.ReceiverID,
			&transaction.Avatar,
			&transaction.FullName,
			&transaction.PhoneNumber,
			&transaction.TransactionType,
			&transaction.Total,
		); err != nil {
			return models.UserHistoryTransactions{}, err
		}

		transactions = append(transactions, transaction)
	}

	return models.UserHistoryTransactions{
		ID: uid, Transactions: transactions,
	}, nil
}

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

// Untuk mengganti password
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
