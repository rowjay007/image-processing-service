CREATE TABLE IF NOT EXISTS variants (
    id UUID PRIMARY KEY,
    image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    variant_key TEXT NOT NULL,
    spec_hash TEXT NOT NULL,
    size BIGINT NOT NULL,
    mime_type TEXT NOT NULL,
    width INTEGER,
    height INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_variants_image_id ON variants(image_id);
-- Ensure we don't process the same variant spec twice for the same image
CREATE UNIQUE INDEX IF NOT EXISTS idx_variants_image_id_spec_hash ON variants(image_id, spec_hash);
