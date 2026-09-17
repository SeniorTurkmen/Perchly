DROP TABLE IF EXISTS credit_balances;
DROP TABLE IF EXISTS user_quotas;

DROP INDEX IF EXISTS conversations_user_id_idx;
ALTER TABLE conversations DROP COLUMN IF EXISTS user_id;

DROP TABLE IF EXISTS users;
