package repositories

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/models"
)

type TransactionRepository struct {
	DB *pgxpool.Pool
}

func NewTransactionRepository(DB *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{DB: DB}
}

func (r *TransactionRepository) getOrCreateParticipant(ctx context.Context, tx pgx.Tx, typ string, refID int) (int, error) {
	var id int
	err := tx.QueryRow(ctx, `
		INSERT INTO participants (type, ref_id)
		VALUES ($1, $2)
		ON CONFLICT (type, ref_id) DO UPDATE SET ref_id = EXCLUDED.ref_id
		RETURNING id
	`, typ, refID).Scan(&id)

	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *TransactionRepository) CreateTransaction(ctx context.Context, txReq *models.TransactionRequest, userID int) (err error) {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}

	// Rollback kalau error
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var senderID, receiverID int

	switch txReq.Type {
	case "top_up":
		if txReq.InternalAccountID == nil {
			return errors.New("internal account id required for top_up")
		}
		if txReq.Amount <= 0 {
			return errors.New("amount must be greater than zero")
		}

		// Internal (sender)
		senderID, err = r.getOrCreateParticipant(ctx, tx, "internal", *txReq.InternalAccountID)
		if err != nil {
			return err
		}

		// Wallet user (receiver)
		receiverID, err = r.getOrCreateParticipant(ctx, tx, "wallet", userID)
		if err != nil {
			return err
		}

	case "transfer":
		if txReq.ReceiverAccountID == nil {
			return errors.New("receiver account id required for transfer")
		}
		if txReq.Amount <= 0 {
			return errors.New("amount must be greater than zero")
		}

		// Wallet sender
		senderID, err = r.getOrCreateParticipant(ctx, tx, "wallet", userID)
		if err != nil {
			return err
		}

		// Wallet receiver
		receiverID, err = r.getOrCreateParticipant(ctx, tx, "wallet", *txReq.ReceiverAccountID)
		if err != nil {
			return err
		}

	default:
		return errors.New("invalid transaction type")
	}

	// Catat transaksi
	_, err = tx.Exec(ctx, `
		INSERT INTO transactions (type, amount, total, note, id_sender, id_receiver)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, txReq.Type, txReq.Amount, txReq.Total, txReq.Note, senderID, receiverID)
	if err != nil {
		return err
	}

	// Update saldo sesuai tipe transaksi
	switch txReq.Type {
	case "transfer":
		// Kurangi saldo sender
		_, err = tx.Exec(ctx, `
			UPDATE wallets w
			SET balance = balance - $1, updated_at = NOW()
			FROM participants p
			WHERE p.id = $2 AND p.ref_id = w.id AND p.type = 'wallet'
		`, txReq.Amount, senderID)
		if err != nil {
			return err
		}

		// Tambah saldo receiver
		_, err = tx.Exec(ctx, `
			UPDATE wallets w
			SET balance = balance + $1, updated_at = NOW()
			FROM participants p
			WHERE p.id = $2 AND p.ref_id = w.id AND p.type = 'wallet'
		`, txReq.Amount, receiverID)
		if err != nil {
			return err
		}

	case "top_up":
		// Tambah saldo receiver (wallet user)
		_, err = tx.Exec(ctx, `
			UPDATE wallets w
			SET balance = balance + $1, updated_at = NOW()
			FROM participants p
			WHERE p.id = $2 AND p.ref_id = w.id AND p.type = 'wallet'
		`, txReq.Amount, receiverID)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
