package storage

/*
Тесты через создание контейнера с постгрес
Проверка:
	1. Успешно ли создаются команды
	2. Повторное создание команды с тем же именнем
	3. Получение информацие по несуществующему имени
	4. Проверка обновления данных
	5. Проверка на праильно получение информации о пользователе

*/
import (
	"context"
	"fmt"
	"testing"
	"time"

	"test-task/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type testPostgresContainer struct {
	testcontainers.Container
	ConnectionString string
}

func setupTestPostgres(ctx context.Context) (*testPostgresContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:15-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	host, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		return nil, err
	}

	connStr := fmt.Sprintf("postgres://testuser:testpass@%s:%s/testdb?sslmode=disable", host, port.Port())

	return &testPostgresContainer{
		Container:        container,
		ConnectionString: connStr,
	}, nil
}

func createTestTables(ctx context.Context, pool *pgxpool.Pool) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			user_id VARCHAR(50) PRIMARY KEY,
			username VARCHAR(100) NOT NULL,
			team_name VARCHAR(100) NOT NULL,
			is_active BOOLEAN NOT NULL DEFAULT true
		);

		INSERT INTO users (user_id, username, team_name, is_active) VALUES
			('user1', 'john_doe', 'Team Alpha', true),
			('user2', 'jane_smith', 'Team Beta', false),
			('user3', 'bob_wilson', 'Team Gamma', true);
	`

	_, err := pool.Exec(ctx, query)
	return err
}

func setupTestDatabase(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	pgContainer, err := setupTestPostgres(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			t.Logf("failed to terminate container: %s", err)
		}
	})

	pool, err := pgxpool.New(ctx, pgContainer.ConnectionString)
	require.NoError(t, err)
	t.Cleanup(func() {
		pool.Close()
	})

	err = createTestTables(ctx, pool)
	require.NoError(t, err)

	return pool
}

func TestUserPostgresStorage_UpdateUserActive(t *testing.T) {
	pool := setupTestDatabase(t)
	storage := NewUserPostgresStorage(pool)
	ctx := context.Background()

	t.Run("successfully update user active status", func(t *testing.T) {
		tx, err := storage.UserBeginTx(ctx)
		require.NoError(t, err)

		err = storage.UpdateUserActiveTx(ctx, tx, "user1", false)
		require.NoError(t, err)

		err = tx.Commit(ctx)
		require.NoError(t, err)

		tx2, err := storage.UserBeginTx(ctx)
		require.NoError(t, err)
		defer tx2.Rollback(ctx)

		user, err := storage.GetUserTx(ctx, tx2, "user1")
		require.NoError(t, err)
		assert.False(t, user.IsActive)
	})

	t.Run("update non-existent user", func(t *testing.T) {
		tx, err := storage.UserBeginTx(ctx)
		require.NoError(t, err)
		defer tx.Rollback(ctx)

		err = storage.UpdateUserActiveTx(ctx, tx, "nonexistent", true)

		assert.Error(t, err)
		assert.Equal(t, models.ErrNotFound, err)
	})
}

func TestNewUserPostgresStorage(t *testing.T) {
	pool := &pgxpool.Pool{}
	storage := NewUserPostgresStorage(pool)

	assert.NotNil(t, storage)
	assert.Equal(t, pool, storage.pool)
}
