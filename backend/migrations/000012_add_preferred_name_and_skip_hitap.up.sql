ALTER TABLE onboarding_profiles
    ADD COLUMN preferred_name TEXT NULL,
    ADD COLUMN skip_hitap BOOLEAN NOT NULL DEFAULT false;
