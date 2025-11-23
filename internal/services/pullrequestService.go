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
	"time"

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

func (s *PullRequestService) executeWithRetry(ctx context.Context, operation func() error) error {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(time.Duration(attempt) * 100 * time.Millisecond):
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

func (s *PullRequestService) CreatePR(ctx context.Context, req models.CreatePRRequest) (*models.PullRequest, error) {
	var result *models.PullRequest

	err := s.executeWithRetry(ctx, func() error {
		tx, err := s.PullRequestServ.PRBeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		author, err := s.userStorage.GetUserTx(ctx, tx, req.AuthorID)
		if err != nil {
			return models.ErrNotFound
		}

		team, err := s.teamStorage.GetTeamInfoTx(ctx, tx, author.TeamName)
		if err != nil {
			return models.ErrNotFound
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
				return models.ErrPRExists
			}
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}

		result = &pr
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *PullRequestService) MergePR(ctx context.Context, prID string) (*models.PullRequest, error) {
	var result *models.PullRequest

	err := s.executeWithRetry(ctx, func() error {
		tx, err := s.PullRequestServ.PRBeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		pr, err := s.PullRequestServ.GetPRByIDTx(ctx, tx, prID)
		if err != nil {
			return models.ErrNotFound
		}

		if pr.Status == "MERGED" {
			result = pr
			return nil
		}

		err = s.PullRequestServ.MergePRTx(ctx, tx, prID)
		if err != nil {
			return err
		}

		pr, err = s.PullRequestServ.GetPRByIDTx(ctx, tx, prID)
		if err != nil {
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}

		result = pr
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *PullRequestService) ReassignReviewer(ctx context.Context, req models.ReassignRequest) (*models.PullRequest, string, error) {
	var resultPR *models.PullRequest
	var resultReviewer string

	err := s.executeWithRetry(ctx, func() error {
		tx, err := s.PullRequestServ.PRBeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		pr, err := s.PullRequestServ.GetPRByIDTx(ctx, tx, req.PullRequestID)
		if err != nil {
			return models.ErrNotFound
		}

		if pr.Status == "MERGED" {
			return models.ErrPRMerged
		}

		if !contains(pr.AssignedReviewers, req.OldUserID) {
			return models.ErrNotAssigned
		}

		author, err := s.userStorage.GetUserTx(ctx, tx, pr.AuthorID)
		if err != nil {
			return models.ErrNotFound
		}

		newReviewer, err := s.findReplacementReviewer(ctx, tx, author.TeamName, pr.AssignedReviewers, req.OldUserID, pr.AuthorID)
		if err != nil {
			return models.ErrNoCandidate
		}

		newReviewers := replaceInSlice(pr.AssignedReviewers, req.OldUserID, newReviewer)
		err = s.PullRequestServ.UpdatePRReviewersTx(ctx, tx, req.PullRequestID, newReviewers)
		if err != nil {
			return err
		}

		updatedPR, err := s.PullRequestServ.GetPRByIDTx(ctx, tx, req.PullRequestID)
		if err != nil {
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}

		resultPR = updatedPR
		resultReviewer = newReviewer
		return nil
	})

	if err != nil {
		return nil, "", err
	}

	return resultPR, resultReviewer, nil
}

func (s *PullRequestService) GetUserReviews(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	var result []models.PullRequestShort

	err := s.executeWithRetry(ctx, func() error {
		tx, err := s.PullRequestServ.PRBeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		_, err = s.userStorage.GetUserTx(ctx, tx, userID)
		if err != nil {
			return models.ErrNotFound
		}

		prs, err := s.PullRequestServ.GetPRsByReviewerTx(ctx, tx, userID)
		if err != nil {
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}

		result = prs
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
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
