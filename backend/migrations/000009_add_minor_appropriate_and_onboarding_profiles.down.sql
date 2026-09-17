DROP TRIGGER IF EXISTS onboarding_profiles_set_updated_at ON onboarding_profiles;
DROP TABLE IF EXISTS onboarding_profiles;

ALTER TABLE personas DROP COLUMN IF EXISTS is_minor_appropriate;
