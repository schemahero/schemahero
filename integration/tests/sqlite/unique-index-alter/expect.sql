drop index if exists idx_users_email;
create unique index idx_users_email on users (email);
