package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/models"
)

type TransactionRepository struct {
	db *pgxpool.Pool
}

func NewTransactionRepository(db *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{db: db}
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

func (r *TransactionRepository) CreateTransaction(ctx context.Context, txReq *models.TransactionRequest, userID int) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // rollback jika error

	var senderID, receiverID int

	switch txReq.Type {
	case "top_up":
		// Sender: internal account
		if txReq.InternalAccountID == nil {
			return errors.New("internal account id required for top_up")
		}
		senderID, err = r.getOrCreateParticipant(ctx, tx, "internal", *txReq.InternalAccountID)
		if err != nil {
			return err
		}

		// Receiver: wallet (userID)
		receiverID, err = r.getOrCreateParticipant(ctx, tx, "wallet", userID)
		if err != nil {
			return err
		}

	case "transfer":
		// Sender: wallet (userID)
		senderID, err = r.getOrCreateParticipant(ctx, tx, "wallet", userID)
		if err != nil {
			return err
		}

		// Receiver: wallet (lain)
		if txReq.ReceiverAccountID == nil {
			return errors.New("receiver account id required for transfer")
		}
		receiverID, err = r.getOrCreateParticipant(ctx, tx, "wallet", *txReq.ReceiverAccountID)
		if err != nil {
			return err
		}

	default:
		return fmt.Errorf("invalid transaction type: %s", txReq.Type)
	}

	// Insert transaksi
	var transaction models.Transaction
	query := `
		INSERT INTO transactions (type, amount, total, note, id_sender, id_receiver)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, type, amount, total, note, id_sender, id_receiver, created_at
	`
	err = tx.QueryRow(ctx, query,
		txReq.Type,
		txReq.Amount,
		txReq.Total,
		txReq.Note,
		senderID,
		receiverID,
	).Scan(
		&transaction.ID,
		&transaction.Type,
		&transaction.Amount,
		&transaction.Total,
		&transaction.Note,
		&transaction.SenderID,
		&transaction.ReceiverID,
		&transaction.CreatedAt,
	)
	if err != nil {
		return err
	}

	// Commit transaksi
	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
