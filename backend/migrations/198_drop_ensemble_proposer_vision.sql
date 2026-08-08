-- Drop two pieces of ensemble_proposers schema that no code reads.
--
-- 1. The per-member `vision` flag was rolled back. The fan-out now hands every
--    member the caller's original body (images included) and gives the
--    aggregator the raw body too, so nothing consults the column; the ent schema
--    no longer declares the field either. Left in place it is live-database-only
--    state that every future schema diff has to re-explain.
--
--    Its original migration file (198_ensemble_proposer_vision.sql) was removed
--    together with the rollback, but its schema_migrations row survives and can
--    never be replayed. Applied migrations are keyed by filename, so this file
--    runs regardless of that dangling record.
--
-- 2. idx_ensemble_proposers_group_role is a strict column prefix of
--    idx_ensemble_proposers_unique_active (group_id, role, model), under the
--    identical `deleted_at IS NULL` predicate. Any lookup it could serve is
--    already served by that unique index, so it only costs write amplification.
ALTER TABLE ensemble_proposers
    DROP COLUMN IF EXISTS vision;

DROP INDEX IF EXISTS idx_ensemble_proposers_group_role;
