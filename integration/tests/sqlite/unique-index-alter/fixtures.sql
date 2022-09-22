create table users (id integer primary key not null, email text not null, phone text not null default '');
create unique index idx_users_email on users (phone);
