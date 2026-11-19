create table users (
  id integer primary key not null,
  email varchar(255) not null,
  created_at datetime not null default CURRENT_TIMESTAMP,
  updated_at datetime not null default CURRENT_TIMESTAMP
);
