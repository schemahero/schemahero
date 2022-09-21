create table users (id integer primary key not null, email text not null);
create index idx_users_email on users (email);
