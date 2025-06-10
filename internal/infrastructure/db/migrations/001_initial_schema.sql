-- Initial schema migration
DROP TABLE IF EXISTS pages;
DROP TABLE IF EXISTS documents;
DROP TABLE IF EXISTS splits;

CREATE TABLE splits (
    id TEXT PRIMARY KEY,
    client_id TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

CREATE TABLE documents (
    id TEXT PRIMARY KEY,
    split_id TEXT NOT NULL,
    name TEXT NOT NULL,
    classification TEXT,
    filename TEXT,
    short_description TEXT,
    start_page TEXT,
    end_page TEXT,
    FOREIGN KEY (split_id) REFERENCES splits(id)
);

CREATE TABLE pages (
    id TEXT PRIMARY KEY,
    split_id TEXT NOT NULL,
    document_id TEXT,
    page_number TEXT NOT NULL,
    url TEXT NOT NULL,
    FOREIGN KEY (split_id) REFERENCES splits(id),
    FOREIGN KEY (document_id) REFERENCES documents(id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_documents_split_id ON documents(split_id);
CREATE INDEX IF NOT EXISTS idx_pages_document_id ON pages(document_id);
CREATE INDEX IF NOT EXISTS idx_pages_split_id ON pages(split_id); 