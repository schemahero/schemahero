create table users (
  id integer primary key not null,
  email varchar(255) not null,
  "tz_1" timestamp,
  "tz_2" timestamp,
  "tz_3" timestamp
);

create index users_tz_priority_1 on users (tz_1 desc, tz_2 asc);
