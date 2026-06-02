package repository

import "context"

type UnitOfWork interface {
	Users() UserRepository

	BeginTx(ctx context.Context) (UnitOfWork, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}
