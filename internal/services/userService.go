package services

/*
Функции:
	1. Выставление активности пользоватлеля
	2. Получение информации о юзере

Фича - указываем в GetUserTx nil вместо индекса, он автоматом выполняется через
пул
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
	tx, err := s.userStorage.UserBeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	err = s.userStorage.UpdateUserActiveTx(ctx, tx, userID, isActive)
	if err != nil {
		return nil, err
	}

	res, err := s.userStorage.GetUserTx(ctx, tx, userID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return res, nil
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*models.User, error) {

	user, err := s.userStorage.GetUserTx(ctx, nil, userID)
	if err != nil {
		return nil, err
	}

	return user, nil
}
