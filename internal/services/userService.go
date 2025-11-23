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
	"time"
)

type UserService struct {
	userStorage storage.UserStorage
}

func NewUserService(userStorage storage.UserStorage) *UserService {
	return &UserService{
		userStorage: userStorage,
	}
}

func (s *UserService) executeWithRetry(ctx context.Context, operation func() error) error {
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

func (s *UserService) SetUserActive(ctx context.Context, userID string, isActive bool) (*models.User, error) {
	var result *models.User

	err := s.executeWithRetry(ctx, func() error {
		tx, err := s.userStorage.UserBeginTx(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)

		err = s.userStorage.UpdateUserActiveTx(ctx, tx, userID, isActive)
		if err != nil {
			return err
		}

		res, err := s.userStorage.GetUserTx(ctx, tx, userID)
		if err != nil {
			return err
		}

		if err := tx.Commit(ctx); err != nil {
			return err
		}

		result = res
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	var result *models.User

	err := s.executeWithRetry(ctx, func() error {
		user, err := s.userStorage.GetUserTx(ctx, nil, userID)
		if err != nil {
			return err
		}

		result = user
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
