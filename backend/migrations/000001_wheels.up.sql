-- SPI-7 step 1: Neon initial schema (PG 16).
-- Tables: wheels, wheel_items, spins, pageviews (events funnel).
-- pageviews doubles as the RecordEvent store: PAGEVIEW rows carry path,
-- SPIN_START/SPIN_END rows carry wheel_id/spin_id/sig. Idempotency key is
-- (session_id, client_ts, type, spin_id) per service.proto:50-52.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Wheels ---------------------------------------------------------------
CREATE TABLE IF NOT EXISTS wheels (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT        NOT NULL,
  config     JSONB       NOT NULL DEFAULT '{}'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Items ----------------------------------------------------------------
CREATE TABLE IF NOT EXISTS wheel_items (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  wheel_id   UUID        NOT NULL REFERENCES wheels (id) ON DELETE CASCADE,
  option     TEXT        NOT NULL,
  color      TEXT        NOT NULL DEFAULT '',
  weight     DOUBLE PRECISION NOT NULL DEFAULT 1.0 CHECK (weight > 0),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS wheel_items_wheel_id_idx ON wheel_items (wheel_id);

-- Spins (server decision audit log) ------------------------------------
CREATE TABLE IF NOT EXISTS spins (
  id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  wheel_id       UUID        NOT NULL REFERENCES wheels (id) ON DELETE CASCADE,
  winner_idx     INT         NOT NULL CHECK (winner_idx >= 0),
  winner_item_id UUID        REFERENCES wheel_items (id) ON DELETE SET NULL,
  seed_server    BYTEA       NOT NULL,
  seed_client    TEXT        NOT NULL DEFAULT '',
  items_hash     TEXT        NOT NULL DEFAULT '',
  nonce          TEXT        NOT NULL DEFAULT '',
  sig            TEXT        NOT NULL DEFAULT '',
  session_id     TEXT        NOT NULL DEFAULT '',
  ip_hash        TEXT        NOT NULL DEFAULT '',
  ua             TEXT        NOT NULL DEFAULT '',
  cf_ray         TEXT        NOT NULL DEFAULT '',
  expires_at     TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS spins_wheel_created_idx ON spins (wheel_id, created_at DESC);

-- Pageviews / events funnel --------------------------------------------
CREATE TABLE IF NOT EXISTS pageviews (
  id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  session_id TEXT        NOT NULL,
  type       TEXT        NOT NULL DEFAULT 'PAGEVIEW'
                        CHECK (type IN ('PAGEVIEW', 'SPIN_START', 'SPIN_END')),
  wheel_id   UUID        REFERENCES wheels (id) ON DELETE CASCADE,
  spin_id    UUID        REFERENCES spins (id) ON DELETE SET NULL,
  path       TEXT        NOT NULL DEFAULT '',
  sig        TEXT        NOT NULL DEFAULT '',
  client_ts  TIMESTAMPTZ NOT NULL DEFAULT now(),
  ip_hash    TEXT        NOT NULL DEFAULT '',
  ua         TEXT        NOT NULL DEFAULT '',
  cf_ray     TEXT        NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS pageviews_session_created_idx ON pageviews (session_id, created_at DESC);
CREATE INDEX IF NOT EXISTS pageviews_spin_idx ON pageviews (spin_id) WHERE spin_id IS NOT NULL;
-- Idempotency: (session_id, client_ts, type, spin_id). spin_id is nullable,
-- so coalesce it; repeated delivery of the same event is a no-op.
CREATE UNIQUE INDEX IF NOT EXISTS pageviews_idempotency_uniq
  ON pageviews (session_id, type, client_ts, (COALESCE(spin_id::text, '')));

-- updated_at trigger for wheels ----------------------------------------
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS wheels_updated_at ON wheels;
CREATE TRIGGER wheels_updated_at
  BEFORE UPDATE ON wheels
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
