alter table "events" alter column "created_at" type timestamp with time zone;
update "events" set "created_at" = "created_at" at time zone 'UTC';