package services

import (
	"context"
	"subscription-budget/internal/models"
)

type TeamManager interface {
	CreateTeam(ctx context.Context, team models.Team) (*models.Team, error)
	GetTeam(ctx context.Context, teamName string) (*models.Team, error)
}

type UserManager interface {
	SetUserActive(ctx context.Context, userID string, isActive bool) (*models.User, error)
}

type PullRequestManager interface {
	CreatePR(ctx context.Context, req models.CreatePRRequest) (*models.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*models.PullRequest, error)
	ReassignReviewer(ctx context.Context, req models.ReassignRequest) (*models.PullRequest, string, error)
	GetUserReviews(ctx context.Context, userID string) ([]models.PullRequestShort, error)
}
