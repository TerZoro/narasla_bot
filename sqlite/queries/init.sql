CREATE TABLE IF NOT EXISTS pages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id INTEGER NOT NULL,
    chat_id INTEGER NOT NULL,
    url TEXT NOT NULL,
    user_name TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(owner_id, url)
);

CREATE INDEX IF NOT EXISTS idx_pages_owner_id ON pages(owner_id);


CREATE TABLE IF NOT EXISTS users (
    owner_id INTEGER PRIMARY KEY,
    chat_id INTEGER NOT NULL,
    user_name TEXT,
    timezone TEXT NOT NULL DEFAULT 'Asia/Almaty',
    enabled INTEGER NOT NULL CHECK (enabled IN (0, 1)) DEFAULT 1,
    send_hour INTEGER NOT NULL CHECK (send_hour BETWEEN 0 AND 23) DEFAULT 12,
    send_minute INTEGER NOT NULL CHECK (send_minute BETWEEN 0 AND 59) DEFAULT 0,
    last_send_at INTEGER
);

CREATE INDEX IF NOT EXISTS idx_users_enabled ON users(enabled);
