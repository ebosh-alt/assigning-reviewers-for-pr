-- +goose Up
-- +goose StatementBegin
CREATE TABLE pr_reassignment_history (
    id SERIAL PRIMARY KEY,
    pr_id TEXT NOT NULL REFERENCES pull_requests(id) ON DELETE CASCADE,
    old_reviewer_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    new_reviewer_id TEXT REFERENCES users(id) ON DELETE RESTRICT,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_pr_reassignment_history_pr_id ON pr_reassignment_history(pr_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pr_reassignment_history;
-- +goose StatementEnd
