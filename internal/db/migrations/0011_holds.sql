-- Hold / pin timers on live fights.
ALTER TABLE fight_state ADD COLUMN hold_seconds INTEGER NOT NULL DEFAULT 0;
ALTER TABLE fight_state ADD COLUMN hold_corner TEXT NOT NULL DEFAULT '';
