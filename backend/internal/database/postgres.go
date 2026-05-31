package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresPool creates a new connection pool to PostgreSQL
func NewPostgresPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	config.MaxConns = 20
	config.MinConns = 2
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	fmt.Println("✅ Connected to PostgreSQL")
	return pool, nil
}

// RunMigrations executes all SQL migration files
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migration := `
	-- Users table
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		full_name VARCHAR(255) NOT NULL DEFAULT '',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	-- Resumes table
	CREATE TABLE IF NOT EXISTS resumes (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		file_path VARCHAR(500),
		file_name VARCHAR(255),
		raw_text TEXT,
		parsed_data JSONB,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);

	-- Jobs table
	CREATE TABLE IF NOT EXISTS jobs (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		external_id VARCHAR(255),
		source VARCHAR(50) NOT NULL,
		title VARCHAR(500) NOT NULL,
		company VARCHAR(255),
		location VARCHAR(255),
		description TEXT,
		requirements TEXT,
		salary_range VARCHAR(100),
		job_type VARCHAR(50),
		experience_level VARCHAR(50),
		skills JSONB DEFAULT '[]'::jsonb,
		url VARCHAR(1000),
		posted_at TIMESTAMPTZ,
		scraped_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		is_active BOOLEAN NOT NULL DEFAULT TRUE,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(source, external_id)
	);

	-- Job matches table
	CREATE TABLE IF NOT EXISTS job_matches (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
		relevance_score FLOAT NOT NULL DEFAULT 0,
		matched_skills JSONB DEFAULT '[]'::jsonb,
		status VARCHAR(20) NOT NULL DEFAULT 'new',
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
		UNIQUE(user_id, job_id)
	);

	-- Indexes
	CREATE INDEX IF NOT EXISTS idx_jobs_source ON jobs(source);
	CREATE INDEX IF NOT EXISTS idx_jobs_skills ON jobs USING GIN(skills);
	CREATE INDEX IF NOT EXISTS idx_jobs_posted_at ON jobs(posted_at DESC);
	CREATE INDEX IF NOT EXISTS idx_jobs_is_active ON jobs(is_active);
	CREATE INDEX IF NOT EXISTS idx_job_matches_user ON job_matches(user_id, relevance_score DESC);
	CREATE INDEX IF NOT EXISTS idx_job_matches_status ON job_matches(user_id, status);
	CREATE INDEX IF NOT EXISTS idx_resumes_user ON resumes(user_id);
	`

	_, err := pool.Exec(ctx, migration)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	fmt.Println("✅ Database migrations completed")
	return nil
}
