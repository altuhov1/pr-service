package storage

import (
	"context"
	"subscription-budget/internal/models"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestPRDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
			"TZ":                "UTC",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	port, err := postgresContainer.MappedPort(ctx, "5432")
	require.NoError(t, err)

	connStr := "postgres://testuser:testpass@localhost:" + port.Port() + "/testdb?sslmode=disable"

	config, err := pgxpool.ParseConfig(connStr)
	require.NoError(t, err)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS pull_requests (
			pull_request_id TEXT PRIMARY KEY,
			pull_request_name TEXT NOT NULL,
			author_id TEXT NOT NULL,
			status TEXT NOT NULL,
			assigned_reviewers TEXT[],
			created_at TIMESTAMP WITH TIME ZONE NOT NULL,
			merged_at TIMESTAMP WITH TIME ZONE
		)
	`)
	require.NoError(t, err)

	t.Cleanup(func() {
		pool.Close()
		postgresContainer.Terminate(ctx)
	})

	return pool
}

func TestPullRequestPostgresStorage_Integration(t *testing.T) {
	pool := setupTestPRDB(t)
	storage := NewPullRequestPostgresStorage(pool)
	ctx := context.Background()

	t.Run("Create and Get PR", func(t *testing.T) {
		tx, err := storage.PRBeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		testPR := models.PullRequest{
			PullRequestID:     "PR-001",
			PullRequestName:   "Test Feature",
			AuthorID:          "user1",
			Status:            "OPEN",
			AssignedReviewers: []string{"user2", "user3"},
			CreatedAt:         time.Now().UTC(),
		}

		err = storage.CreatePRTx(ctx, tx, testPR)
		require.NoError(t, err)

		retrievedPR, err := storage.GetPRByIDTx(ctx, tx, testPR.PullRequestID)
		require.NoError(t, err)

		assert.Equal(t, testPR.PullRequestID, retrievedPR.PullRequestID)
		assert.Equal(t, testPR.PullRequestName, retrievedPR.PullRequestName)
		assert.Equal(t, testPR.AuthorID, retrievedPR.AuthorID)
		assert.Equal(t, testPR.Status, retrievedPR.Status)
		assert.ElementsMatch(t, testPR.AssignedReviewers, retrievedPR.AssignedReviewers)
		assert.False(t, retrievedPR.CreatedAt.IsZero())
		assert.WithinDuration(t, testPR.CreatedAt, retrievedPR.CreatedAt, 5*time.Second)

		err = tx.Commit(ctx)
		require.NoError(t, err)
	})

	t.Run("Merge PR and verify status", func(t *testing.T) {
		tx, err := storage.PRBeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		testPR := models.PullRequest{
			PullRequestID:     "PR-MERGE-TEST",
			PullRequestName:   "Merge Test",
			AuthorID:          "user1",
			Status:            "OPEN",
			AssignedReviewers: []string{"user2"},
			CreatedAt:         time.Now().UTC(),
		}

		err = storage.CreatePRTx(ctx, tx, testPR)
		require.NoError(t, err)

		err = storage.MergePRTx(ctx, tx, testPR.PullRequestID)
		require.NoError(t, err)

		mergedPR, err := storage.GetPRByIDTx(ctx, tx, testPR.PullRequestID)
		require.NoError(t, err)
		assert.Equal(t, "MERGED", mergedPR.Status)
		assert.NotNil(t, mergedPR.MergedAt)
		assert.WithinDuration(t, time.Now().UTC(), *mergedPR.MergedAt, 5*time.Second)

		err = storage.MergePRTx(ctx, tx, testPR.PullRequestID)
		require.NoError(t, err)

		err = tx.Commit(ctx)
		require.NoError(t, err)
	})
}
