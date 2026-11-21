-- Per-discipline competition format and medal policy.

CREATE TABLE IF NOT EXISTS discipline_formats (
  event_id       TEXT NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  discipline_id  TEXT NOT NULL,
  format         TEXT NOT NULL DEFAULT 'round_robin',
  knockout       TEXT NOT NULL DEFAULT 'none',
  third_place    TEXT NOT NULL DEFAULT 'rematch',
  dual_bronze    INTEGER NOT NULL DEFAULT 0,
  retry_ties     INTEGER NOT NULL DEFAULT 0,
  order_policy   TEXT NOT NULL DEFAULT 'age_then_rank',
  PRIMARY KEY (event_id, discipline_id)
);
