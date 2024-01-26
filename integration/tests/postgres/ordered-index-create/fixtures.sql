create table users (
  id integer primary key not null,
  email varchar(255) not null,
  "tz_1" timestamp,
  "tz_2" timestamp with time zone
);
