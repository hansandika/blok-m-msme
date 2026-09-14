-- Track when a moderator applies an approved suggestion onto live places.
-- Status stays pending/approved/rejected; applied_at is the apply latch (idempotent).

ALTER TABLE suggestions
  ADD COLUMN IF NOT EXISTS applied_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS applied_place_id UUID REFERENCES places(id) ON DELETE SET NULL;
