create table users (id integer primary key not null, email text not null, phone text not null default '');
create index custom_idx_users_email on users (email);
