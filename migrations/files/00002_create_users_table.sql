-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.users(
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    display_name TEXT NOT NULL CHECK (char_length(display_name) > 0),
    username TEXT UNIQUE,
    metadata JSONB DEFAULT NULL,
    email TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT NULL,
    email_verified_at TIMESTAMPTZ DEFAULT NULL,
    last_login_at TIMESTAMPTZ DEFAULT NULL,
    CONSTRAINT chk_email_format CHECK (char_length(email) > 3 AND email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$'),
    -- Username only allows alphanumeric characters and underscores, must be between 3 and 32 characters long
    CONSTRAINT chk_username_format CHECK (username IS NULL OR username ~ '^[a-zA-Z0-9_]{3,32}$')
);

-- Constraints
CREATE INDEX IF NOT EXISTS idx_users_display_name ON public.users USING gin (display_name gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_users_username ON public.users USING GIN (username gin_trgm_ops) WHERE username IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_email ON public.users (LOWER(email));
CREATE UNIQUE INDEX idx_users_normalized_username ON public.users (LOWER(username));
CREATE UNIQUE INDEX idx_users_normalized_email ON public.users (LOWER(email));
CREATE INDEX IF NOT EXISTS idx_users_created_at ON public.users (created_at);
CREATE INDEX IF NOT EXISTS idx_users_updated_at ON public.users (updated_at) WHERE updated_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_metadata_gin ON public.users USING GIN (metadata);
CREATE INDEX IF NOT EXISTS idx_users_last_login_at ON public.users (last_login_at) WHERE last_login_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_users_email_verified_at ON public.users (email_verified_at) WHERE email_verified_at IS NOT NULL;
CREATE TRIGGER trg_users_updated_at BEFORE UPDATE ON public.users FOR EACH ROW EXECUTE FUNCTION fn_updated_at_value();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_users_updated_at ON public.users;
DROP INDEX IF EXISTS idx_users_last_login_at;
DROP INDEX IF EXISTS idx_users_normalized_email;
DROP INDEX IF EXISTS idx_users_normalized_username;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_metadata_gin;
DROP INDEX IF EXISTS idx_users_updated_at;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_display_name;
DROP TABLE IF EXISTS public.users;
-- +goose StatementEnd
