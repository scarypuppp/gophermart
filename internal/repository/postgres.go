package repository

import "github.com/jmoiron/sqlx"

func NewPostgresDB(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}
