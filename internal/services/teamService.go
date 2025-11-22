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
)

type TeamService struct {
	storage storage.TeamStorage
}

func NewTeamService(storage storage.TeamStorage) *TeamService {
	return &TeamService{
		storage: storage,
	}
}

func (s *TeamService) CreateTeam(ctx context.Context, team models.Team) (*models.Team, error) {
	tx, err := s.storage.TeamBeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = s.storage.CreateTeamTx(ctx, tx, team)
	if err != nil {
		return nil, err
	}

	createdTeam, err := s.storage.GetTeamInfoTx(ctx, tx, team.TeamName)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return createdTeam, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {

	team, err := s.storage.GetTeamInfoTx(ctx, nil, teamName)
	if err != nil {
		return nil, err
	}

	return team, nil
}
