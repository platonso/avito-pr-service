-- +goose Up

CREATE INDEX idx_users_team_active
    ON users(team_name, is_active);

CREATE INDEX idx_pr_reviewers_reviewer_id
    ON pr_reviewers(reviewer_id);

CREATE INDEX idx_pr_reviewers_pr_id
    ON pr_reviewers(pr_id);

-- +goose Down

DROP INDEX IF EXISTS idx_pr_reviewers_pr_id;
DROP INDEX IF EXISTS idx_pr_reviewers_reviewer_id;
DROP INDEX IF EXISTS idx_users_team_active;
