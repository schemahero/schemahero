create table "users" ("id" integer not null, "email" text not null, "name" text not null default 'salah', "ts_1" integer, "ts_2" real, "ts_3" text, primary key ("id"));
create index idx_users_email on users (email);
