package storage

import (
	"context"
	"test-task/internal/models"
)

type PullReqStorage interface {
	CreatePR(ctx context.Context, pr models.PullRequest) error
	GetPRByID(ctx context.Context, prID string) (*models.PullRequest, error)
	MergePR(ctx context.Context, prID string) error
	UpdatePRReviewers(ctx context.Context, prID string, reviewers []string) error
	GetPRsByReviewer(ctx context.Context, userID string) ([]models.PullRequestShort, error)
	CheckPRExists(ctx context.Context, prID string) (bool, error)
}

type TeamStorage interface {
	CreateTeam(ctx context.Context, team models.Team) error
	GetTeamInfo(ctx context.Context, teamName string) (*models.Team, error)
}

type UserStorage interface {
	GetUser(ctx context.Context, userID string) (*models.User, error)
	UpdateUserActive(ctx context.Context, userID string, isActive bool) error
}
