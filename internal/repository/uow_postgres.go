package repository

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type UnitOfWorkPostgres struct {
	db    *sqlx.DB
	tx    *sqlx.Tx
	users *UserRepositoryPostgres
}

func NewUnitOfWorkPostgres(db *sqlx.DB) *UnitOfWorkPostgres {
	return &UnitOfWorkPostgres{
		db:    db,
		users: NewUserRepositoryPostgres(db),
	}
}

func (u *UnitOfWorkPostgres) Users() UserRepository {
	return u.users
}

func (u *UnitOfWorkPostgres) BeginTx(ctx context.Context) (UnitOfWork, error) {
	tx, err := u.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &UnitOfWorkPostgres{
		db:    u.db,
		tx:    tx,
		users: NewUserRepositoryPostgresTx(tx),
	}, nil
}

func (u *UnitOfWorkPostgres) Commit(_ context.Context) error {
	if u.tx == nil {
		return nil
	}
	return u.tx.Commit()
}

func (u *UnitOfWorkPostgres) Rollback(_ context.Context) error {
	if u.tx == nil {
		return nil
	}
	return u.tx.Rollback()
}
