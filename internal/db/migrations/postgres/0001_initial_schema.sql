-- migrate:up
CREATE TABLE IF NOT EXISTS documents (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    title TEXT,
    url TEXT UNIQUE,
    source TEXT,
    cover_url TEXT,
    author TEXT,
    studio TEXT,
    synopsis TEXT,
    genres TEXT,
    status TEXT,
    type TEXT DEFAULT 'text',
    is_in_library BOOLEAN DEFAULT false,
    is_archived BOOLEAN DEFAULT false,
    is_local BOOLEAN DEFAULT false,
    local_path TEXT,
    group_id INTEGER
);
CREATE INDEX IF NOT EXISTS idx_documents_deleted_at ON documents(deleted_at);
CREATE INDEX IF NOT EXISTS idx_documents_title ON documents(title);
CREATE INDEX IF NOT EXISTS idx_documents_url ON documents(url);
CREATE INDEX IF NOT EXISTS idx_documents_source ON documents(source);
CREATE INDEX IF NOT EXISTS idx_documents_type ON documents(type);
CREATE INDEX IF NOT EXISTS idx_documents_group_id ON documents(group_id);

CREATE TABLE IF NOT EXISTS chapters (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    document_id INTEGER,
    title TEXT,
    url TEXT,
    content TEXT,
    metadata TEXT,
    order_val INTEGER,
    status TEXT,
    is_read BOOLEAN DEFAULT false
);
CREATE INDEX IF NOT EXISTS idx_chapters_deleted_at ON chapters(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_doc_url ON chapters(document_id, url);
CREATE INDEX IF NOT EXISTS idx_chapters_order ON chapters(order_val);

CREATE TABLE IF NOT EXISTS reading_progresses (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    document_id INTEGER UNIQUE,
    chapter_id INTEGER,
    scroll_pos DOUBLE PRECISION,
    client_updated_at BIGINT
);

CREATE TABLE IF NOT EXISTS groups (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    name TEXT
);

CREATE TABLE IF NOT EXISTS cache_items (
    key TEXT PRIMARY KEY,
    value BYTEA,
    expires_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS bookmarks (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    document_id INTEGER,
    chapter_id INTEGER,
    title TEXT
);

CREATE TABLE IF NOT EXISTS notes (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    document_id INTEGER,
    content TEXT,
    quote TEXT
);

CREATE TABLE IF NOT EXISTS plugins (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE,
    is_enabled BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS library_paths (
    id SERIAL PRIMARY KEY,
    path TEXT,
    pattern TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reading_stats (
    id SERIAL PRIMARY KEY,
    date TEXT,
    read_seconds INTEGER,
    documents_read INTEGER,
    chapters_read INTEGER
);
