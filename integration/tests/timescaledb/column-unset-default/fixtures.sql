create table users (
  id integer primary key not null,
  email varchar(255) not null,
  account_type varchar(10) not null default 'trial',
  num_seats integer not null default 5
);
