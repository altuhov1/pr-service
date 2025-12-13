package storage

/*
Основные функции:
	1. Создание команды
	2. Получение информации о команде
	3. Создать транзакцию

Создание команды проихсодит атомарно.
При создании происходит проверка через SQL запрос на то, существет
ли человек, если да - обновляем его данные.

Поиск юзеров за log из-за индексов

Фича - если Tx - nil, то используем просто pool
*/

import (
	"context"
	"fmt"
	"subscription-budget/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamPostgresStorage struct {
	pool *pgxpool.Pool
}

func NewTeamPostgresStorage(pool *pgxpool.Pool) *TeamPostgresStorage {
	return &TeamPostgresStorage{pool: pool}
}

func (s *TeamPostgresStorage) TeamBeginTx(ctx context.Context) (pgx.Tx, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	return tx, nil
}

func (s *TeamPostgresStorage) CreateTeamTx(ctx context.Context, tx pgx.Tx, team models.Team) error {
	var exists bool
	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)", team.TeamName)
	} else {
		row = s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)", team.TeamName)
	}

	err := row.Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check team existence: %w", err)
	}
	if exists {
		return models.ErrTeamExists
	}

	if tx != nil {
		_, err = tx.Exec(ctx, "INSERT INTO teams (name) VALUES ($1)", team.TeamName)
	} else {
		_, err = s.pool.Exec(ctx, "INSERT INTO teams (name) VALUES ($1)", team.TeamName)
	}
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	for _, member := range team.Members {
		if err := s.createUserTx(ctx, tx, member); err != nil {
			return fmt.Errorf("failed to create user %s: %w", member.UserID, err)
		}
	}

	return nil
}

func (s *TeamPostgresStorage) createUserTx(ctx context.Context, tx pgx.Tx, user models.User) error {
	query := `
		INSERT INTO users (user_id, username, team_name, is_active) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) 
		DO UPDATE SET username = EXCLUDED.username, team_name = EXCLUDED.team_name, is_active = EXCLUDED.is_active
	`

	if tx != nil {
		_, err := tx.Exec(ctx, query, user.UserID, user.Username, user.TeamName, user.IsActive)
		if err != nil {
			return fmt.Errorf("failed to create/update user: %w", err)
		}
	} else {
		_, err := s.pool.Exec(ctx, query, user.UserID, user.Username, user.TeamName, user.IsActive)
		if err != nil {
			return fmt.Errorf("failed to create/update user: %w", err)
		}
	}
	return nil
}

func (s *TeamPostgresStorage) GetTeamInfoTx(ctx context.Context, tx pgx.Tx, teamName string) (*models.Team, error) {
	query := `
        SELECT 
            t.name as team_name, 
            u.user_id, 
            u.username, 
            u.team_name, 
            u.is_active
        FROM teams t
        JOIN users u ON u.team_name = t.name
        WHERE t.name = $1
        ORDER BY u.user_id
    `

	var rows pgx.Rows
	var err error

	if tx != nil {
		rows, err = tx.Query(ctx, query, teamName)
	} else {
		rows, err = s.pool.Query(ctx, query, teamName)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query team: %w", err)
	}
	defer rows.Close()

	var team models.Team
	var members []models.User

	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&team.TeamName,
			&user.UserID,
			&user.Username,
			&user.TeamName,
			&user.IsActive,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		members = append(members, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating team members: %w", err)
	}

	if len(members) == 0 {
		return nil, models.ErrNotFound
	}

	team.Members = members
	return &team, nil
}
