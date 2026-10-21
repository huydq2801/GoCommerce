-- drop trigger
DROP TRIGGER IF EXISTS set_user_updated_at ON users;

-- drop triger function
DROP FUNCTION IF EXISTS update_user_updated_at_column();

-- drop index
DROP INDEX IF EXISTS idx_users_level;
DROP INDEX IF EXISTS idx_users_status;
DROP INDEX IF EXISTS idx_users_created_at;
DROP INDEX IF EXISTS idx_users_deleted_at;
DROP INDEX IF EXISTS idx_users_email_status;

-- drop table
DROP TABLE IF EXISTS users;

-- drop extension
DROP EXTENSION IF EXISTS "pgcrypto";
