package storage

/*
Основные функции:
	1. Создание команды
	2. Получение информации о команде

Создание команды проихсодит атомарно.
При создании происходит проверка через SQL запрос на то, существет
ли человек, если да - обновляем его данные.

Поиск юзеров за log из-за индексов
*/
import (
	"context"
	"fmt"
	"test-task/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamPostgresStorage struct {
	pool *pgxpool.Pool
}

func NewTeamPostgresStorage(pool *pgxpool.Pool) *TeamPostgresStorage {
	return &TeamPostgresStorage{pool: pool}
}

func (s *TeamPostgresStorage) CreateTeam(ctx context.Context, team models.Team) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM teams WHERE name = $1)", team.TeamName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check team existence: %w", err)
	}
	if exists {
		return models.ErrTeamExists
	}

	_, err = tx.Exec(ctx, "INSERT INTO teams (name) VALUES ($1)", team.TeamName)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	for _, member := range team.Members {
		if err := s.createUser(ctx, tx, member); err != nil {
			return fmt.Errorf("failed to create user %s: %w", member.UserID, err)
		}
	}

	return tx.Commit(ctx)
}

func (s *TeamPostgresStorage) createUser(ctx context.Context, tx pgx.Tx, user models.User) error {
	query := `
		INSERT INTO users (user_id, username, team_name, is_active) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) 
		DO UPDATE SET username = EXCLUDED.username, team_name = EXCLUDED.team_name, is_active = EXCLUDED.is_active
	`
	_, err := tx.Exec(ctx, query, user.UserID, user.Username, user.TeamName, user.IsActive)
	if err != nil {
		return fmt.Errorf("failed to create/update user: %w", err)
	}
	return nil
}

func (s *TeamPostgresStorage) GetTeamInfo(ctx context.Context, teamName string) (*models.Team, error) {
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

	rows, err := s.pool.Query(ctx, query, teamName)
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
