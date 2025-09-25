package repositories

import (
	"context"
	"errors"
)

func (ur *Auth) CreatePIN(rctx context.Context, newPin string, uid int) error {
	sql := `
		UPDATE accounts
		SET pin = $1
		WHERE id = $2
	`

	ctag, err := ur.db.Exec(rctx, sql, newPin, uid)
	if err != nil {
		return err
	}
	if ctag.RowsAffected() == 0 {
		return errors.New("unable to create user's PIN")
	}

	return nil
}
