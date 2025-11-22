set -e

until pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"; do
  sleep 2
done

psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE TABLE IF NOT EXISTS teams (
        name TEXT PRIMARY KEY
    );
    
    CREATE TABLE IF NOT EXISTS users (
        user_id TEXT PRIMARY KEY,
        username TEXT NOT NULL,
        team_name TEXT NOT NULL REFERENCES teams(name) ON DELETE CASCADE,
        is_active BOOLEAN NOT NULL DEFAULT true
    );


    CREATE TABLE IF NOT EXISTS pull_requests (
        pull_request_id TEXT PRIMARY KEY,
        pull_request_name TEXT NOT NULL,
        author_id TEXT NOT NULL REFERENCES users(user_id),
        status TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
        assigned_reviewers TEXT[] NOT NULL DEFAULT '{}',
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        merged_at TIMESTAMPTZ
    );

    CREATE INDEX IF NOT EXISTS idx_users_team ON users(team_name);
    CREATE INDEX IF NOT EXISTS idx_pull_requests_reviewers ON pull_requests USING GIN(assigned_reviewers);

    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "$POSTGRES_USER";
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO "$POSTGRES_USER";
EOSQL