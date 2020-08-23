create table users (
  id integer primary key not null,
  email varchar(255) not null
);

create table projects (
  id integer primary key not null
) default charset=macroman;
