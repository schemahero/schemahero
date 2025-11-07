create table users (
  id integer primary key not null,
  login text,
  name text not null
);

create table other (
  id integer primary key not null,
  something text not null
);
