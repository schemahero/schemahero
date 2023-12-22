create table users (
  id integer primary key not null,
  email varchar(255) not null
) CHARSET=latin1;

create table projects (
  id integer primary key not null,
  data varchar(255) not null
) CHARSET=latin1;
