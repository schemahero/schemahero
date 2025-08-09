-- Data Migration: test-data-migration (Operation 1);
UPDATE "test_users" SET "region" = 'us-east-1', "status" = 'active' WHERE status IS NULL;
-- Data Migration: test-data-migration (Operation 2);
UPDATE "test_users" SET "full_name" = CONCAT(first_name, ' ', last_name), "display_email" = LOWER(email) WHERE full_name IS NULL OR display_email IS NULL;
