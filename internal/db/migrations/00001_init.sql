-- +goose Up

-- Create teams table
CREATE TABLE IF NOT EXISTS teams (
    team_name TEXT PRIMARY KEY
);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
     user_id TEXT PRIMARY KEY,
     username TEXT NOT NULL,
     team_name TEXT NOT NULL REFERENCES teams(team_name),
     is_active BOOLEAN NOT NULL DEFAULT TRUE
);

-- Create pull_requests table
CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id TEXT PRIMARY KEY,
    pull_request_name TEXT NOT NULL,
    author_id TEXT NOT NULL REFERENCES users(user_id),
    status TEXT NOT NULL DEFAULT 'OPEN' CHECK(status IN ('OPEN', 'MERGED')),
    created_at TIMESTAMPTZ,
    merged_at TIMESTAMPTZ
);

-- Create pr_reviewers table
CREATE TABLE IF NOT EXISTS pr_reviewers (
    pr_id TEXT NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id TEXT NOT NULL REFERENCES users(user_id),
    PRIMARY KEY (pr_id, reviewer_id)
);

-- +goose Down

DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;