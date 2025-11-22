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
	"test-task/internal/models"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	container, err := postgres.Run(ctx,
		"postgres:15-alpine",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2),
		),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, container.Terminate(ctx))
	})

	connStr, err := container.ConnectionString(ctx)
	require.NoError(t, err)

	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS teams (
			name TEXT PRIMARY KEY
		);
		
		CREATE TABLE IF NOT EXISTS users (
			user_id TEXT PRIMARY KEY,
			username TEXT NOT NULL,
			team_name TEXT NOT NULL REFERENCES teams(name) ON DELETE CASCADE,
			is_active BOOLEAN NOT NULL DEFAULT true
		);

		CREATE INDEX IF NOT EXISTS idx_users_team ON users(team_name);
		CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);
	`)
	require.NoError(t, err)

	return pool
}

func TestTeamPostgresStorage_CreateTeam_Success(t *testing.T) {
	pool := setupTestDB(t)
	storage := NewTeamPostgresStorage(pool)
	ctx := context.Background()

	team := models.Team{
		TeamName: "backend",
		Members: []models.User{
			{UserID: "u1", Username: "Alice", TeamName: "backend", IsActive: true},
			{UserID: "u2", Username: "Bob", TeamName: "backend", IsActive: true},
		},
	}

	tx, err := storage.TeamBeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	err = storage.CreateTeamTx(ctx, tx, team)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	tx, err = storage.TeamBeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	createdTeam, err := storage.GetTeamInfoTx(ctx, tx, "backend")
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	assert.Equal(t, "backend", createdTeam.TeamName)
	assert.Len(t, createdTeam.Members, 2)
}

func TestTeamPostgresStorage_CreateTeam_AlreadyExists(t *testing.T) {
	pool := setupTestDB(t)
	storage := NewTeamPostgresStorage(pool)
	ctx := context.Background()

	team := models.Team{
		TeamName: "payments",
		Members: []models.User{
			{UserID: "u1", Username: "Alice", TeamName: "payments", IsActive: true},
		},
	}

	tx, err := storage.TeamBeginTx(ctx)
	require.NoError(t, err)

	err = storage.CreateTeamTx(ctx, tx, team)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	tx, err = storage.TeamBeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	err = storage.CreateTeamTx(ctx, tx, team)
	assert.ErrorIs(t, err, models.ErrTeamExists)

	tx.Rollback(ctx)
}

func TestTeamPostgresStorage_GetTeamInfo_NotFound(t *testing.T) {
	pool := setupTestDB(t)
	storage := NewTeamPostgresStorage(pool)
	ctx := context.Background()

	tx, err := storage.TeamBeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	team, err := storage.GetTeamInfoTx(ctx, tx, "nonexistent")
	assert.ErrorIs(t, err, models.ErrNotFound)
	assert.Nil(t, team)
}

func TestTeamPostgresStorage_CreateTeam_UpdatesUserTeam(t *testing.T) {
	pool := setupTestDB(t)
	storage := NewTeamPostgresStorage(pool)
	ctx := context.Background()

	tx, err := storage.TeamBeginTx(ctx)
	require.NoError(t, err)

	team1 := models.Team{
		TeamName: "team1",
		Members: []models.User{
			{UserID: "u1", Username: "Alice", TeamName: "team1", IsActive: true},
		},
	}
	err = storage.CreateTeamTx(ctx, tx, team1)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	tx, err = storage.TeamBeginTx(ctx)
	require.NoError(t, err)

	team2 := models.Team{
		TeamName: "team2",
		Members: []models.User{
			{UserID: "u1", Username: "Alice", TeamName: "team2", IsActive: false},
		},
	}
	err = storage.CreateTeamTx(ctx, tx, team2)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	tx, err = storage.TeamBeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	team, err := storage.GetTeamInfoTx(ctx, tx, "team2")
	require.NoError(t, err)
	assert.Equal(t, "team2", team.Members[0].TeamName)
	assert.False(t, team.Members[0].IsActive)
}

func TestTeamPostgresStorage_GetTeamInfo_Success(t *testing.T) {
	pool := setupTestDB(t)
	storage := NewTeamPostgresStorage(pool)
	ctx := context.Background()

	expectedTeam := models.Team{
		TeamName: "frontend",
		Members: []models.User{
			{UserID: "u1", Username: "Alice", TeamName: "frontend", IsActive: true},
			{UserID: "u2", Username: "Bob", TeamName: "frontend", IsActive: true},
			{UserID: "u3", Username: "Charlie", TeamName: "frontend", IsActive: false},
		},
	}

	tx, err := storage.TeamBeginTx(ctx)
	require.NoError(t, err)

	err = storage.CreateTeamTx(ctx, tx, expectedTeam)
	require.NoError(t, err)

	err = tx.Commit(ctx)
	require.NoError(t, err)

	tx, err = storage.TeamBeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	actualTeam, err := storage.GetTeamInfoTx(ctx, tx, "frontend")
	require.NoError(t, err)
	require.NotNil(t, actualTeam)

	assert.Equal(t, expectedTeam.TeamName, actualTeam.TeamName)
	assert.Len(t, actualTeam.Members, 3)

	memberMap := make(map[string]models.User)
	for _, member := range actualTeam.Members {
		memberMap[member.UserID] = member
	}

	assert.Contains(t, memberMap, "u1")
	assert.Equal(t, "Alice", memberMap["u1"].Username)
	assert.Equal(t, "frontend", memberMap["u1"].TeamName)
	assert.True(t, memberMap["u1"].IsActive)

	assert.Contains(t, memberMap, "u2")
	assert.Equal(t, "Bob", memberMap["u2"].Username)
	assert.Equal(t, "frontend", memberMap["u2"].TeamName)
	assert.True(t, memberMap["u2"].IsActive)

	assert.Contains(t, memberMap, "u3")
	assert.Equal(t, "Charlie", memberMap["u3"].Username)
	assert.Equal(t, "frontend", memberMap["u3"].TeamName)
	assert.False(t, memberMap["u3"].IsActive)
}

func TestTeamPostgresStorage_Transaction_Rollback(t *testing.T) {
	pool := setupTestDB(t)
	storage := NewTeamPostgresStorage(pool)
	ctx := context.Background()

	team := models.Team{
		TeamName: "rollback_test",
		Members: []models.User{
			{UserID: "u1", Username: "TestUser", TeamName: "rollback_test", IsActive: true},
		},
	}

	tx, err := storage.TeamBeginTx(ctx)
	require.NoError(t, err)

	err = storage.CreateTeamTx(ctx, tx, team)
	require.NoError(t, err)


	err = tx.Rollback(ctx)
	require.NoError(t, err)

	tx, err = storage.TeamBeginTx(ctx)
	require.NoError(t, err)
	defer tx.Rollback(ctx)

	_, err = storage.GetTeamInfoTx(ctx, tx, "rollback_test")
	assert.ErrorIs(t, err, models.ErrNotFound)
}
