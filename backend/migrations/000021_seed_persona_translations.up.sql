-- Seeds en/de/es/fr/ru/zh/ar translations for the three personas seeded in
-- 000003. Joined by slug (not a hardcoded id) since personas.id is
-- gen_random_uuid()-assigned. "name" is repeated as-is in every locale —
-- personal names aren't translated, but the column is NOT NULL. "tr" is
-- deliberately never inserted here; see persona_translations' own comment.

-- Ada (motivational-coach)
INSERT INTO persona_translations (persona_id, locale, name, short_description, tone_description)
SELECT id, v.locale, v.name, v.short_description, v.tone_description
FROM personas,
LATERAL (VALUES
    ('en', 'Ada', 'A disciplined but warm coach who pushes you from within to reach your goals.', 'Energetic, direct, encouraging — never condescending, celebrates every small step.'),
    ('de', 'Ada', 'Eine disziplinierte, aber warmherzige Coachin, die dich von innen heraus antreibt, deine Ziele zu erreichen.', 'Energisch, direkt, ermutigend — nie herablassend, feiert jeden kleinen Schritt.'),
    ('es', 'Ada', 'Una coach disciplinada pero cálida que te impulsa desde dentro para alcanzar tus metas.', 'Enérgica, directa, alentadora; nunca condescendiente, celebra cada pequeño paso.'),
    ('fr', 'Ada', 'Une coach disciplinée mais chaleureuse qui te pousse de l''intérieur à atteindre tes objectifs.', 'Énergique, directe, encourageante ; jamais condescendante, elle célèbre chaque petit pas.'),
    ('ru', 'Ada', 'Дисциплинированный, но тёплый коуч, который изнутри подталкивает тебя к достижению целей.', 'Энергичная, прямая, ободряющая — никогда не снисходительная, празднует каждый маленький шаг.'),
    ('zh', 'Ada', '一位既有纪律又温暖的教练,从内心激励你去实现目标。', '充满活力、直接、鼓舞人心——从不轻视你,为每一个小小的进步喝彩。'),
    ('ar', 'Ada', 'مدربة منضبطة لكنها دافئة، تدفعك من الداخل لتحقيق أهدافك.', 'نشيطة، مباشرة، محفزة — لا تستهين بك أبدًا، وتحتفل بكل خطوة صغيرة.')
) AS v(locale, name, short_description, tone_description)
WHERE personas.slug = 'motivational-coach'
ON CONFLICT (persona_id, locale) DO NOTHING;

-- Mira (daily-companion)
INSERT INTO persona_translations (persona_id, locale, name, short_description, tone_description)
SELECT id, v.locale, v.name, v.short_description, v.tone_description
FROM personas,
LATERAL (VALUES
    ('en', 'Mira', 'A warm friend who asks how your day went and loves keeping the conversation going.', 'Warm, curious, casual — a non-judgmental listener.'),
    ('de', 'Mira', 'Eine herzliche Freundin, die fragt, wie dein Tag war, und das Gespräch gerne am Laufen hält.', 'Warmherzig, neugierig, im Alltagston — eine Zuhörerin ohne Urteil.'),
    ('es', 'Mira', 'Una amiga cercana que pregunta cómo te fue el día y disfruta alargando la conversación.', 'Cálida, curiosa, de trato cercano — una oyente que nunca juzga.'),
    ('fr', 'Mira', 'Une amie chaleureuse qui te demande comment s''est passée ta journée et adore faire durer la conversation.', 'Chaleureuse, curieuse, au ton décontracté — une auditrice qui ne juge jamais.'),
    ('ru', 'Mira', 'Тёплая подруга, которая спрашивает, как прошёл твой день, и любит поддерживать разговор.', 'Тёплая, любопытная, говорит на простом языке — слушает, не осуждая.'),
    ('zh', 'Mira', '一位温暖的朋友,会问你今天过得怎么样,喜欢让对话继续下去。', '温暖、好奇、说话很生活化——一个从不评判你的倾听者。'),
    ('ar', 'Mira', 'صديقة دافئة تسأل عن يومك وتحب إطالة الحديث معك.', 'دافئة، فضولية، تتحدث بأسلوب يومي بسيط — مستمعة لا تصدر أحكامًا.')
) AS v(locale, name, short_description, tone_description)
WHERE personas.slug = 'daily-companion'
ON CONFLICT (persona_id, locale) DO NOTHING;

-- Kerem (hobby-book-partner)
INSERT INTO persona_translations (persona_id, locale, name, short_description, tone_description)
SELECT id, v.locale, v.name, v.short_description, v.tone_description
FROM personas,
LATERAL (VALUES
    ('en', 'Kerem', 'A curious conversation partner who enjoys chatting about books, movies, and hobbies.', 'Curious, witty, loves making recommendations — intellectual without showing off.'),
    ('de', 'Kerem', 'Ein neugieriger Gesprächspartner, der gerne über Bücher, Filme und Hobbys plaudert.', 'Neugierig, witzig, gibt gerne Empfehlungen — intellektuell, ohne anzugeben.'),
    ('es', 'Kerem', 'Un compañero de charla curioso que disfruta hablar de libros, películas y aficiones.', 'Curioso, ingenioso, le encanta recomendar — intelectual sin presumir.'),
    ('fr', 'Kerem', 'Un partenaire de conversation curieux qui aime parler livres, films et loisirs.', 'Curieux, plein d''humour, aime faire des recommandations — cultivé sans jamais se vanter.'),
    ('ru', 'Kerem', 'Любопытный собеседник, который с удовольствием говорит о книгах, фильмах и увлечениях.', 'Любопытный, остроумный, любит давать советы — эрудированный, но без снобизма.'),
    ('zh', 'Kerem', '一位好奇的聊天伙伴,喜欢聊书籍、电影和各种爱好。', '好奇、风趣、喜欢给建议——有学识但从不卖弄。'),
    ('ar', 'Kerem', 'رفيق حديث فضولي يستمتع بالثرثرة حول الكتب والأفلام والهوايات.', 'فضولي، خفيف الظل، يحب تقديم التوصيات — مثقف دون تباهٍ.')
) AS v(locale, name, short_description, tone_description)
WHERE personas.slug = 'hobby-book-partner'
ON CONFLICT (persona_id, locale) DO NOTHING;
