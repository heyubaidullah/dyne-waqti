-- Idempotent schema for mosque-display. All statements are safe to re-run on
-- every startup; CREATE TABLE IF NOT EXISTS never touches existing data.
-- If a future release needs to alter an existing table's columns, add a
-- schema_version row to `settings` and a small ordered ALTER-statement
-- runner in migrate.go — not needed at this scale yet.

CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS prayer_schedules (
    date TEXT PRIMARY KEY,
    fajr_iqamah TEXT NOT NULL,
    dhuhr_iqamah TEXT NOT NULL,
    asr_iqamah TEXT NOT NULL,
    maghrib_iqamah TEXT NOT NULL,
    isha_iqamah TEXT NOT NULL,
    jumuah_iqamah TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS slides (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    type TEXT CHECK(type IN ('image', 'text_verse')) NOT NULL,
    content_url_or_text TEXT NOT NULL,
    arabic_text TEXT,
    is_active INTEGER DEFAULT 1,
    expiration_date TEXT,
    display_duration_sec INTEGER DEFAULT 10,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS emergency_notices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    deceased_name TEXT NOT NULL,
    prayer_time TEXT NOT NULL,
    location TEXT NOT NULL,
    is_active INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admin_credentials (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    passphrase_hash TEXT NOT NULL,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS login_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip_or_session TEXT NOT NULL,
    attempted_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    success INTEGER NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_login_attempts_ip_time
    ON login_attempts (ip_or_session, attempted_at);

CREATE INDEX IF NOT EXISTS idx_slides_active
    ON slides (is_active);
