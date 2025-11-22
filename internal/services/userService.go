package services
/*
Функции:
	1. Изменение активности пользователя
	2.
*/
import (
	"context"
	"test-task/internal/models"
	"test-task/internal/storage"
)

type UserService struct {
	userStorage storage.UserStorage
}

func NewUserService(userStorage storage.UserStorage) *UserService {
	return &UserService{
		userStorage: userStorage,
	}
}

func (s *UserService) SetUserActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	err := s.userStorage.UpdateUserActive(ctx, userID, isActive)
	if err != nil {
		return nil, err
	}

	return s.userStorage.GetUser(ctx, userID)
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	return s.userStorage.GetUser(ctx, userID)
}
