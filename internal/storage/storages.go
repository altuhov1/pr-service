package storage

import (
	"context"
	"test-task/internal/models"

	"github.com/jackc/pgx/v5"
)

type PullReqStorage interface {
	CreatePRTx(ctx context.Context, tx pgx.Tx, pr models.PullRequest) error
	GetPRByIDTx(ctx context.Context, tx pgx.Tx, prID string) (*models.PullRequest, error)
	MergePRTx(ctx context.Context, tx pgx.Tx, prID string) error
	UpdatePRReviewersTx(ctx context.Context, tx pgx.Tx, prID string, reviewers []string) error
	GetPRsByReviewerTx(ctx context.Context, tx pgx.Tx, userID string) ([]models.PullRequestShort, error)

	PRBeginTx(ctx context.Context) (pgx.Tx, error)
}

type TeamStorage interface {
	CreateTeamTx(ctx context.Context, tx pgx.Tx, team models.Team) error
	GetTeamInfoTx(ctx context.Context, tx pgx.Tx, teamName string) (*models.Team, error)
	TeamBeginTx(ctx context.Context) (pgx.Tx, error)
}

type UserStorage interface {
	GetUserTx(ctx context.Context, tx pgx.Tx, userID string) (*models.User, error)
	UpdateUserActiveTx(ctx context.Context, tx pgx.Tx, userID string, isActive bool) error
	UserBeginTx(ctx context.Context) (pgx.Tx, error)
}
