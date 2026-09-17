-- Fail closed, not fail open: a persona must be explicitly marked
-- appropriate for minors, rather than defaulting to visible/usable and
-- requiring someone to remember to restrict it later.
ALTER TABLE personas
    ADD COLUMN is_minor_appropriate BOOLEAN NOT NULL DEFAULT false;

-- Every persona seeded so far was deliberately authored as a platonic,
-- non-romantic companion (see their system prompts in
-- 000003_seed_personas) — reviewed and safe for minors.
UPDATE personas SET is_minor_appropriate = true;

CREATE TABLE onboarding_profiles (
    user_id                UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    -- Wire values match the iOS OnboardingProfile.AgeRange enum's case
    -- names exactly (Swift's default Codable raw value).
    age_range              TEXT NOT NULL
        CHECK (age_range IN ('under18', 'age18to24', 'age25to34', 'age35plus')),
    -- Recomputed server-side from age_range on every write (see
    -- OnboardingService) — never trusted from the client as-is, since
    -- persona age-gating depends on it.
    is_minor               BOOLEAN NOT NULL,
    mood_preference        TEXT
        CHECK (mood_preference IS NULL OR mood_preference IN ('motivation', 'dailyChat', 'hobbyTalk', 'skipped')),
    notifications_granted  BOOLEAN NOT NULL,
    selected_persona_id    UUID REFERENCES personas(id),
    created_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER onboarding_profiles_set_updated_at
    BEFORE UPDATE ON onboarding_profiles
    FOR EACH ROW
    EXECUTE FUNCTION set_updated_at();
