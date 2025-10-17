package repositories

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prospera/internals/models"
)

type InternalAccountRepository struct {
	db *pgxpool.Pool
}

func NewInternalAccountRepository(db *pgxpool.Pool) *InternalAccountRepository {
	return &InternalAccountRepository{db: db}
}

func (r *InternalAccountRepository) GetAll(ctx context.Context) ([]models.InternalAccount, error) {
	query := `
		SELECT id, name, img, tax, created_at
		FROM internal_accounts
		ORDER BY id ASC;
	`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []models.InternalAccount
	for rows.Next() {
		var ia models.InternalAccount
		if err := rows.Scan(
			&ia.ID,
			&ia.Name,
			&ia.Img,
			&ia.Tax,
			&ia.CreatedAt,
		); err != nil {
			return nil, err
		}
		accounts = append(accounts, ia)
	}

	return accounts, nil
}
