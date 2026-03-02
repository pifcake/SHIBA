ALTER TABLE model_profiles DROP COLUMN IF EXISTS gender;
ALTER TABLE model_profiles DROP COLUMN IF EXISTS hair_color;
ALTER TABLE model_profiles DROP COLUMN IF EXISTS hair_length;
ALTER TABLE model_profiles DROP COLUMN IF EXISTS hair_structure;
ALTER TABLE model_profiles DROP COLUMN IF EXISTS eye_color;
ALTER TABLE model_profiles DROP COLUMN IF EXISTS clothing_size_top;
ALTER TABLE model_profiles DROP COLUMN IF EXISTS clothing_size_bot;

DELETE FROM categories WHERE name IN (
    'Модели для съемок / Фотомодели',
    'Промо модели',
    'Бельевые модели',
    'Модели на показ',
    'Фотосъемка',
    'Видеосъемка'
);
