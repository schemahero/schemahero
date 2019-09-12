create table users (
  id integer primary key not null,
  email varchar(255) not null,
  phone varchar(10) not null default ''
);
