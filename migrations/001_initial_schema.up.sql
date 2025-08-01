
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE files (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    filename TEXT NOT NULL,
    media_type VARCHAR(255) NOT NULL,
    upload_timestamp TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    md5 VARCHAR(32) UNIQUE -- A Unique constraint for the file to prevent multiple uploads of the same file
);

CREATE TABLE hashes (
    id BIGSERIAL PRIMARY KEY,
    file_uuid UUID NOT NULL REFERENCES files(uuid) ON DELETE CASCADE,
    algorithm VARCHAR(50) NOT NULL,
    hash_value VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(file_uuid, algorithm) -- A file should only have one hash per algorithm
);

CREATE TABLE watermark_history (
    id BIGSERIAL PRIMARY KEY,
    file_uuid UUID NOT NULL REFERENCES files(uuid) ON DELETE CASCADE,
    algorithm VARCHAR(50) NOT NULL,
    embedded_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_files_uuid ON files(uuid);
CREATE INDEX idx_hashes_file_uuid ON hashes(file_uuid);
CREATE INDEX idx_hashes_algorithm_hash_value ON hashes(algorithm, hash_value);
CREATE INDEX idx_watermark_history_file_uuid ON watermark_history(file_uuid);
