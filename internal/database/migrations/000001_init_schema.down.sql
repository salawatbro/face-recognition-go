-- Drop index
DROP INDEX IF EXISTS idx_faces_user_id;

-- Drop tables in reverse order
DROP TABLE IF EXISTS settings;
DROP TABLE IF EXISTS faces;
DROP TABLE IF EXISTS users;
