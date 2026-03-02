-- =============================================================================
-- SHIBA Demo Seed Data
-- Safe to run multiple times (idempotent)
-- =============================================================================
-- Credentials:
--   admin@shiba.test      / Admin1234!
--   agency1@shiba.test    / Agency1234!   (Elite Models Agency)
--   agency2@shiba.test    / Agency1234!   (Stars Talent Management)
--   agency3@shiba.test    / Agency1234!   (Moscow Fashion Group)
--   model1@shiba.test     / Model1234!    (Анна Смирнова)
--   model2@shiba.test     / Model1234!    (Мария Иванова)
--   model3@shiba.test     / Model1234!    (Екатерина Козлова)
--   model4@shiba.test     / Model1234!    (Ольга Петрова)
--   model5@shiba.test     / Model1234!    (Дарья Новикова)
--   model6@shiba.test     / Model1234!    (Полина Морозова)
-- =============================================================================

DO $$
DECLARE
  v_admin_id     UUID;
  -- agencies
  v_ag1_user     UUID; v_ag1_prof UUID;
  v_ag2_user     UUID; v_ag2_prof UUID;
  v_ag3_user     UUID; v_ag3_prof UUID;
  -- models
  v_m1_user      UUID; v_m1_prof  UUID;
  v_m2_user      UUID; v_m2_prof  UUID;
  v_m3_user      UUID; v_m3_prof  UUID;
  v_m4_user      UUID; v_m4_prof  UUID;
  v_m5_user      UUID; v_m5_prof  UUID;
  v_m6_user      UUID; v_m6_prof  UUID;
  -- castings
  v_c1 UUID; v_c2 UUID; v_c3 UUID; v_c4 UUID;
  v_c5 UUID; v_c6 UUID; v_c7 UUID; v_c8 UUID;

BEGIN

-- ─────────────────────────────────────────────────────────────────────────────
-- ADMIN
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO users (email, password_hash, role, status)
VALUES ('admin@shiba.test', crypt('Admin1234!', gen_salt('bf', 10)), 'admin', 'active')
ON CONFLICT (email) DO NOTHING
RETURNING id INTO v_admin_id;

IF v_admin_id IS NULL THEN
  SELECT id INTO v_admin_id FROM users WHERE email = 'admin@shiba.test';
END IF;

-- ─────────────────────────────────────────────────────────────────────────────
-- AGENCIES
-- ─────────────────────────────────────────────────────────────────────────────

-- Agency 1: Elite Models Agency
INSERT INTO users (email, password_hash, role, status)
VALUES ('agency1@shiba.test', crypt('Agency1234!', gen_salt('bf', 10)), 'agency', 'active')
ON CONFLICT (email) DO NOTHING
RETURNING id INTO v_ag1_user;
IF v_ag1_user IS NULL THEN
  SELECT id INTO v_ag1_user FROM users WHERE email = 'agency1@shiba.test';
END IF;

INSERT INTO agency_profiles (user_id, company_name, description, phone, status, approved_by, approved_at)
VALUES (
  v_ag1_user,
  'Elite Models Agency',
  'Ведущее агентство по подбору моделей для fashion и commercial съёмок. Работаем с ведущими брендами России и зарубежья с 2010 года. Наши модели появлялись на обложках Vogue, Elle и Harper''s Bazaar.',
  '+7 495 123-45-67',
  'active', v_admin_id, NOW()
)
ON CONFLICT (user_id) DO NOTHING
RETURNING id INTO v_ag1_prof;
IF v_ag1_prof IS NULL THEN
  SELECT id INTO v_ag1_prof FROM agency_profiles WHERE user_id = v_ag1_user;
END IF;

-- Agency 2: Stars Talent Management
INSERT INTO users (email, password_hash, role, status)
VALUES ('agency2@shiba.test', crypt('Agency1234!', gen_salt('bf', 10)), 'agency', 'active')
ON CONFLICT (email) DO NOTHING
RETURNING id INTO v_ag2_user;
IF v_ag2_user IS NULL THEN
  SELECT id INTO v_ag2_user FROM users WHERE email = 'agency2@shiba.test';
END IF;

INSERT INTO agency_profiles (user_id, company_name, description, phone, status, approved_by, approved_at)
VALUES (
  v_ag2_user,
  'Stars Talent Management',
  'Специализируемся на фитнес-моделях, спортивных брендах и рекламных кампаниях для активного образа жизни. Сотрудничаем с Nike, Adidas, Reebok и другими спортивными брендами.',
  '+7 812 987-65-43',
  'active', v_admin_id, NOW()
)
ON CONFLICT (user_id) DO NOTHING
RETURNING id INTO v_ag2_prof;
IF v_ag2_prof IS NULL THEN
  SELECT id INTO v_ag2_prof FROM agency_profiles WHERE user_id = v_ag2_user;
END IF;

-- Agency 3: Moscow Fashion Group
INSERT INTO users (email, password_hash, role, status)
VALUES ('agency3@shiba.test', crypt('Agency1234!', gen_salt('bf', 10)), 'agency', 'active')
ON CONFLICT (email) DO NOTHING
RETURNING id INTO v_ag3_user;
IF v_ag3_user IS NULL THEN
  SELECT id INTO v_ag3_user FROM users WHERE email = 'agency3@shiba.test';
END IF;

INSERT INTO agency_profiles (user_id, company_name, description, phone, status, approved_by, approved_at)
VALUES (
  v_ag3_user,
  'Moscow Fashion Group',
  'Крупнейшее московское агентство с портфолио из 200+ моделей. Организуем показы, кастинги и съёмки для российских и международных дизайнеров. Партнёры Mercedes-Benz Fashion Week Russia.',
  '+7 495 777-88-99',
  'active', v_admin_id, NOW()
)
ON CONFLICT (user_id) DO NOTHING
RETURNING id INTO v_ag3_prof;
IF v_ag3_prof IS NULL THEN
  SELECT id INTO v_ag3_prof FROM agency_profiles WHERE user_id = v_ag3_user;
END IF;

-- ─────────────────────────────────────────────────────────────────────────────
-- MODELS
-- ─────────────────────────────────────────────────────────────────────────────

-- Model 1: Анна Смирнова (fashion/editorial)
INSERT INTO users (email, password_hash, role, status)
VALUES ('model1@shiba.test', crypt('Model1234!', gen_salt('bf', 10)), 'model', 'active')
ON CONFLICT (email) DO NOTHING
RETURNING id INTO v_m1_user;
IF v_m1_user IS NULL THEN
  SELECT id INTO v_m1_user FROM users WHERE email = 'model1@shiba.test';
END IF;

INSERT INTO model_profiles (
  user_id, first_name, last_name, birth_date,
  city, country, willing_to_relocate,
  height_cm, weight_kg, chest_cm, waist_cm, hips_cm,
  shoe_size, clothing_size, phone, bio
)
VALUES (
  v_m1_user, 'Анна', 'Смирнова', '1998-05-15',
  'Москва', 'Россия', true,
  175, 55, 86, 60, 90,
  37.0, 'XS/S', '+7 916 234-56-78',
  'Профессиональная fashion-модель с 5-летним опытом. Работала с брендами Zara, H&M, IKEA. Специализируюсь на editorial и рекламных съёмках. Имею опыт работы на подиуме.'
)
ON CONFLICT (user_id) DO NOTHING
RETURNING id INTO v_m1_prof;
IF v_m1_prof IS NULL THEN
  SELECT id INTO v_m1_prof FROM model_profiles WHERE user_id = v_m1_user;
END IF;

INSERT INTO model_categories (model_profile_id, category_id)
SELECT v_m1_prof, id FROM categories WHERE name IN ('Fashion', 'Editorial', 'Beauty')
ON CONFLICT DO NOTHING;

-- Model 2: Мария Иванова (commercial/beauty)
INSERT INTO users (email, password_hash, role, status)
VALUES ('model2@shiba.test', crypt('Model1234!', gen_salt('bf', 10)), 'model', 'active')
ON CONFLICT (email) DO NOTHING
RETURNING id INTO v_m2_user;
IF v_m2_user IS NULL THEN
  SELECT id INTO v_m2_user FROM users WHERE email = 'model2@shiba.test';
END IF;

INSERT INTO model_profiles (
  user_id, first_name, last_name, birth_date,
  city, country, willing_to_relocate,
  height_cm, weight_kg, chest_cm, waist_cm, hips_cm,
  shoe_size, clothing_size, phone, bio
)
VALUES (
  v_m2_user, 'Мария', 'Иванова', '2000-11-22',
  'Санкт-Петербург', 'Россия', true,
  170, 52, 84, 62, 88,
  36.5, 'XS', '+7 921 345-67-89',
  'Коммерческая модель и beauty-специалист. Опыт в рекламных кампаниях косметических брендов. Работала с L''Oréal, Maybelline, Faberlic. Знаю три языка: русский, английский, французский.'
)
ON CONFLICT (user_id) DO NOTHING
RETURNING id INTO v_m2_prof;
IF v_m2_prof IS NULL THEN
  SELECT id INTO v_m2_prof FROM model_profiles WHERE user_id = v_m2_user;
END IF;

INSERT INTO model_categories (model_profile_id, category_id)
SELECT v_m2_prof, id FROM categories WHERE name IN ('Commercial', 'Beauty', 'Fashion')
ON CONFLICT DO NOTHING;

-- Model 3: Екатерина Козлова (fitness)
INSERT INTO users (email, password_hash, role, status)
VALUES ('model3@shiba.test', crypt('Model1234!', gen_salt('bf', 10)), 'model', 'active')
ON CONFLICT (email) DO NOTHING
RETURNING id INTO v_m3_user;
IF v_m3_user IS NULL THEN
  SELECT id INTO v_m3_user FROM users WHERE email = 'model3@shiba.test';
END IF;

INSERT INTO model_profiles (
  user_id, first_name, last_name, birth_date,
  city, country, willing_to_relocate,
  height_cm, weight_kg, chest_cm, waist_cm, hips_cm,
  shoe_size, clothing_size, phone, bio
)
VALUES (
  v_m3_user, 'Екатерина', 'Козлова', '1996-08-03',
  'Екатеринбург', 'Россия', true,
  172, 58, 88, 64, 92,
  37.5, 'S', '+7 922 456-78-90',
  'Фитнес-модель и сертифицированный персональный тренер. КМС по лёгкой атлетике. Сотрудничала со спортивными брендами Reebok, Under Armour, 2XU. Веду собственный фитнес-блог с аудиторией 150K.'
)
ON CONFLICT (user_id) DO NOTHING
RETURNING id INTO v_m3_prof;
IF v_m3_prof IS NULL THEN
  SELECT id INTO v_m3_prof FROM model_profiles WHERE user_id = v_m3_user;
END IF;

INSERT INTO model_categories (model_profile_id, category_id)
SELECT v_m3_prof, id FROM categories WHERE name IN ('Fitness', 'Commercial', 'Swimwear')
ON CONFLICT DO NOTHING;

-- Model 4: Ольга Петрова (plus size)
INSERT INTO users (email, password_hash, role, status)
VALUES ('model4@shiba.test', crypt('Model1234!', gen_salt('bf', 10)), 'model', 'active')
ON CONFLICT (email) DO NOTHING
RETURNING id INTO v_m4_user;
IF v_m4_user IS NULL THEN
  SELECT id INTO v_m4_user FROM users WHERE email = 'model4@shiba.test';
END IF;

INSERT INTO model_profiles (
  user_id, first_name, last_name, birth_date,
  city, country, willing_to_relocate,
  height_cm, weight_kg, chest_cm, waist_cm, hips_cm,
  shoe_size, clothing_size, phone, bio
)
VALUES (
  v_m4_user, 'Ольга', 'Петрова', '1994-03-17',
  'Москва', 'Россия', false,
  168, 72, 98, 78, 106,
  38.0, 'L/XL', '+7 916 567-89-01',
  'Plus-size модель с богатым портфолио. Участвовала в кампаниях брендов Torrid, Marina Rinaldi, Lane Bryant Russia. Амбассадор движения body positivity. Выступаю за разнообразие в индустрии моды.'
)
ON CONFLICT (user_id) DO NOTHING
RETURNING id INTO v_m4_prof;
IF v_m4_prof IS NULL THEN
  SELECT id INTO v_m4_prof FROM model_profiles WHERE user_id = v_m4_user;
END IF;

INSERT INTO model_categories (model_profile_id, category_id)
SELECT v_m4_prof, id FROM categories WHERE name IN ('Plus Size', 'Commercial', 'Fashion')
ON CONFLICT DO NOTHING;

-- Model 5: Дарья Новикова (editorial/alternative)
INSERT INTO users (email, password_hash, role, status)
VALUES ('model5@shiba.test', crypt('Model1234!', gen_salt('bf', 10)), 'model', 'active')
ON CONFLICT (email) DO NOTHING
RETURNING id INTO v_m5_user;
IF v_m5_user IS NULL THEN
  SELECT id INTO v_m5_user FROM users WHERE email = 'model5@shiba.test';
END IF;

INSERT INTO model_profiles (
  user_id, first_name, last_name, birth_date,
  city, country, willing_to_relocate,
  height_cm, weight_kg, chest_cm, waist_cm, hips_cm,
  shoe_size, clothing_size, phone, bio
)
VALUES (
  v_m5_user, 'Дарья', 'Новикова', '2001-07-29',
  'Новосибирск', 'Россия', true,
  178, 57, 87, 61, 91,
  38.5, 'S/M', '+7 913 678-90-12',
  'Editorial и alternative модель. Снималась для журналов Tatler, Esquire, GQ Russia. Имею нестандартную внешность, что делает меня уникальной для авторских проектов. Опыт в театре и современном танце.'
)
ON CONFLICT (user_id) DO NOTHING
RETURNING id INTO v_m5_prof;
IF v_m5_prof IS NULL THEN
  SELECT id INTO v_m5_prof FROM model_profiles WHERE user_id = v_m5_user;
END IF;

INSERT INTO model_categories (model_profile_id, category_id)
SELECT v_m5_prof, id FROM categories WHERE name IN ('Editorial', 'Alternative', 'Fashion')
ON CONFLICT DO NOTHING;

-- Model 6: Полина Морозова (petite/commercial)
INSERT INTO users (email, password_hash, role, status)
VALUES ('model6@shiba.test', crypt('Model1234!', gen_salt('bf', 10)), 'model', 'active')
ON CONFLICT (email) DO NOTHING
RETURNING id INTO v_m6_user;
IF v_m6_user IS NULL THEN
  SELECT id INTO v_m6_user FROM users WHERE email = 'model6@shiba.test';
END IF;

INSERT INTO model_profiles (
  user_id, first_name, last_name, birth_date,
  city, country, willing_to_relocate,
  height_cm, weight_kg, chest_cm, waist_cm, hips_cm,
  shoe_size, clothing_size, phone, bio
)
VALUES (
  v_m6_user, 'Полина', 'Морозова', '2003-01-11',
  'Казань', 'Россия', true,
  162, 48, 82, 58, 86,
  35.5, 'XS', '+7 917 789-01-23',
  'Petite модель. Специализируюсь на коммерческих съёмках и рекламе. Работала с OZON, Wildberries, СберМегаМаркет. Имею опыт в детской рекламе (съёмки с детьми от 3 до 10 лет). Всегда пунктуальна и профессиональна.'
)
ON CONFLICT (user_id) DO NOTHING
RETURNING id INTO v_m6_prof;
IF v_m6_prof IS NULL THEN
  SELECT id INTO v_m6_prof FROM model_profiles WHERE user_id = v_m6_user;
END IF;

INSERT INTO model_categories (model_profile_id, category_id)
SELECT v_m6_prof, id FROM categories WHERE name IN ('Petite', 'Commercial', 'Kids')
ON CONFLICT DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- CASTINGS
-- ─────────────────────────────────────────────────────────────────────────────

-- Casting 1 (Elite / Fashion)
INSERT INTO castings (agency_profile_id, title, description, city, casting_date, category_id, status)
SELECT
  v_ag1_prof,
  'Кастинг для весенней коллекции 2026',
  'Ищем моделей для съёмки весенней коллекции бренда SHIBA. Рост от 172 см, размер XS–S. Съёмка состоится 20–22 марта в нашей студии в Москве. Оплата: 15 000 ₽/день.',
  'Москва', '2026-03-20',
  (SELECT id FROM categories WHERE name = 'Fashion'), 'active'
WHERE NOT EXISTS (
  SELECT 1 FROM castings WHERE agency_profile_id = v_ag1_prof
    AND title = 'Кастинг для весенней коллекции 2026'
)
RETURNING id INTO v_c1;
IF v_c1 IS NULL THEN
  SELECT id INTO v_c1 FROM castings WHERE agency_profile_id = v_ag1_prof
    AND title = 'Кастинг для весенней коллекции 2026';
END IF;

-- Casting 2 (Elite / Editorial)
INSERT INTO castings (agency_profile_id, title, description, city, casting_date, category_id, status)
SELECT
  v_ag1_prof,
  'Editorial съёмка для Vogue Russia',
  'Приглашаем editorial-моделей для участия в эксклюзивной съёмке для Vogue Russia. Тема: «Русская зима». Рост от 175 см. Опыт editorial обязателен. Оплата обсуждается индивидуально.',
  'Москва', '2026-04-05',
  (SELECT id FROM categories WHERE name = 'Editorial'), 'active'
WHERE NOT EXISTS (
  SELECT 1 FROM castings WHERE agency_profile_id = v_ag1_prof
    AND title = 'Editorial съёмка для Vogue Russia'
)
RETURNING id INTO v_c2;
IF v_c2 IS NULL THEN
  SELECT id INTO v_c2 FROM castings WHERE agency_profile_id = v_ag1_prof
    AND title = 'Editorial съёмка для Vogue Russia';
END IF;

-- Casting 3 (Elite / Beauty)
INSERT INTO castings (agency_profile_id, title, description, city, casting_date, category_id, status)
SELECT
  v_ag1_prof,
  'Реклама косметики Faberlic — весна',
  'Ищем beauty-модель для рекламной кампании нового ухода за кожей Faberlic. Требования: фотогеничное лицо, опыт beauty-съёмок, чистая кожа. Гонорар: 25 000 ₽.',
  'Москва', '2026-03-28',
  (SELECT id FROM categories WHERE name = 'Beauty'), 'active'
WHERE NOT EXISTS (
  SELECT 1 FROM castings WHERE agency_profile_id = v_ag1_prof
    AND title = 'Реклама косметики Faberlic — весна'
)
RETURNING id INTO v_c3;
IF v_c3 IS NULL THEN
  SELECT id INTO v_c3 FROM castings WHERE agency_profile_id = v_ag1_prof
    AND title = 'Реклама косметики Faberlic — весна';
END IF;

-- Casting 4 (Stars / Fitness)
INSERT INTO castings (agency_profile_id, title, description, city, casting_date, category_id, status)
SELECT
  v_ag2_prof,
  'Фитнес-кампания Reebok Spring 2026',
  'Stars Talent Management ищет фитнес-моделей для весенней кампании Reebok. Нужны атлетичные девушки с подтянутым телом. Рост от 168 см. Съёмка в Москве и Сочи. Гонорар: 30 000 ₽.',
  'Москва', '2026-04-10',
  (SELECT id FROM categories WHERE name = 'Fitness'), 'active'
WHERE NOT EXISTS (
  SELECT 1 FROM castings WHERE agency_profile_id = v_ag2_prof
    AND title = 'Фитнес-кампания Reebok Spring 2026'
)
RETURNING id INTO v_c4;
IF v_c4 IS NULL THEN
  SELECT id INTO v_c4 FROM castings WHERE agency_profile_id = v_ag2_prof
    AND title = 'Фитнес-кампания Reebok Spring 2026';
END IF;

-- Casting 5 (Stars / Swimwear)
INSERT INTO castings (agency_profile_id, title, description, city, casting_date, category_id, status)
SELECT
  v_ag2_prof,
  'Каталог купальников — лето 2026',
  'Ищем моделей для съёмки летнего каталога купальников. Рост 168–180 см, размер S–M. Съёмка пройдёт на Мальдивах в мае. Всё включено + гонорар 50 000 ₽.',
  'Мальдивы', '2026-05-15',
  (SELECT id FROM categories WHERE name = 'Swimwear'), 'active'
WHERE NOT EXISTS (
  SELECT 1 FROM castings WHERE agency_profile_id = v_ag2_prof
    AND title = 'Каталог купальников — лето 2026'
)
RETURNING id INTO v_c5;
IF v_c5 IS NULL THEN
  SELECT id INTO v_c5 FROM castings WHERE agency_profile_id = v_ag2_prof
    AND title = 'Каталог купальников — лето 2026';
END IF;

-- Casting 6 (Stars / Commercial)
INSERT INTO castings (agency_profile_id, title, description, city, casting_date, category_id, status)
SELECT
  v_ag2_prof,
  'Рекламный ролик Сбербанка',
  'Ищем коммерческую модель для съёмки рекламного ролика Сбербанка. Образ: успешная деловая женщина 25–35 лет. Рост 165–175 см. Гонорар за день съёмки: 45 000 ₽.',
  'Москва', '2026-03-25',
  (SELECT id FROM categories WHERE name = 'Commercial'), 'active'
WHERE NOT EXISTS (
  SELECT 1 FROM castings WHERE agency_profile_id = v_ag2_prof
    AND title = 'Рекламный ролик Сбербанка'
)
RETURNING id INTO v_c6;
IF v_c6 IS NULL THEN
  SELECT id INTO v_c6 FROM castings WHERE agency_profile_id = v_ag2_prof
    AND title = 'Рекламный ролик Сбербанка';
END IF;

-- Casting 7 (Moscow Fashion Group / Plus Size)
INSERT INTO castings (agency_profile_id, title, description, city, casting_date, category_id, status)
SELECT
  v_ag3_prof,
  'Mercedes-Benz Fashion Week — Plus Size',
  'Moscow Fashion Group приглашает plus-size моделей для участия в показе на MBFW Russia. Размеры L–3XL. Уникальная возможность выступить на главном модном событии страны. Гонорар 20 000 ₽.',
  'Москва', '2026-04-20',
  (SELECT id FROM categories WHERE name = 'Plus Size'), 'active'
WHERE NOT EXISTS (
  SELECT 1 FROM castings WHERE agency_profile_id = v_ag3_prof
    AND title = 'Mercedes-Benz Fashion Week — Plus Size'
)
RETURNING id INTO v_c7;
IF v_c7 IS NULL THEN
  SELECT id INTO v_c7 FROM castings WHERE agency_profile_id = v_ag3_prof
    AND title = 'Mercedes-Benz Fashion Week — Plus Size';
END IF;

-- Casting 8 (Moscow Fashion Group / Alternative)
INSERT INTO castings (agency_profile_id, title, description, city, casting_date, category_id, status)
SELECT
  v_ag3_prof,
  'Авторская съёмка Татлер — «Другие»',
  'Ищем alternative-моделей для арт-проекта журнала Tatler. Приветствуется нестандартная внешность, татуировки, необычный типаж. Это ваш шанс заявить о себе! Гонорар обсуждается.',
  'Санкт-Петербург', '2026-04-15',
  (SELECT id FROM categories WHERE name = 'Alternative'), 'active'
WHERE NOT EXISTS (
  SELECT 1 FROM castings WHERE agency_profile_id = v_ag3_prof
    AND title = 'Авторская съёмка Татлер — «Другие»'
)
RETURNING id INTO v_c8;
IF v_c8 IS NULL THEN
  SELECT id INTO v_c8 FROM castings WHERE agency_profile_id = v_ag3_prof
    AND title = 'Авторская съёмка Татлер — «Другие»';
END IF;

-- ─────────────────────────────────────────────────────────────────────────────
-- APPLICATIONS (models applying to castings)
-- ─────────────────────────────────────────────────────────────────────────────

-- Anna applies to spring fashion casting & editorial
INSERT INTO applications (casting_id, model_profile_id, status, model_message)
VALUES (v_c1, v_m1_prof, 'pending',
  'Здравствуйте! Я заинтересована в участии в вашей весенней съёмке. Имею опыт работы с аналогичными брендами. Готова предоставить дополнительное портфолио по запросу.')
ON CONFLICT (casting_id, model_profile_id) DO NOTHING;

INSERT INTO applications (casting_id, model_profile_id, status, model_message)
VALUES (v_c2, v_m1_prof, 'viewed',
  'Обожаю editorial проекты! Vogue Russia — это мечта. Приложу расширенное портфолио с editorial работами.')
ON CONFLICT (casting_id, model_profile_id) DO NOTHING;

-- Maria applies to beauty casting & commercial
INSERT INTO applications (casting_id, model_profile_id, status, model_message)
VALUES (v_c3, v_m2_prof, 'accepted',
  'Имею большой опыт в beauty-съёмках. Работала с Faberlic ранее. Знаю все их требования к подаче продукта.')
ON CONFLICT (casting_id, model_profile_id) DO NOTHING;

INSERT INTO applications (casting_id, model_profile_id, status, model_message, agency_response)
VALUES (v_c6, v_m2_prof, 'accepted',
  'Деловой образ — моё второе «я». Имею опыт в банковской рекламе.',
  'Отличное портфолио! Вы нам подходите. Свяжемся для обсуждения деталей.')
ON CONFLICT (casting_id, model_profile_id) DO NOTHING;

-- Katya applies to fitness & swimwear
INSERT INTO applications (casting_id, model_profile_id, status, model_message)
VALUES (v_c4, v_m3_prof, 'accepted',
  'КМС по лёгкой атлетике, 7 лет занимаюсь спортом. Работала с Reebok на региональных кампаниях. Готова к интенсивной съёмке.')
ON CONFLICT (casting_id, model_profile_id) DO NOTHING;

INSERT INTO applications (casting_id, model_profile_id, status, model_message)
VALUES (v_c5, v_m3_prof, 'pending',
  'Мечтаю о съёмке на Мальдивах! Имею отличную физическую форму и опыт съёмок в купальниках.')
ON CONFLICT (casting_id, model_profile_id) DO NOTHING;

-- Olga applies to plus size casting
INSERT INTO applications (casting_id, model_profile_id, status, model_message)
VALUES (v_c7, v_m4_prof, 'pending',
  'MBFW — это то, к чему я шла 3 года. Я представляю настоящую красоту во всём её разнообразии.')
ON CONFLICT (casting_id, model_profile_id) DO NOTHING;

-- Darya applies to editorial & alternative
INSERT INTO applications (casting_id, model_profile_id, status, model_message)
VALUES (v_c2, v_m5_prof, 'rejected',
  'Снималась для Tatler и GQ. Editorial — моя специализация.')
ON CONFLICT (casting_id, model_profile_id) DO NOTHING;

INSERT INTO applications (casting_id, model_profile_id, status, model_message)
VALUES (v_c8, v_m5_prof, 'viewed',
  'Это моё! Люблю авторские проекты, всегда иду на эксперименты. Есть несколько тату и необычная внешность.')
ON CONFLICT (casting_id, model_profile_id) DO NOTHING;

-- Polina applies to commercial & petite-friendly castings
INSERT INTO applications (casting_id, model_profile_id, status, model_message)
VALUES (v_c6, v_m6_prof, 'pending',
  'Опыт в коммерческих съёмках для крупных брендов. Пунктуальна, профессиональна, легко нахожу нужный образ.')
ON CONFLICT (casting_id, model_profile_id) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- INVITATIONS (agencies inviting models)
-- ─────────────────────────────────────────────────────────────────────────────

-- Elite invites Katya to fitness casting (cross-category invitation)
INSERT INTO invitations (agency_profile_id, model_profile_id, casting_id, message, status)
VALUES (
  v_ag1_prof, v_m3_prof, v_c1,
  'Екатерина, мы изучили ваше портфолио и хотим пригласить вас на нашу весеннюю съёмку. Ваша спортивная форма подойдёт для нового образа коллекции.',
  'pending'
)
ON CONFLICT DO NOTHING;

-- Stars invites Anna to fitness campaign
INSERT INTO invitations (agency_profile_id, model_profile_id, casting_id, message, status)
VALUES (
  v_ag2_prof, v_m1_prof, v_c4,
  'Анна, ваша модельная внешность идеально подойдёт для нашего нового проекта с Reebok. Мы верим, что это сотрудничество будет продуктивным!',
  'pending'
)
ON CONFLICT DO NOTHING;

-- Moscow Fashion Group invites Darya to alternative project
INSERT INTO invitations (agency_profile_id, model_profile_id, casting_id, message, status)
VALUES (
  v_ag3_prof, v_m5_prof, v_c8,
  'Дарья, ваш нестандартный образ — именно то, что нам нужно для арт-проекта Tatler. Уверены, что ваше участие сделает проект особенным.',
  'accepted'
)
ON CONFLICT DO NOTHING;

END $$;
