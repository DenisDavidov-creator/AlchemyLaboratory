package transactor

import (
	"context"

	"github.com/jmoiron/sqlx"
)

//go:generate mockery --name=TransactorInterface
type TransactorInterface interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type ctxKey string

const txKey ctxKey = "db_transaction"

type PTransactor struct {
	db *sqlx.DB
}

func NewPtransactor(db *sqlx.DB) *PTransactor {
	return &PTransactor{
		db: db,
	}
}
func (t *PTransactor) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.Beginx()
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txKey, tx)

	err = fn(txCtx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func GetTx(ctx context.Context) *sqlx.Tx {
	tx, ok := ctx.Value(txKey).(*sqlx.Tx)
	if ok {
		return tx
	}
	return nil
}

func InjectTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}
