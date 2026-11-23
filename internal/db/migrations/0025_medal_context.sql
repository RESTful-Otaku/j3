-- Medals are awarded per category AND per discipline/event type. The old
-- per-fighter/kind index made that impossible, so replace it with stable award
-- slots that can be recomputed deterministically.

ALTER TABLE medals ADD COLUMN discipline_id TEXT;
ALTER TABLE medals ADD COLUMN event_type TEXT NOT NULL DEFAULT 'fight';
ALTER TABLE medals ADD COLUMN award_key TEXT;
ALTER TABLE medals ADD COLUMN source TEXT NOT NULL DEFAULT 'manual';

ALTER TABLE fights ADD COLUMN discipline_id TEXT;
ALTER TABLE kata_performances ADD COLUMN discipline_id TEXT;
ALTER TABLE kata_performances ADD COLUMN category_id TEXT;

DROP INDEX IF EXISTS medals_unique_per_fighter_kind;
CREATE UNIQUE INDEX IF NOT EXISTS medals_award_slot
  ON medals(event_id, award_key)
  WHERE award_key IS NOT NULL AND award_key <> '';
CREATE INDEX IF NOT EXISTS medals_event_context ON medals(event_id, event_type, discipline_id, category_id);

UPDATE fights SET discipline_id = COALESCE(discipline_id,
  CASE
    WHEN lower(round_label) LIKE '%randori%' OR lower(round_label) LIKE '%katame%'
      THEN (SELECT id FROM discipline_rules dr WHERE dr.event_id = fights.event_id AND lower(dr.name) LIKE '%randori%' LIMIT 1)
    WHEN lower(round_label) LIKE '%ground%' OR lower(round_label) LIKE '%ne-waza%'
      THEN (SELECT id FROM discipline_rules dr WHERE dr.event_id = fights.event_id AND lower(dr.name) LIKE '%ground%' LIMIT 1)
    WHEN lower(round_label) LIKE '%continuous%'
      THEN (SELECT id FROM discipline_rules dr WHERE dr.event_id = fights.event_id AND lower(dr.name) LIKE '%continuous%' LIMIT 1)
    WHEN lower(round_label) LIKE '%spar%'
      THEN (SELECT id FROM discipline_rules dr WHERE dr.event_id = fights.event_id AND lower(dr.name) LIKE '%spar%' LIMIT 1)
    ELSE (SELECT id FROM discipline_rules dr WHERE dr.event_id = fights.event_id AND dr.kind = 'fight' ORDER BY rowid LIMIT 1)
  END)
WHERE discipline_id IS NULL OR discipline_id = '';

UPDATE kata_performances SET discipline_id = COALESCE(discipline_id,
  CASE
    WHEN lower(kata_type) LIKE '%weapon%' OR round_number > 1
      THEN (SELECT id FROM discipline_rules dr WHERE dr.event_id = kata_performances.event_id AND lower(dr.name) LIKE '%weapon%' LIMIT 1)
    ELSE (SELECT id FROM discipline_rules dr WHERE dr.event_id = kata_performances.event_id AND lower(dr.name) LIKE '%kata%' LIMIT 1)
  END)
WHERE discipline_id IS NULL OR discipline_id = '';

UPDATE kata_performances SET category_id = COALESCE(category_id,
  (SELECT e.category_id FROM entrants e
   WHERE e.event_id = kata_performances.event_id AND e.fighter_id = kata_performances.fighter_id
     AND e.category_id IS NOT NULL AND e.category_id <> '' LIMIT 1),
  (SELECT id FROM categories WHERE event_id = kata_performances.event_id ORDER BY sort_order, id LIMIT 1))
WHERE category_id IS NULL OR category_id = '';
