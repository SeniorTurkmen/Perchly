DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS email_verification_codes;

DROP TRIGGER IF EXISTS users_set_updated_at ON users;

ALTER TABLE users
    DROP COLUMN IF EXISTS updated_at,
    DROP COLUMN IF EXISTS device_id,
    DROP COLUMN IF EXISTS is_anonymous,
    DROP COLUMN IF EXISTS display_name,
    DROP COLUMN IF EXISTS email_verified_at,
    DROP COLUMN IF EXISTS email;
