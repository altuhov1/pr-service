package migrations

import (
	"context"
	"database/sql"

	"github.com/pressly/goose/v3"
)

func init() {
	goose.AddMigrationContext(upCreateSubscriptionsTable, downCreateSubscriptionsTable)
}

func upCreateSubscriptionsTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS teams (
		name TEXT PRIMARY KEY
	);
	`)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS users (
		user_id TEXT PRIMARY KEY,
		username TEXT NOT NULL,
		team_name TEXT NOT NULL REFERENCES teams(name) ON DELETE CASCADE,
		is_active BOOLEAN NOT NULL DEFAULT true
	);
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
	CREATE TABLE IF NOT EXISTS pull_requests (
		pull_request_id TEXT PRIMARY KEY,
		pull_request_name TEXT NOT NULL,
		author_id TEXT NOT NULL REFERENCES users(user_id),
		status TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
		assigned_reviewers TEXT[] NOT NULL DEFAULT '{}',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		merged_at TIMESTAMPTZ
	);
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_users_team ON users(team_name);
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		CREATE INDEX IF NOT EXISTS idx_pull_requests_reviewers ON pull_requests USING GIN(assigned_reviewers);
	`)
	if err != nil {
		return err
	}
	return err
}

func downCreateSubscriptionsTable(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		DROP TABLE IF EXISTS teams CASCADE;
	`)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		DROP TABLE IF EXISTS users CASCADE;
	`)
	if err != nil {
		return err
	}
	return nil
}
