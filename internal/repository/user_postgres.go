package repository

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/scarypuppp/gophermart/internal/entities"
)

type UserRepositoryPostgres struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

func NewUserRepositoryPostgres(db *sqlx.DB) UserRepositoryPostgres {
	return UserRepositoryPostgres{db: db}
}

func NewUserRepositoryPostgresTx(tx *sqlx.Tx) UserRepositoryPostgres {
	return UserRepositoryPostgres{tx: tx}
}

func (r *UserRepositoryPostgres) GetByID(ctx context.Context, id int64) (*entities.User, error) {
	var user entities.User
	err := sqlx.GetContext(ctx, r.getExecutor(), &user,
		`SELECT id, login, password FROM users WHERE id = $1`, id,
	)
	if err != nil {
		return nil, fmt.Errorf("GetByID (id=%d): %w", id, err)
	}
	return &user, nil
}

func (r *UserRepositoryPostgres) GetByLogin(ctx context.Context, login string) (*entities.User, error) {
	var user entities.User
	err := sqlx.GetContext(ctx, r.getExecutor(), &user,
		`SELECT id, login, password FROM users WHERE login = $1`, login,
	)
	if err != nil {
		return nil, fmt.Errorf("GetByLogin (login=%s): %w", login, err)
	}
	return &user, nil
}
func (r *UserRepositoryPostgres) CreateUser(ctx context.Context, user entities.User) (*entities.User, error) {
	const insertQuery = `
    INSERT INTO users (login, password)
    VALUES (:login, :password)
    RETURNING id`
	var newID int64
	rows, err := sqlx.NamedQueryContext(ctx, r.getExecutor(), insertQuery, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Next() {
		rows.Scan(&newID)
	}
	user.ID = newID
	return &user, err
}

func (r *UserRepositoryPostgres) getExecutor() sqlx.ExtContext {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}
