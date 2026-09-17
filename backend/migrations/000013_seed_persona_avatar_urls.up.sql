-- Placeholder avatars (DiceBear, a public HTTPS avatar-generation
-- service) until real illustrated artwork exists for Ada/Mira/Kerem.
-- iOS already falls back to an initial letter when avatar_url is null,
-- so this only needs to be swapped for real asset URLs later — no
-- client change required either way.
UPDATE personas SET avatar_url = 'https://api.dicebear.com/9.x/notionists-neutral/png?seed=ada&backgroundColor=FF6B35'
    WHERE slug = 'motivational-coach';
UPDATE personas SET avatar_url = 'https://api.dicebear.com/9.x/notionists-neutral/png?seed=mira&backgroundColor=4A90D9'
    WHERE slug = 'daily-companion';
UPDATE personas SET avatar_url = 'https://api.dicebear.com/9.x/notionists-neutral/png?seed=kerem&backgroundColor=6FCF97'
    WHERE slug = 'hobby-book-partner';
