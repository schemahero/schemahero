create table users (
  id integer primary key not null,
  email varchar(255) not null,
  account_type varchar(10),
  num_seats integer
);
