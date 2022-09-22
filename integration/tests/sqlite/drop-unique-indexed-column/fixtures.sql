create table users (id integer primary key not null, email text not null);
create unique index idx_users_email on users (email);
