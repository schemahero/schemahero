CREATE SCHEMA test;

CREATE TABLE test.users (
  id SERIAL PRIMARY KEY,
  username TEXT NOT NULL,
  email TEXT NOT NULL,
  active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE FUNCTION test.get_user_count_with_parameter(table_name text) RETURNS bigint AS $$
DECLARE
    user_count bigint;
BEGIN
    SELECT COUNT(*) INTO user_count FROM table_name;
    RETURN user_count;
END;
$$ language PLpgSQL;
