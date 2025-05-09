-- Enable the pgvector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create a table with a vector column 
-- This will be used to verify that SchemaHero doesn't 
-- continuously suggest altering the vector type
CREATE TABLE embeddings (
  id INTEGER PRIMARY KEY NOT NULL,
  embedding vector NOT NULL
);

-- Insert a sample record to verify functionality
INSERT INTO embeddings (id, embedding) VALUES (1, '[1,2,3]');
