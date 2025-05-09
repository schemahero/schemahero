CREATE SCHEMA test;

CREATE TABLE public.users (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL,
  email TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE test.users (
  id SERIAL PRIMARY KEY,
  username TEXT NOT NULL,
  email TEXT NOT NULL,
  active BOOLEAN DEFAULT TRUE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
