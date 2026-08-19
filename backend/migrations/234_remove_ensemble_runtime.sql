-- Remove the retired TokenPort Ensemble persistence after DSH became the
-- independently deployed coordinator. Both statements are idempotent so a
-- partially migrated environment can restart safely.
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM groups WHERE platform = 'ensemble') THEN
        RAISE EXCEPTION 'cannot remove Ensemble persistence while platform=ensemble groups remain; migrate them to a concrete or composite platform first';
    END IF;
END $$;

DROP TABLE IF EXISTS ensemble_proposers;
ALTER TABLE groups DROP COLUMN IF EXISTS ensemble_config;
