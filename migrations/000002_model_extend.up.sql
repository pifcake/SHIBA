-- New model profile fields
ALTER TABLE model_profiles ADD COLUMN IF NOT EXISTS gender             VARCHAR(10)  NOT NULL DEFAULT '';
ALTER TABLE model_profiles ADD COLUMN IF NOT EXISTS hair_color         VARCHAR(50)  NOT NULL DEFAULT '';
ALTER TABLE model_profiles ADD COLUMN IF NOT EXISTS hair_length        VARCHAR(50)  NOT NULL DEFAULT '';
ALTER TABLE model_profiles ADD COLUMN IF NOT EXISTS hair_structure     VARCHAR(50)  NOT NULL DEFAULT '';
ALTER TABLE model_profiles ADD COLUMN IF NOT EXISTS eye_color          VARCHAR(50)  NOT NULL DEFAULT '';
ALTER TABLE model_profiles ADD COLUMN IF NOT EXISTS clothing_size_top  VARCHAR(20)  NOT NULL DEFAULT '';
ALTER TABLE model_profiles ADD COLUMN IF NOT EXISTS clothing_size_bot  VARCHAR(20)  NOT NULL DEFAULT '';

-- New categories
INSERT INTO categories (name) VALUES
    ('Модели для съемок / Фотомодели'),
    ('Промо модели'),
    ('Бельевые модели'),
    ('Модели на показ'),
    ('Фотосъемка'),
    ('Видеосъемка')
ON CONFLICT (name) DO NOTHING;
