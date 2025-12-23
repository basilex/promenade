-- Create languages table for language reference data
CREATE TABLE IF NOT EXISTS core_languages (
    id UUID PRIMARY KEY DEFAULT uuid_v7(),
    name VARCHAR(100) NOT NULL,                 -- English name (e.g., "English")
    native_name VARCHAR(100) NOT NULL,          -- Native name (e.g., "English")
    code CHAR(2) NOT NULL UNIQUE,               -- ISO 639-1 code (e.g., "en")
    iso639_2 CHAR(3) NOT NULL UNIQUE,           -- ISO 639-2 code (e.g., "eng")
    is_rtl BOOLEAN NOT NULL DEFAULT false,      -- Right-to-left (Arabic, Hebrew)
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_core_languages_code ON core_languages(code);
CREATE INDEX idx_core_languages_is_active ON core_languages(is_active);
CREATE INDEX idx_core_languages_sort_order ON core_languages(sort_order);

-- Insert common languages (sorted by usage)
INSERT INTO core_languages (name, native_name, code, iso639_2, is_rtl, sort_order) VALUES
('English', 'English', 'en', 'eng', false, 1),
('Russian', 'Русский', 'ru', 'rus', false, 2),
('Spanish', 'Español', 'es', 'spa', false, 3),
('Chinese', '中文', 'zh', 'zho', false, 4),
('French', 'Français', 'fr', 'fra', false, 5),
('German', 'Deutsch', 'de', 'deu', false, 6),
('Italian', 'Italiano', 'it', 'ita', false, 7),
('Portuguese', 'Português', 'pt', 'por', false, 8),
('Japanese', '日本語', 'ja', 'jpn', false, 9),
('Korean', '한국어', 'ko', 'kor', false, 10),
('Arabic', 'العربية', 'ar', 'ara', true, 11),
('Turkish', 'Türkçe', 'tr', 'tur', false, 12),
('Polish', 'Polski', 'pl', 'pol', false, 13),
('Ukrainian', 'Українська', 'uk', 'ukr', false, 14),
('Dutch', 'Nederlands', 'nl', 'nld', false, 15),
('Swedish', 'Svenska', 'sv', 'swe', false, 16),
('Hindi', 'हिन्दी', 'hi', 'hin', false, 17),
('Hebrew', 'עברית', 'he', 'heb', true, 18),
('Thai', 'ไทย', 'th', 'tha', false, 19),
('Vietnamese', 'Tiếng Việt', 'vi', 'vie', false, 20);

-- Add comment
COMMENT ON TABLE core_languages IS 'Reference data for languages (ISO 639-1 and ISO 639-2)';
