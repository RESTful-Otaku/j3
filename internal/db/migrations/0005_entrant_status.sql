-- On-the-day dropouts, no-shows and re-entry.
ALTER TABLE entrants ADD COLUMN status TEXT NOT NULL DEFAULT 'active';
