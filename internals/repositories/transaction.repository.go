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

func (r *TransactionRepository) CreateTransaction(ctx context.Context, txReq *models.TransactionRequest, userID int) error {
	tx, err := r.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var senderID, receiverID int

	switch txReq.Type {
	case "top_up":
		if txReq.InternalAccountID == nil {
			return errors.New("internal account id required for top_up")
		}
		senderID, err = r.getOrCreateParticipant(ctx, tx, "internal", *txReq.InternalAccountID)
		if err != nil {
			return err
		}
		receiverID, err = r.getOrCreateParticipant(ctx, tx, "wallet", userID)
		if err != nil {
			return err
		}

	case "transfer":
		senderID, err = r.getOrCreateParticipant(ctx, tx, "wallet", userID)
		if err != nil {
			return err
		}
		if txReq.ReceiverAccountID == nil {
			return errors.New("receiver account id required for transfer")
		}
		receiverID, err = r.getOrCreateParticipant(ctx, tx, "wallet", *txReq.ReceiverAccountID)
		if err != nil {
			return err
		}

	default:
		return errors.New("invalid transaction type")
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO transactions (type, amount, total, note, id_sender, id_receiver)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, txReq.Type, txReq.Amount, txReq.Total, txReq.Note, senderID, receiverID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
