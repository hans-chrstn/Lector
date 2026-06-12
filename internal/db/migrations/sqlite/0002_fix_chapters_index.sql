DROP INDEX IF EXISTS idx_doc_url;
CREATE UNIQUE INDEX idx_doc_url ON chapters(document_id, url);
