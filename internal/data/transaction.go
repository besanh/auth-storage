package data

import (
	"context"
	"database/sql"
	"server/internal/biz"
)

type ctxTransactionKey struct{}

type transactionManager struct {
	db *sql.DB
}

func NewTransactionManager(d *Data) biz.Transaction {
	return &transactionManager{db: d.DB}
}

func (tm *transactionManager) ExecTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	if err := fn(context.WithValue(ctx, ctxTransactionKey{}, tx)); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func FromContext(ctx context.Context) (*sql.Tx, bool) {
	tx, ok := ctx.Value(ctxTransactionKey{}).(*sql.Tx)
	return tx, ok
}
