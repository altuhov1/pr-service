package storage

/*
Основные фукнции:
	1. Получение данных о юзере по индексу
	2. Обновление активности юзера
	3. Создать транзакцию

Фича - если Tx - nil, то используем просто pool
*/

import (
	"context"
	"fmt"
	"subscription-budget/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPostgresStorage struct {
	pool *pgxpool.Pool
}

func NewUserPostgresStorage(pool *pgxpool.Pool) *UserPostgresStorage {
	return &UserPostgresStorage{pool: pool}
}

func (s *UserPostgresStorage) UserBeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

func (s *UserPostgresStorage) GetUserTx(ctx context.Context, tx pgx.Tx, userID string) (*models.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users 
		WHERE user_id = $1
	`

	var user models.User
	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, query, userID)
	} else {
		row = s.pool.QueryRow(ctx, query, userID)
	}

	err := row.Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *UserPostgresStorage) UpdateUserActiveTx(ctx context.Context, tx pgx.Tx, userID string, isActive bool) error {
	query := `
		UPDATE users 
		SET is_active = $1
		WHERE user_id = $2
	`

	var result pgconn.CommandTag
	var err error

	if tx != nil {
		result, err = tx.Exec(ctx, query, isActive, userID)
	} else {
		result, err = s.pool.Exec(ctx, query, isActive, userID)
	}

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrNotFound
	}

	return nil
}
