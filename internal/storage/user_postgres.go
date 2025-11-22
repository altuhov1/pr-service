package storage

/*
Основные фукнции:
	1. Получение данных о юзере по индексу
	2. Обновление активности юзера
*/
import (
	"context"
	"fmt"
	"test-task/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserPostgresStorage struct {
	pool *pgxpool.Pool
}

func NewUserPostgresStorage(pool *pgxpool.Pool) *UserPostgresStorage {
	return &UserPostgresStorage{pool: pool}
}

func (s *UserPostgresStorage) GetUser(ctx context.Context, userID string) (*models.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users 
		WHERE user_id = $1
	`

	var user models.User
	err := s.pool.QueryRow(ctx, query, userID).Scan(
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

func (s *UserPostgresStorage) UpdateUserActive(ctx context.Context, userID string, isActive bool) error {
	query := `
		UPDATE users 
		SET is_active = $1
		WHERE user_id = $2
	`

	result, err := s.pool.Exec(ctx, query, isActive, userID)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrNotFound
	}

	return nil
}
