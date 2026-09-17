ALTER TABLE messages
    ADD COLUMN reaction_emoji TEXT NULL
        CHECK (reaction_emoji IS NULL OR reaction_emoji IN ('❤️', '😂', '👍', '👎', '‼️', '❓'));
