-- +goose Up
-- +goose StatementBegin
CREATE TABLE teams (
                       id SERIAL PRIMARY KEY,
                       name TEXT NOT NULL UNIQUE
);

CREATE TABLE users (
                       id TEXT PRIMARY KEY,
                       username TEXT NOT NULL,
                       team_id INTEGER NOT NULL REFERENCES teams(id) ON DELETE RESTRICT,
                       is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE pull_requests (
                               id TEXT PRIMARY KEY,
                               name TEXT NOT NULL,
                               author_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
                               status TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               merged_at TIMESTAMPTZ
);

CREATE TABLE pr_reviewers (
                              pr_id TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
                              reviewer_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
                              PRIMARY KEY (pr_id, reviewer_id)
);

CREATE INDEX idx_users_team_id ON users(team_id);
CREATE INDEX idx_pr_reviewers_reviewer_id ON pr_reviewers(reviewer_id);
CREATE INDEX idx_pr_status ON pull_requests(status);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
-- +goose StatementEnd
