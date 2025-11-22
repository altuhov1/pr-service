package storage

/*
Основные функции:
	1. Создание PR
	2. Получение PR по id
	3. Merge
	4. Обновить ревьюеров
	5. По ревьюеру найти PR
	6. Проверить существование PR
	7. Создать транзакцию



Если у нас уже  "Merge" в таблице PR, то при выполнении функции Merge у нас
ничего не произойдет, все произодйте в штатном порядке
*/

import (
	"context"
	"fmt"
	"test-task/internal/models"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PullRequestPostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPullRequestPostgresStorage(pool *pgxpool.Pool) *PullRequestPostgresStorage {
	return &PullRequestPostgresStorage{pool: pool}
}

func (s *PullRequestPostgresStorage) CreatePRTx(ctx context.Context, tx pgx.Tx, pr models.PullRequest) error {
	query := `
		INSERT INTO pull_requests (
			pull_request_id, 
			pull_request_name, 
			author_id, 
			status, 
			assigned_reviewers,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := tx.Exec(ctx, query,
		pr.PullRequestID,
		pr.PullRequestName,
		pr.AuthorID,
		pr.Status,
		pr.AssignedReviewers,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	return nil
}

func (s *PullRequestPostgresStorage) GetPRByIDTx(ctx context.Context, tx pgx.Tx, prID string) (*models.PullRequest, error) {
	query := `
		SELECT 
			pull_request_id,
			pull_request_name,
			author_id,
			status,
			assigned_reviewers,
			created_at,
			merged_at
		FROM pull_requests 
		WHERE pull_request_id = $1
	`

	var pr models.PullRequest
	var mergedAt *time.Time

	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, query, prID)
	} else {
		row = s.pool.QueryRow(ctx, query, prID)
	}

	err := row.Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&pr.AssignedReviewers,
		&pr.CreatedAt,
		&mergedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, models.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	if mergedAt != nil {
		pr.MergedAt = mergedAt
	}

	return &pr, nil
}

func (s *PullRequestPostgresStorage) MergePRTx(ctx context.Context, tx pgx.Tx, prID string) error {
	query := `
		UPDATE pull_requests 
		SET status = $1, merged_at = $2
		WHERE pull_request_id = $3 AND status != $1
	`

	var result pgconn.CommandTag
	var err error

	if tx != nil {
		result, err = tx.Exec(ctx, query, "MERGED", time.Now(), prID)
	} else {
		result, err = s.pool.Exec(ctx, query, "MERGED", time.Now(), prID)
	}

	if err != nil {
		return fmt.Errorf("failed to merge PR: %w", err)
	}

	if result.RowsAffected() == 0 {
		exists, err := s.checkPRExistsTx(ctx, tx, prID)
		if err != nil {
			return err
		}
		if !exists {
			return models.ErrNotFound
		}
	}

	return nil
}

func (s *PullRequestPostgresStorage) UpdatePRReviewersTx(ctx context.Context, tx pgx.Tx, prID string, reviewers []string) error {
	query := `
		UPDATE pull_requests 
		SET assigned_reviewers = $1
		WHERE pull_request_id = $2 AND status = $3
	`

	var result pgconn.CommandTag
	var err error

	if tx != nil {
		result, err = tx.Exec(ctx, query, reviewers, prID, "OPEN")
	} else {
		result, err = s.pool.Exec(ctx, query, reviewers, prID, "OPEN")
	}

	if err != nil {
		return fmt.Errorf("failed to update PR reviewers: %w", err)
	}

	if result.RowsAffected() == 0 {
		return models.ErrPRMerged
	}

	return nil
}

func (s *PullRequestPostgresStorage) GetPRsByReviewerTx(ctx context.Context, tx pgx.Tx, userID string) ([]models.PullRequestShort, error) {
	query := `
		SELECT 
			pull_request_id,
			pull_request_name,
			author_id,
			status
		FROM pull_requests 
		WHERE $1 = ANY(assigned_reviewers)
		ORDER BY created_at DESC
	`

	var rows pgx.Rows
	var err error

	if tx != nil {
		rows, err = tx.Query(ctx, query, userID)
	} else {
		rows, err = s.pool.Query(ctx, query, userID)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to query PRs by reviewer: %w", err)
	}
	defer rows.Close()

	var prs []models.PullRequestShort
	for rows.Next() {
		var pr models.PullRequestShort
		err := rows.Scan(
			&pr.PullRequestID,
			&pr.PullRequestName,
			&pr.AuthorID,
			&pr.Status,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating PRs: %w", err)
	}

	return prs, nil
}

func (s *PullRequestPostgresStorage) PRBeginTx(ctx context.Context) (pgx.Tx, error) {
	return s.pool.BeginTx(ctx, pgx.TxOptions{
		IsoLevel:   pgx.Serializable,
		AccessMode: pgx.ReadWrite,
	})
}

func (s *PullRequestPostgresStorage) checkPRExistsTx(ctx context.Context, tx pgx.Tx, prID string) (bool, error) {
	var exists bool

	var row pgx.Row
	if tx != nil {
		row = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", prID)
	} else {
		row = s.pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", prID)
	}

	err := row.Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check PR existence: %w", err)
	}

	return exists, nil
}
