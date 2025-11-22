package services
/*
Функции:
	1. Создание pr
	2. Merge
	3. Переназначение пользоватля
	4. По пользователю найти Ревью

Основная сложность в написании сервиса была связана с возможным рейс кондишн.
Было исправлено за счет транзакций 
*/
import (
	"context"
	"strings"
	"test-task/internal/models"
	"test-task/internal/storage"

	"github.com/jackc/pgx/v5"
)

type PullRequestService struct {
	PullRequestServ storage.PullReqStorage
	userStorage     storage.UserStorage
	teamStorage     storage.TeamStorage
}

func NewPullRequestService(
	PullRequestServ storage.PullReqStorage,
	userStorage storage.UserStorage,
	teamStorage storage.TeamStorage,
) *PullRequestService {
	return &PullRequestService{
		PullRequestServ: PullRequestServ,
		userStorage:     userStorage,
		teamStorage:     teamStorage,
	}
}

func (s *PullRequestService) CreatePR(ctx context.Context, req models.CreatePRRequest) (*models.PullRequest, error) {
	tx, err := s.PullRequestServ.PRBeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	author, err := s.userStorage.GetUserTx(ctx, tx, req.AuthorID)
	if err != nil {
		return nil, models.ErrNotFound
	}

	team, err := s.teamStorage.GetTeamInfoTx(ctx, tx, author.TeamName)
	if err != nil {
		return nil, models.ErrNotFound
	}

	reviewers := s.findReviewersFromTeam(team, req.AuthorID)

	pr := models.PullRequest{
		PullRequestID:     req.PullRequestID,
		PullRequestName:   req.PullRequestName,
		AuthorID:          req.AuthorID,
		Status:            "OPEN",
		AssignedReviewers: reviewers,
	}

	err = s.PullRequestServ.CreatePRTx(ctx, tx, pr)
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, models.ErrPRExists
		}
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &pr, nil
}

func (s *PullRequestService) findReviewersFromTeam(team *models.Team, authorID string) []string {
	var reviewers []string
	for _, member := range team.Members {
		if member.UserID == authorID || !member.IsActive {
			continue
		}
		reviewers = append(reviewers, member.UserID)
		if len(reviewers) >= 2 {
			break
		}
	}
	return reviewers
}

func (s *PullRequestService) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	tx, err := s.PullRequestServ.PRBeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = s.PullRequestServ.GetPRByIDTx(ctx, tx, prID)
	if err != nil {
		return nil, models.ErrNotFound
	}

	err = s.PullRequestServ.MergePRTx(ctx, tx, prID)
	if err != nil {
		return nil, err
	}

	pr, err := s.PullRequestServ.GetPRByIDTx(ctx, tx, prID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return pr, nil
}

func (s *PullRequestService) ReassignReviewer(ctx context.Context, req models.ReassignRequest) (*models.PullRequest, string, error) {
	tx, err := s.PullRequestServ.PRBeginTx(ctx)
	if err != nil {
		return nil, "", err
	}
	defer tx.Rollback(ctx)

	pr, err := s.PullRequestServ.GetPRByIDTx(ctx, tx, req.PullRequestID)
	if err != nil {
		return nil, "", models.ErrNotFound
	}

	if pr.Status == "MERGED" {
		return nil, "", models.ErrPRMerged
	}

	if !contains(pr.AssignedReviewers, req.OldUserID) {
		return nil, "", models.ErrNotAssigned
	}

	author, err := s.userStorage.GetUserTx(ctx, tx, pr.AuthorID)
	if err != nil {
		return nil, "", models.ErrNotFound
	}

	newReviewer, err := s.findReplacementReviewer(ctx, tx, author.TeamName, pr.AssignedReviewers, req.OldUserID, pr.AuthorID)
	if err != nil {
		return nil, "", models.ErrNoCandidate
	}

	newReviewers := replaceInSlice(pr.AssignedReviewers, req.OldUserID, newReviewer)
	err = s.PullRequestServ.UpdatePRReviewersTx(ctx, tx, req.PullRequestID, newReviewers)
	if err != nil {
		return nil, "", err
	}

	updatedPR, err := s.PullRequestServ.GetPRByIDTx(ctx, tx, req.PullRequestID)
	if err != nil {
		return nil, "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, "", err
	}

	return updatedPR, newReviewer, nil
}

func (s *PullRequestService) GetUserReviews(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	tx, err := s.PullRequestServ.PRBeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = s.userStorage.GetUserTx(ctx, tx, userID)
	if err != nil {
		return nil, models.ErrNotFound
	}

	prs, err := s.PullRequestServ.GetPRsByReviewerTx(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return prs, nil
}

func (s *PullRequestService) findReplacementReviewer(ctx context.Context, tx pgx.Tx, teamName string, currentReviewers []string, oldUserID string, authorID string) (string, error) {
	team, err := s.teamStorage.GetTeamInfoTx(ctx, tx, teamName)
	if err != nil {
		return "", err
	}

	for _, member := range team.Members {
		if member.UserID == authorID ||
			!member.IsActive ||
			contains(currentReviewers, member.UserID) ||
			member.UserID == oldUserID {
			continue
		}
		return member.UserID, nil
	}

	return "", models.ErrNoCandidate
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func replaceInSlice(slice []string, old string, new string) []string {
	result := make([]string, len(slice))
	for i, item := range slice {
		if item == old {
			result[i] = new
		} else {
			result[i] = item
		}
	}
	return result
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errorStr := err.Error()
	return strings.Contains(errorStr, "unique constraint") ||
		strings.Contains(errorStr, "duplicate key") ||
		err == models.ErrPRExists
}
