-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS public.notes (
    id UUID NOT NULL PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES public.users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    content JSONB NOT NULL,
    content_text TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT NULL
);

CREATE INDEX IF NOT EXISTS idx_notes_title_trgm ON public.notes USING GIN (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_notes_content_text_trgm ON public.notes USING GIN (content_text gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_notes_created_at ON public.notes (created_at);
CREATE TRIGGER trg_notes_updated_at BEFORE UPDATE ON public.notes FOR EACH ROW EXECUTE FUNCTION fn_updated_at_value();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trg_notes_updated_at ON public.notes;
DROP INDEX IF EXISTS idx_notes_created_at;
DROP INDEX IF EXISTS idx_notes_title_trgm;
DROP INDEX IF EXISTS idx_notes_content_text_trgm;

DROP TABLE IF EXISTS public.notes;
-- +goose StatementEnd
