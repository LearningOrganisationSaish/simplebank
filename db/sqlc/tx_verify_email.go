package db

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
)

type VerifyEmailTxParams struct {
	VerifyId   int64
	SecretCode string
}

type VerifyEmailTxResult struct {
	User        User
	VerifyEmail VerifyEmail
}

func (store *SQLStore) VerifyEmailTx(ctx context.Context, params VerifyEmailTxParams) (VerifyEmailTxResult, error) {
	var result VerifyEmailTxResult

	err := store.execTx(ctx, func(queries *Queries) error {
		var err error

		result.VerifyEmail, err = queries.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:         params.VerifyId,
			SecretCode: params.SecretCode,
		})
		if err != nil {
			return err
		}

		result.User, err = queries.UpdateUser(ctx, UpdateUserParams{IsEmailVerified: pgtype.Bool{
			Bool:  true,
			Valid: true,
		}, Username: result.VerifyEmail.Username})
		return err
	})

	return result, err

}
