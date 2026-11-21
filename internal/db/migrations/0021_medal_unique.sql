CREATE UNIQUE INDEX IF NOT EXISTS medals_unique_per_fighter_kind
  ON medals(event_id, IFNULL(category_id, ''), fighter_id, kind);
