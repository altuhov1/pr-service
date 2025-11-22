package services

/*
Основные функции:
	1. Создание пул ревеста(по айди человека находим команду, находим свободных
и отправяем в бд)
	2. Merge 
	3. Переопределение пользователя на pr
	4. Посмотреть  по юзер айди ревьюеров
*/
import (
	"context"
	"strings"
	"test-task/internal/models"
	"test-task/internal/storage"
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
	author, err := s.userStorage.GetUser(ctx, req.AuthorID)
	if err != nil {
		return nil, models.ErrNotFound
	}

	team, err := s.teamStorage.GetTeamInfo(ctx, author.TeamName)
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

	err = s.PullRequestServ.CreatePR(ctx, pr)
	if err != nil {
		if isUniqueConstraintError(err) {
			return nil, models.ErrPRExists
		}
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
	_, err := s.PullRequestServ.GetPRByID(ctx, prID)
	if err != nil {
		return nil, models.ErrNotFound
	}
	err = s.PullRequestServ.MergePR(ctx, prID)
	if err != nil {
		return nil, err
	}

	return s.PullRequestServ.GetPRByID(ctx, prID)
}

func (s *PullRequestService) ReassignReviewer(ctx context.Context, req models.ReassignRequest) (*models.PullRequest, string, error) {
	pr, err := s.PullRequestServ.GetPRByID(ctx, req.PullRequestID)
	if err != nil {
		return nil, "", models.ErrNotFound
	}

	if pr.Status == "MERGED" {
		return nil, "", models.ErrPRMerged
	}

	if !contains(pr.AssignedReviewers, req.OldUserID) {
		return nil, "", models.ErrNotAssigned
	}

	author, err := s.userStorage.GetUser(ctx, pr.AuthorID)
	if err != nil {
		return nil, "", models.ErrNotFound
	}

	newReviewer, err := s.findReplacementReviewer(ctx, author.TeamName, pr.AssignedReviewers, req.OldUserID, pr.AuthorID)
	if err != nil {
		return nil, "", models.ErrNoCandidate
	}

	newReviewers := replaceInSlice(pr.AssignedReviewers, req.OldUserID, newReviewer)
	err = s.PullRequestServ.UpdatePRReviewers(ctx, req.PullRequestID, newReviewers)
	if err != nil {
		return nil, "", err
	}

	updatedPR, err := s.PullRequestServ.GetPRByID(ctx, req.PullRequestID)
	return updatedPR, newReviewer, err
}

func (s *PullRequestService) GetUserReviews(ctx context.Context, userID string) ([]models.PullRequestShort, error) {
	_, err := s.userStorage.GetUser(ctx, userID)
	if err != nil {
		return nil, models.ErrNotFound
	}

	return s.PullRequestServ.GetPRsByReviewer(ctx, userID)
}

func (s *PullRequestService) findReplacementReviewer(ctx context.Context, teamName string, currentReviewers []string, oldUserID string, authorID string) (string, error) {
	team, err := s.teamStorage.GetTeamInfo(ctx, teamName)
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
