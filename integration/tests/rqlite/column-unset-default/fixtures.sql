create table users (id integer primary key not null, email text not null, account_type text not null default 'trial', num_seats integer not null default 5);
