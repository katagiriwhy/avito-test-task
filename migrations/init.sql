CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');

CREATE TABLE IF NOT EXISTS public.teams (
    team_name TEXT PRIMARY KEY
);

CREATE TABLE IF NOT EXISTS public.users (
       user_id TEXT PRIMARY KEY,
       username TEXT NOT NULL,
       team_name TEXT NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
       is_active BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);

CREATE TABLE IF NOT EXISTS pull_requests(
        pull_request_id TEXT PRIMARY KEY,
        pull_request_name TEXT NOT NULL,
        author_id TEXT NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
        status pr_status NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
        merged_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_pr_author ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pr_status ON pull_requests(status);

CREATE TABLE IF NOT EXISTS reviewers (
    pull_request_id TEXT NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    reviewer_id TEXT NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,

    PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE INDEX IF NOT EXISTS idx_reviewers_user ON reviewers(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_reviewers_pr ON reviewers(pull_request_id);
