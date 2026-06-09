DROP TRIGGER IF EXISTS set_timestamp_recoveries ON "recoveries";
DROP TRIGGER IF EXISTS set_timestamp_users ON "users";

DROP INDEX IF EXISTS idx_recoveries_code;
DROP INDEX IF EXISTS idx_recoveries_email;

DROP TABLE IF EXISTS "recoveries";
DROP TABLE IF EXISTS "sessions";

DROP TABLE IF EXISTS "users";

DROP FUNCTION IF EXISTS trigger_set_timestamp();