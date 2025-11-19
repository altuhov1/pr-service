set -e

until pg_isready -U "$POSTGRES_USER" -d "$POSTGRES_DB"; do
  sleep 2
done


psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
    CREATE TABLE IF NOT EXISTS users (
        user_id TEXT PRIMARY KEY,
        is_active BOOLEAN NOT NULL DEFAULT false
    );

    CREATE TABLE IF NOT EXISTS teams (
        name TEXT PRIMARY KEY,
        user_ids TEXT[] NOT NULL DEFAULT '{}'
    );

    CREATE TABLE IF NOT EXISTS pull_requests (
        id SERIAL PRIMARY KEY,
        status VARCHAR(50) NOT NULL,
        inspector_ids TEXT[] NOT NULL DEFAULT '{}',
        is_merged BOOLEAN NOT NULL DEFAULT false,
        merged_at TIMESTAMPTZ,
        created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
    );

    CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
    CREATE INDEX IF NOT EXISTS idx_pull_requests_status ON pull_requests(status);
    CREATE INDEX IF NOT EXISTS idx_pull_requests_is_merged ON pull_requests(is_merged);
    CREATE INDEX IF NOT EXISTS idx_pull_requests_created_at ON pull_requests(created_at);

    GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO "$POSTGRES_USER";
    GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO "$POSTGRES_USER";
EOSQL
