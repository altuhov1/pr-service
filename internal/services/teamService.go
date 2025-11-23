package services

/*
Функции:
	1. Создание команды
	2. Получение информации о комнаде

Фича - указываем в GetTeamInfoTx nil вместо индекса, он автоматом выполняется через
пул
*/
import (
	"context"
	"test-task/internal/models"
	"test-task/internal/storage"
	"time"
)

type TeamService struct {
	storage storage.TeamStorage
}

func NewTeamService(storage storage.TeamStorage) *TeamService {
	return &TeamService{
		storage: storage,
	}
}

func (s *TeamService) executeWithRetryTeam(ctx context.Context, operation func() error) error {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(time.Duration(attempt) * 90 * time.Millisecond):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err := operation()
		if err == nil {
			return nil
		}

		lastErr = err
	}

	return lastErr
}

func (s *TeamService) CreateTeam(ctx context.Context, team models.Team) (*models.Team, error) {
	var result *models.Team

	err := s.executeWithRetryTeam(ctx, func() error {
		tx, err := s.storage.TeamBeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		err = s.storage.CreateTeamTx(ctx, tx, team)
		if err != nil {
			return err
		}

		createdTeam, err := s.storage.GetTeamInfoTx(ctx, tx, team.TeamName)
		if err != nil {
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}

		result = createdTeam
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	var result *models.Team

	err := s.executeWithRetryTeam(ctx, func() error {
		team, err := s.storage.GetTeamInfoTx(ctx, nil, teamName)
		if err != nil {
			return err
		}

		result = team
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
