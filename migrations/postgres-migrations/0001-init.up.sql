CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE tasks(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    status BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);
CREATE INDEX idx_tasks_user_id on tasks(user_id);

CREATE TABLE task_analytics(
    user_id UUID PRIMARY KEYREFERENCES users(id) ON DELETE CASCADE,
    tasks_created BIGINT DEFAULT 0,
    tasks_completed BIGINT DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);