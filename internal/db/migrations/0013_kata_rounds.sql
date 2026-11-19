-- Multi-round kata: shitei (compulsory) / tokuyi (choice) and bunkai scores.
ALTER TABLE kata_performances ADD COLUMN round_number INTEGER NOT NULL DEFAULT 1;
ALTER TABLE kata_performances ADD COLUMN kata_type TEXT NOT NULL DEFAULT 'tokuyi';
ALTER TABLE kata_performances ADD COLUMN bunkai_score REAL NOT NULL DEFAULT 0;
