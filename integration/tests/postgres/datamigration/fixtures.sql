-- Initial test data for data migration integration test
CREATE TABLE test_users (
    id SERIAL PRIMARY KEY,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    email VARCHAR(100),
    status VARCHAR(20),
    region VARCHAR(50),
    full_name VARCHAR(101),
    display_email VARCHAR(100)
);

-- Insert test data with some NULL values and mixed case emails
INSERT INTO test_users (first_name, last_name, email, status, region) VALUES
    ('John', 'Doe', 'john.doe@example.com', 'active', 'us-east-1'),
    ('Jane', 'Smith', 'JANE.SMITH@EXAMPLE.COM', NULL, NULL),
    ('Bob', 'Johnson', 'Bob.Johnson@Example.com', NULL, NULL);
