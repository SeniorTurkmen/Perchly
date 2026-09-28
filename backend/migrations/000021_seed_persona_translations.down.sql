DELETE FROM persona_translations
WHERE locale IN ('en', 'de', 'es', 'fr', 'ru', 'zh', 'ar')
AND persona_id IN (SELECT id FROM personas WHERE slug IN ('motivational-coach', 'daily-companion', 'hobby-book-partner'));
