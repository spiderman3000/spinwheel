-- SPI-7 step 1: deterministic item ordering.
-- Wheel items are addressed by index (winner_idx), so insertion order must
-- survive round-trips. Same-transaction inserts share created_at, so
-- ORDER BY created_at cannot restore order. position is the source of truth.

ALTER TABLE wheel_items
  ADD COLUMN IF NOT EXISTS position INT NOT NULL DEFAULT 0;

-- Backfill existing rows in current (created_at, id) order.
WITH ranked AS (
  SELECT id,
         ROW_NUMBER() OVER (PARTITION BY wheel_id ORDER BY created_at, id) - 1 AS rn
  FROM wheel_items
)
UPDATE wheel_items w SET position = r.rn FROM ranked r WHERE w.id = r.id;

CREATE INDEX IF NOT EXISTS wheel_items_wheel_position_idx
  ON wheel_items (wheel_id, position);
