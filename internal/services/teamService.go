package services

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
	err := s.storage.CreateTeam(ctx, team)
	if err != nil {
		return nil, err
	}

	createdTeam, err := s.storage.GetTeamInfo(ctx, team.TeamName)
	if err != nil {
		return nil, err
	}
	return createdTeam, nil
}

func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*models.Team, error) {
	team, err := s.storage.GetTeamInfo(ctx, teamName)
	if err != nil {
		return nil, err
	}
	return team, nil
}
