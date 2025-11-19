package storage

import "github.com/jackc/pgx/v5/pgxpool"

type UserPostgresStorage struct {
	pool *pgxpool.Pool
}

func NewUserPostgresStorage(pool *pgxpool.Pool) *UserPostgresStorage {

	return &UserPostgresStorage{pool: pool}
}
