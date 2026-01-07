-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    metadata TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create faces table
CREATE TABLE IF NOT EXISTS faces (
    id VARCHAR(36) PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    embedding TEXT NOT NULL,
    quality_score REAL NOT NULL DEFAULT 0,
    enrolled_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on faces.user_id
CREATE INDEX IF NOT EXISTS idx_faces_user_id ON faces(user_id);

-- Create settings table
CREATE TABLE IF NOT EXISTS settings (
    id INTEGER PRIMARY KEY,
    match_threshold REAL NOT NULL DEFAULT 0.6,
    max_faces_per_user INTEGER NOT NULL DEFAULT 10,
    embedding_dimension INTEGER NOT NULL DEFAULT 128
);

-- Insert default settings
INSERT INTO settings (id, match_threshold, max_faces_per_user, embedding_dimension)
VALUES (1, 0.6, 10, 128);
