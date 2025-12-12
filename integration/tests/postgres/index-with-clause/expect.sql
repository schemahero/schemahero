create table "users" ("id" integer, "email" character varying (255) not null, "phone" character varying (10) not null default '', primary key ("id"), constraint "idx_users_email" unique ("email"));
create index idx_users_phone on users (phone) with (fillfactor = 80, gin_pending_list_limit = 64);
