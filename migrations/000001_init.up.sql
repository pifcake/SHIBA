-- Extensions
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Enums
CREATE TYPE user_role AS ENUM ('admin', 'agency', 'model');
CREATE TYPE user_status AS ENUM ('pending_email', 'active', 'blocked');
CREATE TYPE agency_status AS ENUM ('pending_approval', 'active', 'blocked');
CREATE TYPE moderation_status AS ENUM ('pending', 'approved', 'rejected');
CREATE TYPE casting_status AS ENUM ('active', 'closed', 'cancelled');
CREATE TYPE application_status AS ENUM ('pending', 'viewed', 'accepted', 'rejected');
CREATE TYPE complaint_target_type AS ENUM ('casting', 'invitation', 'agency', 'model');
CREATE TYPE complaint_reason AS ENUM ('inappropriate_content', 'harassment', 'fraud', 'other');
CREATE TYPE complaint_status AS ENUM ('open', 'reviewed', 'resolved', 'dismissed');

-- Users
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email       VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    role        user_role NOT NULL,
    status      user_status NOT NULL DEFAULT 'pending_email',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role_status ON users(role, status);

-- Email verifications
CREATE TABLE email_verifications (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token       VARCHAR(255) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    used_at     TIMESTAMPTZ
);

CREATE INDEX idx_email_verifications_token ON email_verifications(token);
CREATE INDEX idx_email_verifications_user_id ON email_verifications(user_id);

-- Refresh tokens
CREATE TABLE refresh_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  VARCHAR(255) NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    revoked_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Reference: categories
CREATE TABLE categories (
    id      SMALLSERIAL PRIMARY KEY,
    name    VARCHAR(100) NOT NULL UNIQUE
);

INSERT INTO categories (name) VALUES
    ('Fashion'),
    ('Commercial'),
    ('Editorial'),
    ('Fitness'),
    ('Lingerie'),
    ('Swimwear'),
    ('Beauty'),
    ('Plus Size'),
    ('Petite'),
    ('Kids'),
    ('Mature'),
    ('Alternative');

-- Reference: photo tags
CREATE TABLE photo_tags (
    id      SMALLSERIAL PRIMARY KEY,
    name    VARCHAR(100) NOT NULL UNIQUE
);

INSERT INTO photo_tags (name) VALUES
    ('portrait'),
    ('full body'),
    ('editorial'),
    ('commercial'),
    ('fashion'),
    ('fitness'),
    ('beauty'),
    ('outdoor'),
    ('studio'),
    ('black and white');

-- Model profiles
CREATE TABLE model_profiles (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    first_name          VARCHAR(100) NOT NULL DEFAULT '',
    last_name           VARCHAR(100) NOT NULL DEFAULT '',
    birth_date          DATE,
    city                VARCHAR(100) NOT NULL DEFAULT '',
    country             VARCHAR(100) NOT NULL DEFAULT '',
    willing_to_relocate BOOLEAN NOT NULL DEFAULT FALSE,
    height_cm           SMALLINT,
    weight_kg           SMALLINT,
    chest_cm            SMALLINT,
    waist_cm            SMALLINT,
    hips_cm             SMALLINT,
    shoe_size           NUMERIC(4,1),
    clothing_size       VARCHAR(20) NOT NULL DEFAULT '',
    phone               VARCHAR(50) NOT NULL DEFAULT '',
    bio                 TEXT NOT NULL DEFAULT '',
    shoot_restrictions  TEXT[] NOT NULL DEFAULT '{}',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_profiles_user_id ON model_profiles(user_id);
CREATE INDEX idx_model_profiles_city ON model_profiles(city);
CREATE INDEX idx_model_profiles_birth_date ON model_profiles(birth_date);
CREATE INDEX idx_model_profiles_willing_to_relocate ON model_profiles(willing_to_relocate);

-- Model categories (M2M)
CREATE TABLE model_categories (
    model_profile_id    UUID NOT NULL REFERENCES model_profiles(id) ON DELETE CASCADE,
    category_id         SMALLINT NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (model_profile_id, category_id)
);

-- Agency profiles
CREATE TABLE agency_profiles (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    company_name    VARCHAR(255) NOT NULL DEFAULT '',
    description     TEXT NOT NULL DEFAULT '',
    phone           VARCHAR(50) NOT NULL DEFAULT '',
    logo_url        TEXT NOT NULL DEFAULT '',
    status          agency_status NOT NULL DEFAULT 'pending_approval',
    approved_by     UUID REFERENCES users(id),
    approved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_agency_profiles_user_id ON agency_profiles(user_id);
CREATE INDEX idx_agency_profiles_status ON agency_profiles(status);

-- Model photos
CREATE TABLE model_photos (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_profile_id    UUID NOT NULL REFERENCES model_profiles(id) ON DELETE CASCADE,
    storage_key         VARCHAR(500) NOT NULL,
    original_name       VARCHAR(255) NOT NULL DEFAULT '',
    size_bytes          BIGINT NOT NULL DEFAULT 0,
    mime_type           VARCHAR(100) NOT NULL DEFAULT '',
    width_px            INT,
    height_px           INT,
    moderation_status   moderation_status NOT NULL DEFAULT 'pending',
    moderated_by        UUID REFERENCES users(id),
    moderated_at        TIMESTAMPTZ,
    rejection_reason    TEXT NOT NULL DEFAULT '',
    is_cover            BOOLEAN NOT NULL DEFAULT FALSE,
    sort_order          SMALLINT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_model_photos_model_profile_id ON model_photos(model_profile_id);
CREATE INDEX idx_model_photos_moderation_status ON model_photos(moderation_status);

-- Photo tags (M2M)
CREATE TABLE model_photo_tags (
    photo_id        UUID NOT NULL REFERENCES model_photos(id) ON DELETE CASCADE,
    photo_tag_id    SMALLINT NOT NULL REFERENCES photo_tags(id) ON DELETE CASCADE,
    PRIMARY KEY (photo_id, photo_tag_id)
);

-- Castings
CREATE TABLE castings (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_profile_id   UUID NOT NULL REFERENCES agency_profiles(id) ON DELETE CASCADE,
    title               VARCHAR(255) NOT NULL,
    description         TEXT NOT NULL DEFAULT '',
    city                VARCHAR(100) NOT NULL DEFAULT '',
    casting_date        DATE,
    category_id         SMALLINT REFERENCES categories(id),
    status              casting_status NOT NULL DEFAULT 'active',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_castings_agency_profile_id ON castings(agency_profile_id);
CREATE INDEX idx_castings_status ON castings(status);
CREATE INDEX idx_castings_city ON castings(city);
CREATE INDEX idx_castings_category_id ON castings(category_id);

-- Applications (model -> casting)
CREATE TABLE applications (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    casting_id          UUID NOT NULL REFERENCES castings(id) ON DELETE CASCADE,
    model_profile_id    UUID NOT NULL REFERENCES model_profiles(id) ON DELETE CASCADE,
    status              application_status NOT NULL DEFAULT 'pending',
    model_message       TEXT NOT NULL DEFAULT '',
    agency_response     TEXT NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (casting_id, model_profile_id)
);

CREATE INDEX idx_applications_casting_id ON applications(casting_id);
CREATE INDEX idx_applications_model_profile_id ON applications(model_profile_id);

-- Invitations (agency -> model)
CREATE TABLE invitations (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agency_profile_id   UUID NOT NULL REFERENCES agency_profiles(id) ON DELETE CASCADE,
    model_profile_id    UUID NOT NULL REFERENCES model_profiles(id) ON DELETE CASCADE,
    casting_id          UUID REFERENCES castings(id) ON DELETE SET NULL,
    message             TEXT NOT NULL DEFAULT '',
    status              application_status NOT NULL DEFAULT 'pending',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_invitations_agency_profile_id ON invitations(agency_profile_id);
CREATE INDEX idx_invitations_model_profile_id ON invitations(model_profile_id);

-- Complaints
CREATE TABLE complaints (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_type     complaint_target_type NOT NULL,
    target_id       UUID NOT NULL,
    reason_category complaint_reason NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    status          complaint_status NOT NULL DEFAULT 'open',
    reviewed_by     UUID REFERENCES users(id),
    resolution_note TEXT NOT NULL DEFAULT '',
    resolved_at     TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_complaints_reporter_id ON complaints(reporter_id);
CREATE INDEX idx_complaints_status ON complaints(status);
CREATE INDEX idx_complaints_target_type_id ON complaints(target_type, target_id);
